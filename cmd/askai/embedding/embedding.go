package embedding

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/embedding"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func New(cfg *config.Config) *cobra.Command {
	var inputStrings []string
	var model string

	cmd := cobra.Command{
		Use:   "embedding",
		Short: `Ask AI to generate an embedding`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := cfg.EndpointConfig()
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			client := endpoint.NewClient()
			ctx := context.Background()

			req := openai.EmbeddingRequest{
				Model: openai.EmbeddingModel(model),
			}

			switch {
			case len(inputStrings) == 0:
				return errors.New("at least one input is required")
			case len(inputStrings) == 1:
				req.Input = inputStrings[0]
			case len(inputStrings) > 1:
				req.Input = inputStrings
			}

			err = embedding.Send(
				ctx,
				client,
				req,
				os.Stdout)
			if err != nil {
				return fmt.Errorf("embedding: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(
		&inputStrings,
		"input",
		[]string{},
		"one or more input strings")
	cmd.Flags().StringVar(
		&model,
		"model",
		"text-embedding-ada-002",
		"embedding model to use")

	return &cmd
}
