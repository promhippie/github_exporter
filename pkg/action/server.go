package action

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/google/go-github/v52/github"
	"github.com/oklog/run"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/exporter-toolkit/web"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/promhippie/github_exporter/pkg/exporter"
	"github.com/promhippie/github_exporter/pkg/middleware"
	"github.com/promhippie/github_exporter/pkg/version"
	"golang.org/x/oauth2"
)

// Server handles the server sub-command.
func Server(cfg *config.Config, logger log.Logger) error {
	level.Info(logger).Log(
		"msg", "Launching GitHub Exporter",
		"version", version.String,
		"revision", version.Revision,
		"date", version.Date,
		"go", version.Go,
	)

	client := github.NewClient(
		oauth2.NewClient(
			context.Background(),
			oauth2.StaticTokenSource(
				&oauth2.Token{
					AccessToken: cfg.Target.Token,
				},
			),
		),
	)

	if cfg.Target.PrivateKey != "" && cfg.Target.AppID != 0 && cfg.Target.InstallID != 0 {
		privateKey, err := contentOrDecode(cfg.Target.PrivateKey)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to read GitHub key",
				"err", err,
			)

			return err
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

			return err
		}

		client = github.NewClient(
			&http.Client{
				Transport: transport,
			},
		)
	}

	if cfg.Target.BaseURL != "" {
		var (
			err error
		)

		client, err = github.NewEnterpriseClient(
			cfg.Target.BaseURL,
			cfg.Target.BaseURL,
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
		)

		if err != nil {
			level.Error(logger).Log(
				"msg", "Failed to parse base URL",
				"err", err,
			)

			return err
		}
	}

	var gr run.Group

	{
		server := &http.Server{
			Addr:         cfg.Server.Addr,
			Handler:      handler(cfg, logger, client),
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

func handler(cfg *config.Config, logger log.Logger, client *github.Client) *chi.Mux {
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
		root.Get(cfg.Server.Path, func(w http.ResponseWriter, r *http.Request) {
			reg.ServeHTTP(w, r)
		})

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
