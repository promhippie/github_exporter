//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/promhippie/github_exporter/pkg/command"
	"github.com/promhippie/github_exporter/pkg/config"
	"github.com/urfave/cli/v2"
)

type flag struct {
	Flag    string
	Default string
	Envs    []string
	Help    string
	List    bool
}

func main() {
	flags := make([]flag, 0)

	for _, f := range command.RootFlags(config.Load()) {
		switch v := f.(type) {
		case *cli.StringFlag:
			flags = append(flags, flag{
				Flag:    v.Name,
				Default: v.Value,
				Envs:    v.EnvVars,
				Help:    v.Usage,
				List:    false,
			})
		case *cli.BoolFlag:
			flags = append(flags, flag{
				Flag:    v.Name,
				Default: fmt.Sprintf("%+v", v.Value),
				Envs:    v.EnvVars,
				Help:    v.Usage,
				List:    false,
			})
		case *cli.DurationFlag:
			flags = append(flags, flag{
				Flag:    v.Name,
				Default: v.Value.String(),
				Envs:    v.EnvVars,
				Help:    v.Usage,
				List:    false,
			})
		case *cli.StringSliceFlag:
			flags = append(flags, flag{
				Flag:    v.Name,
				Default: strings.Join(v.Value.Value(), ", "),
				Envs:    v.EnvVars,
				Help:    v.Usage,
				List:    true,
			})
		default:
			fmt.Printf("unknown type: %s\n", v)
			os.Exit(1)
		}
	}

	f, err := os.Create("docs/partials/envvars.md")

	if err != nil {
		fmt.Printf("failed to create file")
		os.Exit(1)
	}

	defer f.Close()

	last := flags[len(flags)-1]
	for _, row := range flags {
		f.WriteString(
			strings.Join(
				row.Envs,
				", ",
			) + "\n",
		)

		f.WriteString(fmt.Sprintf(
			": %s",
			row.Help,
		))

		if row.List {
			f.WriteString(
				", comma-separated list",
			)
		}

		if row.Default != "" {
			f.WriteString(fmt.Sprintf(
				", defaults to `%s`",
				row.Default,
			))
		}

		f.WriteString("\n")

		if row.Flag != last.Flag {
			f.WriteString("\n")
		}
	}
}
