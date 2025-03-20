package complete

import (
	"context"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/chatcompletion"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func New(cfg *config.Config) *cobra.Command {
	var req openai.ChatCompletionRequest
	var conversation string

	cmd := cobra.Command{
		Use:   "complete",
		Short: `Ask AI to complete a chat`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			endpoint, err := cfg.EndpointConfig()
			if err != nil {
				return fmt.Errorf("new client: %w", err)
			}

			client := endpoint.NewClient()
			ctx := context.Background()

			defaults := openai.ChatCompletionRequest{}
			if endpoint.ChatCompletionDefaults != nil {
				defaults = *endpoint.ChatCompletionDefaults
			}

			if conversation == "" {
				err := mergo.Merge(&req, defaults)
				if err != nil {
					return fmt.Errorf("apply defaults: %w", err)
				}
				req.Messages = append(defaults.Messages, req.Messages...)

				err = chatcompletion.Send(ctx, client, req, os.Stdout)
				if err != nil {
					return fmt.Errorf("complete chat: %w", err)
				}
			} else {
				conv, err := chatcompletion.LoadPersistentConversation(conversation, defaults)
				if err != nil {
					return fmt.Errorf("load %s: %w", conversation, err)
				}

				err = chatcompletion.SendReply(
					ctx,
					client,
					&conv,
					req,
					os.Stdout)
				if err != nil {
					return fmt.Errorf("complete chat: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(
		&conversation,
		"conversation",
		"",
		"a named conversation to start or continue")
	cmd.Flags().IntVar(
		&req.MaxTokens,
		"max-tokens",
		0,
		"The maximum number of tokens that can be generated in the chat completion (deprecated in favor of max-completion-tokens, but older servers may still only support this)")
	cmd.Flags().IntVar(
		&req.MaxCompletionTokens,
		"max-completion-tokens",
		0,
		"The maximum number of tokens that can be generated in the chat completion")
	cmd.Flags().StringVar(
		&req.Model,
		"model",
		"",
		"ai model to use")
	MessageArrayVarP(
		cmd.Flags(),
		"",
		&req.Messages,
		"message",
		"m",
		nil,
		"one or more complete json messages")
	MessageArrayVarP(
		cmd.Flags(),
		"user",
		&req.Messages,
		"user",
		"u",
		nil,
		"one or more user content messages")
	MessageArrayVarP(
		cmd.Flags(),
		"system",
		&req.Messages,
		"system",
		"s",
		nil,
		"one or more system content messages")
	MessageArrayVarP(
		cmd.Flags(),
		"assistant",
		&req.Messages,
		"assistant",
		"a",
		nil,
		"one or more assistant content messages")
	cmd.Flags().BoolVar(
		&req.Stream,
		"stream",
		false,
		"stream the response")
	cmd.Flags().Float32VarP(
		&req.Temperature,
		"temperature",
		"t",
		0,
		"temperature, zero is not set, so if you want zero, use 0.0000001 or similar")

	return &cmd
}
