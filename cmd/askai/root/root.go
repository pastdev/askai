package root

import (
	"github.com/pastdev/askai/cmd/askai/complete"
	cmdcfg "github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/cmd/askai/embedding"
	"github.com/pastdev/askai/cmd/askai/image"
	"github.com/pastdev/askai/cmd/askai/models"
	"github.com/pastdev/askai/cmd/askai/tokens"
	"github.com/pastdev/askai/cmd/askai/version"
	"github.com/pastdev/askai/pkg/log"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var logLevel string
	var logFormat string

	cmd := cobra.Command{
		Use:   "askai",
		Short: `A tool for asking things of an AI.`,
		//nolint: revive // required to match upstream signature
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(logLevel)
			log.SetFormat(logFormat)
		},
	}

	cfg := cmdcfg.AddConfig(&cmd)

	cmd.PersistentFlags().StringVar(&logLevel, "log", "info", "log level")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "pretty", "log format (pretty|json)")

	cmd.AddCommand(complete.New(cfg))
	cmd.AddCommand(embedding.New(cfg))
	cmd.AddCommand(image.New(cfg))
	cmd.AddCommand(models.New(cfg))
	cmd.AddCommand(tokens.New())
	cmd.AddCommand(version.New())

	return &cmd
}
