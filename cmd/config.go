package cmd

import (
	"fmt"
	"time"

	"github.com/urfave/cli/v2"

	"github.com/lz-wang/pvectl/internal/config"
	"github.com/lz-wang/pvectl/internal/output"
)

func newConfigCommand() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Manage pvectl config",
		Subcommands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Initialize a default HomeLab context",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "name", Value: "home", Usage: "context name"},
					&cli.StringFlag{Name: "endpoint", Usage: "PVE API endpoint, for example https://pve.lan:8006/api2/json", Required: true},
					&cli.StringFlag{Name: "token-id", Usage: "PVE API token id, for example automation@pve!pvectl", Required: true},
					&cli.StringFlag{Name: "token-secret-env", Usage: "environment variable containing the PVE API token secret", Required: true},
					&cli.BoolFlag{Name: "insecure", Usage: "skip TLS certificate verification for this context"},
					&cli.StringFlag{Name: "timeout", Value: "30s", Usage: "PVE API request timeout for this context"},
					&cli.StringFlag{Name: "default-output", Value: output.FormatTable, Usage: "default output format: table,json,yaml"},
					&cli.BoolFlag{Name: "overwrite", Usage: "overwrite an existing context"},
					&cli.BoolFlag{Name: "no-use", Usage: "do not set the initialized context as current"},
				},
				Action: func(c *cli.Context) error {
					if err := requireNoExtraArgs(c, 0); err != nil {
						return err
					}
					defaultOutput := output.NormalizeFormat(c.String("default-output"))
					if err := output.ValidateFormat(defaultOutput); err != nil {
						return err
					}
					if _, err := time.ParseDuration(c.String("timeout")); err != nil {
						return fmt.Errorf("invalid timeout %q: %w", c.String("timeout"), err)
					}

					cfg, err := config.LoadOrEmpty(c.String("config"))
					if err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					if err := cfg.InitContext(config.InitOptions{
						Name: c.String("name"),
						Context: config.Context{
							Endpoint:           c.String("endpoint"),
							TokenID:            c.String("token-id"),
							TokenSecretEnv:     c.String("token-secret-env"),
							InsecureSkipVerify: c.Bool("insecure"),
							Timeout:            c.String("timeout"),
							DefaultOutput:      defaultOutput,
						},
						Overwrite: c.Bool("overwrite"),
						Use:       !c.Bool("no-use"),
					}); err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					return config.Save(c.String("config"), cfg)
				},
			},
			{
				Name:      "set-context",
				Usage:     "Create or update a context",
				ArgsUsage: "NAME",
				Flags: []cli.Flag{
					&cli.StringFlag{Name: "endpoint", Usage: "PVE API endpoint, for example https://pve.lan:8006/api2/json", Required: true},
					&cli.StringFlag{Name: "token-id", Usage: "PVE API token id, for example automation@pve!pvectl", Required: true},
					&cli.StringFlag{Name: "token-secret-env", Usage: "environment variable containing the PVE API token secret", Required: true},
					&cli.BoolFlag{Name: "insecure", Usage: "skip TLS certificate verification for this context"},
					&cli.StringFlag{Name: "timeout", Usage: "PVE API request timeout for this context"},
					&cli.StringFlag{Name: "default-output", Usage: "default output format: table,json,yaml"},
				},
				Action: func(c *cli.Context) error {
					if err := requireNoExtraArgs(c, 1); err != nil {
						return err
					}
					defaultOutput := c.String("default-output")
					if defaultOutput != "" {
						if err := output.ValidateFormat(defaultOutput); err != nil {
							return err
						}
					}

					cfg, err := config.LoadOrEmpty(c.String("config"))
					if err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					if err := cfg.SetContext(c.Args().First(), config.Context{
						Endpoint:           c.String("endpoint"),
						TokenID:            c.String("token-id"),
						TokenSecretEnv:     c.String("token-secret-env"),
						InsecureSkipVerify: c.Bool("insecure"),
						Timeout:            c.String("timeout"),
						DefaultOutput:      defaultOutput,
					}); err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					return config.Save(c.String("config"), cfg)
				},
			},
			{
				Name:      "use-context",
				Usage:     "Set the current context",
				ArgsUsage: "NAME",
				Action: func(c *cli.Context) error {
					if err := requireNoExtraArgs(c, 1); err != nil {
						return err
					}
					cfg, err := config.Load(c.String("config"))
					if err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					if err := cfg.UseContext(c.Args().First()); err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					return config.Save(c.String("config"), cfg)
				},
			},
			{
				Name:  "current-context",
				Usage: "Print the current context",
				Action: func(c *cli.Context) error {
					if err := requireNoExtraArgs(c, 0); err != nil {
						return err
					}
					cfg, err := config.Load(c.String("config"))
					if err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					if cfg.CurrentContext == "" {
						return fmt.Errorf("config error: current_context is empty")
					}
					_, err = fmt.Fprintln(c.App.Writer, cfg.CurrentContext)
					return err
				},
			},
			{
				Name:  "view",
				Usage: "Print the config file",
				Action: func(c *cli.Context) error {
					if err := requireNoExtraArgs(c, 0); err != nil {
						return err
					}
					cfg, err := config.Load(c.String("config"))
					if err != nil {
						return fmt.Errorf("config error: %w", err)
					}
					data, err := config.ToYAML(cfg)
					if err != nil {
						return err
					}
					_, err = c.App.Writer.Write(data)
					return err
				},
			},
		},
	}
}
