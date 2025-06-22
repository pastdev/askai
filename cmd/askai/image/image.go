package image

import (
	"context"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/image"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func New(cfg *config.Config) *cobra.Command {
	var req openai.ImageRequest
	var open bool
	var output string
	var outputFileDir string

	cmd := cobra.Command{
		Use:   "image",
		Short: `Ask AI to generate an image`,
		Example: `  # generate a dog picture
  askai image "a picture of a dog"`,
		Args: cobra.ExactArgs(1),
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			prompt := args[0]

			endpoint, err := cfg.EndpointConfig()
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			client := endpoint.NewClient()
			ctx := context.Background()

			defaults := openai.ImageRequest{}
			if endpoint.ImageDefaults != nil {
				defaults = *endpoint.ImageDefaults
			}

			req.Prompt = prompt

			var writer image.ResponseWriter
			switch output {
			case "raw":
				writer = &image.RawResponseWriter{
					Open: open,
					W:    os.Stdout,
				}
			case "file":
				if outputFileDir == "" {
					outputFileDir, err = os.MkdirTemp(os.TempDir(), "askai-*")
					if err != nil {
						return fmt.Errorf("new create temp dir: %w", err)
					}
				}
				writer = &image.FileResponseWriter{
					Dir:  outputFileDir,
					Open: open,
				}
			}

			err = mergo.Merge(&req, defaults)
			if err != nil {
				return fmt.Errorf("new apply defaults: %w", err)
			}

			err = image.Send(ctx, client, req, writer)
			if err != nil {
				return fmt.Errorf("new send: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(
		&req.N,
		"count",
		1,
		"Number of images to generate")
	cmd.Flags().StringVar(
		&req.Model,
		"model",
		"",
		"AI model to use")
	cmd.Flags().BoolVar(
		&open,
		"open",
		false,
		"Open the generated image(s) with the default app for the generated image type")
	cmd.Flags().StringVar(
		&output,
		"output",
		"raw",
		"Format of output, one of: raw, file")
	cmd.Flags().StringVar(
		&outputFileDir,
		"output-file-dir",
		"",
		"The directory to write files to when using --output file (defaults to a temp dir)")
	cmd.Flags().StringVar(
		&req.ResponseFormat,
		"response-format",
		"url",
		"The response format, one of: url b64_json")

	return &cmd
}
