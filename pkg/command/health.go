package command

import (
	"fmt"
	"net/http"

	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/urfave/cli/v3"
)

// Health provides the sub-command to perform a health check.
func Health(cfg *config.Config) *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "Perform health checks",
		Flags: HealthFlags(cfg),
		Action: func(_ *cli.Context) error {
			logger := setupLogger(cfg)

			resp, err := http.Get(
				fmt.Sprintf(
					"http://%s/healthz",
					cfg.Server.Addr,
				),
			)

			if err != nil {
				logger.Error("Failed to request health check",
					"err", err,
				)

				return err
			}

			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				logger.Error("Health check seems to be in bad state",
					"err", err,
					"code", resp.StatusCode,
				)

				return err
			}

			logger.Debug("Health check seems to be fine",
				"code", resp.StatusCode,
			)

			return nil
		},
	}
}

// HealthFlags defines the available health flags.
func HealthFlags(cfg *config.Config) []cli.Flag {
	return []cli.Flag{
		&cli.StringFlag{
			Name:        "web.address",
			Value:       "0.0.0.0:9504",
			Usage:       "Address to bind the metrics server",
			EnvVars:     []string{"GITHUB_EXPORTER_WEB_ADDRESS"},
			Destination: &cfg.Server.Addr,
		},
	}
}
