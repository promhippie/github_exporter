package action

import (
	"context"
	"crypto/tls"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-chi/chi/v5"
	"github.com/google/go-github/v68/github"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/exporter"
	"github.com/promhippie/github_exporter/pkg/middleware"
	"github.com/promhippie/github_exporter/pkg/store"
	"github.com/promhippie/github_exporter/pkg/version"
	"golang.org/x/oauth2"
)

// Server handles the server sub-command.
func Server(cfg *config.Config, db store.Store, logger *slog.Logger) error {
	logger.Info("Launching GitHub Exporter",
		"version", version.String,
		"revision", version.Revision,
		"date", version.Date,
		"go", version.Go,
	)

	client, err := getClient(cfg, logger)

	if err != nil {
		return err
	}

	var gr run.Group

	{
		server := &http.Server{
			Addr:         cfg.Server.Addr,
			Handler:      handler(cfg, db, logger, client),
			ReadTimeout:  5 * time.Second,
			WriteTimeout: cfg.Server.Timeout,
		}

		gr.Add(func() error {
			logger.Info("Starting metrics server",
				"address", cfg.Server.Addr,
			)

			return web.ListenAndServe(
				server,
				&web.FlagConfig{
					WebListenAddresses: sliceP([]string{cfg.Server.Addr}),
					WebSystemdSocket:   boolP(false),
					WebConfigFile:      stringP(cfg.Server.Web),
				},
				logger,
			)
		}, func(reason error) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			if err := server.Shutdown(ctx); err != nil {
				logger.Error("Failed to shutdown metrics gracefully",
					"err", err,
				)

				return
			}

			logger.Info("Metrics shutdown gracefully",
				"reason", reason,
			)
		})
	}

	{
		stop := make(chan os.Signal, 1)

		gr.Add(func() error {
			signal.Notify(stop, os.Interrupt)

			<-stop

			return nil
		}, func(_ error) {
			close(stop)
		})
	}

	return gr.Run()
}

func handler(cfg *config.Config, db store.Store, logger *slog.Logger, client *github.Client) *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer(logger))
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Timeout)
	mux.Use(middleware.Cache)

	if cfg.Server.Pprof {
		mux.Mount("/debug", middleware.Profiler())
	}

	if cfg.Collector.Admin {
		logger.Debug("Admin collector registered")

		registry.MustRegister(exporter.NewAdminCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.Orgs {
		logger.Debug("Org collector registered")

		registry.MustRegister(exporter.NewOrgCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.Repos {
		logger.Debug("Repo collector registered")

		registry.MustRegister(exporter.NewRepoCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.Billing {
		logger.Debug("Billing collector registered")

		registry.MustRegister(exporter.NewBillingCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.Runners {
		logger.Debug("Runner collector registered")

		registry.MustRegister(exporter.NewRunnerCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.WorkflowRuns {
		logger.Debug("WorkflowRun collector registered")

		registry.MustRegister(exporter.NewWorkflowRunCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.WorkflowJobs {
		logger.Debug("WorkflowJob collector registered")

		registry.MustRegister(exporter.NewWorkflowJobCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	reg := promhttp.HandlerFor(
		registry,
		promhttp.HandlerOpts{
			ErrorLog: promLogger{logger},
		},
	)

	mux.NotFound(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cfg.Server.Path, http.StatusMovedPermanently)
	})

	mux.Route("/", func(root chi.Router) {
		root.Handle(cfg.Server.Path, reg)

		if cfg.Collector.WorkflowRuns || cfg.Collector.WorkflowJobs {
			root.HandleFunc(cfg.Webhook.Path, func(w http.ResponseWriter, r *http.Request) {
				secret, err := config.Value(cfg.Webhook.Secret)

				if err != nil {
					logger.Error("Failed to read webhook secret",
						"error", err,
					)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)

					io.WriteString(w, http.StatusText(http.StatusInternalServerError))
					return
				}

				payload, err := github.ValidatePayload(
					r,
					[]byte(secret),
				)

				if err != nil {
					logger.Error("Failed to parse github webhook",
						"error", err,
					)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)

					io.WriteString(w, http.StatusText(http.StatusInternalServerError))
					return
				}

				event, err := github.ParseWebHook(
					github.WebHookType(r),
					payload,
				)

				if err != nil {
					logger.Error("Failed to parse github event",
						"error", err,
					)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)

					io.WriteString(w, http.StatusText(http.StatusInternalServerError))
					return
				}

				switch event := event.(type) {
				case *github.WorkflowRunEvent:
					wfRun := event.GetWorkflowRun()
					logger.Debug("Received webhook request",
						"type", "workflow_run",
						"owner", event.GetRepo().GetOwner().GetLogin(),
						"repo", event.GetRepo().GetName(),
						"workflow", wfRun.GetWorkflowID(),
						"number", wfRun.GetRunNumber(),
						"id", wfRun.GetID(),
						"event", wfRun.GetEvent(),
						"status", wfRun.GetStatus(),
						"conclusion", wfRun.GetConclusion(),
						"actor", wfRun.GetActor().GetLogin(),
						"created_at", wfRun.GetCreatedAt().Time.Unix(),
						"updated_at", wfRun.GetUpdatedAt().Time.Unix(),
						"started_at", wfRun.GetRunStartedAt().Time.Unix(),
					)

					if err := db.StoreWorkflowRunEvent(event); err != nil {
						logger.Error("Failed to store github event",
							"type", "workflow_run",
							"owner", event.GetRepo().GetOwner().GetLogin(),
							"repo", event.GetRepo().GetName(),
							"workflow", wfRun.GetWorkflowID(),
							"number", wfRun.GetRunNumber(),
							"error", err,
						)

						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)

						io.WriteString(w, http.StatusText(http.StatusInternalServerError))
						return
					}
				case *github.WorkflowJobEvent:
					wfJob := event.GetWorkflowJob()
					logger.Debug("received webhook request",
						"type", "workflow_job",
						"owner", event.GetRepo().GetOwner().GetLogin(),
						"repo", event.GetRepo().GetName(),
						"id", wfJob.GetID(),
						"name", wfJob.GetName(),
						"attempt", wfJob.GetRunAttempt(),
						"status", wfJob.GetStatus(),
						"conclusion", wfJob.GetConclusion(),
						"created_at", wfJob.GetCreatedAt().Time.Unix(),
						"started_at", wfJob.GetStartedAt().Time.Unix(),
						"completed_at", wfJob.GetCompletedAt().Time.Unix(),
						"labels", strings.Join(wfJob.Labels, ", "),
					)

					if err := db.StoreWorkflowJobEvent(event); err != nil {
						logger.Error(
							"failed to store github event",
							"type", "workflow_job",
							"owner", event.GetRepo().GetOwner().GetLogin(),
							"repo", event.GetRepo().GetName(),
							"name", wfJob.GetName(),
							"id", wfJob.GetID(),
							"error", err,
						)

						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)

						io.WriteString(w, http.StatusText(http.StatusInternalServerError))
						return
					}
				}

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)

				io.WriteString(w, http.StatusText(http.StatusOK))
			})
		}

		root.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, http.StatusText(http.StatusOK))
		})

		root.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, http.StatusText(http.StatusOK))
		})
	})

	return mux
}

func useEnterprise(cfg *config.Config, _ *slog.Logger) bool {
	return cfg.Target.BaseURL != ""
}

func useApplication(cfg *config.Config, _ *slog.Logger) bool {
	return cfg.Target.PrivateKey != "" && cfg.Target.AppID != 0 && cfg.Target.InstallID != 0
}

func getClient(cfg *config.Config, logger *slog.Logger) (*github.Client, error) {
	if useEnterprise(cfg, logger) {
		return getEnterprise(cfg, logger)
	}

	if useApplication(cfg, logger) {
		privateKey, err := config.Value(cfg.Target.PrivateKey)

		if err != nil {
			logger.Error("Failed to read GitHub key",
				"err", err,
			)

			return nil, err
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			cfg.Target.AppID,
			cfg.Target.InstallID,
			[]byte(privateKey),
		)

		if err != nil {
			logger.Error("Failed to create GitHub transport",
				"err", err,
			)

			return nil, err
		}

		return github.NewClient(
			&http.Client{
				Transport: transport,
			},
		), nil
	}

	accessToken, err := config.Value(cfg.Target.Token)

	if err != nil {
		logger.Error("Failed to read token",
			"err", err,
		)

		return nil, err
	}

	return github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: accessToken,
				},
			),
		),
	), nil
}

func getEnterprise(cfg *config.Config, logger *slog.Logger) (*github.Client, error) {
	if useApplication(cfg, logger) {
		privateKey, err := config.Value(cfg.Target.PrivateKey)

		if err != nil {
			logger.Error("Failed to read GitHub key",
				"err", err,
			)

			return nil, err
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			cfg.Target.AppID,
			cfg.Target.InstallID,
			[]byte(privateKey),
		)

		if err != nil {
			logger.Error("Failed to create GitHub transport",
				"err", err,
			)

			return nil, err
		}

		if !strings.HasSuffix(cfg.Target.BaseURL, "/api/v3") &&
			!strings.HasSuffix(cfg.Target.BaseURL, "/api/v3/") {
			transport.BaseURL = cfg.Target.BaseURL + "/api/v3"
		} else {
			transport.BaseURL = cfg.Target.BaseURL
		}

		client, err := github.NewClient(
			&http.Client{
				Transport: transport,
			},
		).WithEnterpriseURLs(
			cfg.Target.BaseURL,
			cfg.Target.BaseURL,
		)

		if err != nil {
			logger.Error("Failed to parse base URL",
				"err", err,
			)

			return nil, err
		}

		return client, err
	}

	accessToken, err := config.Value(cfg.Target.Token)

	if err != nil {
		logger.Error("Failed to read token",
			"err", err,
		)

		return nil, err
	}

	client, err := github.NewClient(
		oauth2.NewClient(
			context.WithValue(
				context.Background(),
				oauth2.HTTPClient,
				&http.Client{
					Transport: &http.Transport{
						TLSClientConfig: &tls.Config{
							InsecureSkipVerify: cfg.Target.Insecure,
						},
					},
				},
			),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: accessToken,
				},
			),
		),
	).WithEnterpriseURLs(
		cfg.Target.BaseURL,
		cfg.Target.BaseURL,
	)

	if err != nil {
		logger.Error("Failed to parse base URL",
			"err", err,
		)

		return nil, err
	}

	return client, err
}
