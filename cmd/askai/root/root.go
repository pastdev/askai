package root

import (
	"net/http"

	"github.com/pastdev/askai/cmd/askai/complete"
	"github.com/pastdev/askai/cmd/askai/version"
	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	var logLevel string
	var logFormat string

	openaiCfg := openai.ClientConfig{
		APIType: openai.APITypeOpenAI,
		// start ollama on windows, need to specify 0.0.0.0 or it will only
		// listen on 127.0.0.1
		//
		//   PS C:\Users\lucas> $env:OLLAMA_HOST="0.0.0.0:11434"
		//   PS C:\Users\lucas> Get-NetTCPConnection -LocalPort 11434
		//
		//   LocalAddress                        LocalPort RemoteAddress                       RemotePort State       AppliedSetting OwningProcess
		//   ------------                        --------- -------------                       ---------- -----       -------------- -------------
		//   127.0.0.1                           11434     0.0.0.0                             0          Listen                     10276
		//
		//
		//   PS C:\Users\lucas> ollama serve
		//
		// then from WSL need to use the nameserver to reach out to the windows
		// machine:
		//
		//   ltheisen@ltserver ~/git/pastdev/askai
		//   $ grep ^nameserver /etc/resolv.conf
		//   nameserver 172.22.144.1
		BaseURL:            "http://172.22.144.1:11434/v1",
		EmptyMessagesLimit: 300,
		HTTPClient:         &http.Client{},
		OrgID:              "",
	}

	cmd := cobra.Command{
		Use:   "askai",
		Short: `A tool for asking things of an AI.`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			log.SetLevel(logLevel)
			log.SetFormat(logFormat)
			log.Trace().Str("level", logLevel).Str("format", logFormat).Msg("log config")
			log.Debug().Str("level", logLevel).Str("format", logFormat).Msg("log config")
			log.Info().Str("level", logLevel).Str("format", logFormat).Msg("log config")
			log.Warn().Str("level", logLevel).Str("format", logFormat).Msg("log config")
			log.Error().Str("level", logLevel).Str("format", logFormat).Msg("log config")
		},
	}

	cmd.AddCommand(complete.New(&openaiCfg))
	cmd.AddCommand(version.New())

	cmd.PersistentFlags().StringVar(&logLevel, "log", "info", "log level")
	cmd.PersistentFlags().StringVar(&logFormat, "log-format", "pretty", "log format (pretty|json)")

	return &cmd
}
