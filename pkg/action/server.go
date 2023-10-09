package action

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v56/github"
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
func Server(cfg *config.Config, db store.Store, logger log.Logger) error {
	level.Info(logger).Log(
		"msg", "Launching GitHub Exporter",
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
			level.Info(logger).Log(
				"msg", "Starting metrics server",
				"addr", cfg.Server.Addr,
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
				level.Error(logger).Log(
					"msg", "Failed to shutdown metrics gracefully",
					"err", err,
				)

				return
			}

			level.Info(logger).Log(
				"msg", "Metrics shutdown gracefully",
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
		}, func(err error) {
			close(stop)
		})
	}

	return gr.Run()
}

func handler(cfg *config.Config, db store.Store, logger log.Logger, client *github.Client) *chi.Mux {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer(logger))
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Timeout)
	mux.Use(middleware.Cache)

	if cfg.Server.Pprof {
		mux.Mount("/debug", middleware.Profiler())
	}

	if cfg.Collector.Admin {
		level.Debug(logger).Log(
			"msg", "Admin collector registered",
		)

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
		level.Debug(logger).Log(
			"msg", "Org collector registered",
		)

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
		level.Debug(logger).Log(
			"msg", "Repo collector registered",
		)

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
		level.Debug(logger).Log(
			"msg", "Billing collector registered",
		)

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
		level.Debug(logger).Log(
			"msg", "Runner collector registered",
		)

		registry.MustRegister(exporter.NewRunnerCollector(
			logger,
			client,
			db,
			requestFailures,
			requestDuration,
			cfg.Target,
		))
	}

	if cfg.Collector.Workflows {
		level.Debug(logger).Log(
			"msg", "Workflow collector registered",
		)

		registry.MustRegister(exporter.NewWorkflowCollector(
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

		if cfg.Collector.Workflows {
			root.HandleFunc(cfg.Webhook.Path, func(w http.ResponseWriter, r *http.Request) {
				payload, err := github.ValidatePayload(
					r,
					[]byte(cfg.Webhook.Secret),
				)

				if err != nil {
					level.Error(logger).Log(
						"msg", "failed to parse github webhook",
						"error", err,
					)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)

					io.WriteString(w, http.StatusText(http.StatusInternalServerError))
				}

				event, err := github.ParseWebHook(
					github.WebHookType(r),
					payload,
				)

				if err != nil {
					level.Error(logger).Log(
						"msg", "failed to parse github event",
						"error", err,
					)

					w.Header().Set("Content-Type", "text/plain")
					w.WriteHeader(http.StatusInternalServerError)

					io.WriteString(w, http.StatusText(http.StatusInternalServerError))
				}

				switch event := event.(type) {
				case *github.WorkflowRunEvent:
					if err := db.StoreWorkflowRunEvent(event); err != nil {
						level.Error(logger).Log(
							"msg", "failed to store github event",
							"type", "workflow_run",
							"error", err,
						)

						w.Header().Set("Content-Type", "text/plain")
						w.WriteHeader(http.StatusInternalServerError)

						io.WriteString(w, http.StatusText(http.StatusInternalServerError))
					}
				}

				w.Header().Set("Content-Type", "text/plain")
				w.WriteHeader(http.StatusOK)

				io.WriteString(w, http.StatusText(http.StatusOK))
			})
		}

		root.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, http.StatusText(http.StatusOK))
		})

		root.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			io.WriteString(w, http.StatusText(http.StatusOK))
		})
	})

	return mux
}

func boolP(i bool) *bool {
	return &i
}

func stringP(i string) *string {
	return &i
}

func sliceP(i []string) *[]string {
	return &i
}

func contentOrDecode(file string) ([]byte, error) {
	decoded, err := base64.StdEncoding.DecodeString(
		file,
	)

	if err != nil {
		return os.ReadFile(file)
	}

	return decoded, nil
}

func useEnterprise(cfg *config.Config, _ log.Logger) bool {
	return cfg.Target.BaseURL != ""
}

func useApplication(cfg *config.Config, _ log.Logger) bool {
	return cfg.Target.PrivateKey != "" && cfg.Target.AppID != 0 && cfg.Target.InstallID != 0
}

func getClient(cfg *config.Config, logger log.Logger) (*github.Client, error) {
	if useEnterprise(cfg, logger) {
		return getEnterprise(cfg, logger)
	}

	if useApplication(cfg, logger) {
		privateKey, err := contentOrDecode(cfg.Target.PrivateKey)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to read GitHub key",
				"err", err,
			)

			return nil, err
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			cfg.Target.AppID,
			cfg.Target.InstallID,
			privateKey,
		)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to create GitHub transport",
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

	return github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: cfg.Target.Token,
				},
			),
		),
	), nil
}

func getEnterprise(cfg *config.Config, logger log.Logger) (*github.Client, error) {
	if useApplication(cfg, logger) {
		privateKey, err := contentOrDecode(cfg.Target.PrivateKey)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to read GitHub key",
				"err", err,
			)

			return nil, err
		}

		transport, err := ghinstallation.New(
			http.DefaultTransport,
			cfg.Target.AppID,
			cfg.Target.InstallID,
			privateKey,
		)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to create GitHub transport",
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
			level.Error(logger).Log(
				"msg", "Failed to parse base URL",
				"err", err,
			)

			return nil, err
		}

		return client, err
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
					AccessToken: cfg.Target.Token,
				},
			),
		),
	).WithEnterpriseURLs(
		cfg.Target.BaseURL,
		cfg.Target.BaseURL,
	)

	if err != nil {
		level.Error(logger).Log(
			"msg", "Failed to parse base URL",
			"err", err,
		)

		return nil, err
	}

	return client, err
}
