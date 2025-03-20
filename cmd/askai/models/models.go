package models

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/spf13/cobra"
)

func New(cfg *config.Config) *cobra.Command {
	var modelID string

	cmd := cobra.Command{
		Use:   "models",
		Short: `Lists available modelse for for an AI endpoint`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := cfg.EndpointConfig()
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			client := endpoint.NewClient()
			ctx := context.Background()

			var res any
			if modelID == "" {
				res, err = client.ListModels(ctx)
			} else {
				res, err = client.GetModel(ctx, url.QueryEscape(modelID))
			}
			if err != nil {
				return fmt.Errorf("obtain model info: %w", err)
			}

			out, err := json.Marshal(res)
			if err != nil {
				return fmt.Errorf("marshal model info: %w", err)
			}

			fmt.Printf("%s", out)

			return nil
		},
	}

	cmd.Flags().StringVar(
		&modelID,
		"id",
		"",
		"id of model to lookup")

	return &cmd
}
