package complete

import (
	"context"
	"fmt"
	"os"

	"github.com/pastdev/askai/pkg/askai"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func New(clientCfg *openai.ClientConfig) *cobra.Command {
	var req openai.ChatCompletionRequest
	var conversation string

	cmd := cobra.Command{
		Use:   "complete",
		Short: `Ask CI to complete a chat`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := openai.NewClientWithConfig(*clientCfg)

			ctx := context.Background()

			if conversation == "" {
				err := askai.Send(ctx, client, req, os.Stdout)
				if err != nil {
					return fmt.Errorf("complete chat: %w", err)
				}
			} else {
				conv, isNew, err := askai.LoadPersistentConversation(conversation)
				if err != nil {
					return fmt.Errorf("load %s: %w", conversation, err)
				}
				if isNew {
					conv.SetModel(req.Model)
				}

				err = askai.SendReply(ctx, client, &conv, req.Messages, req.Stream, os.Stdout)
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
	cmd.Flags().StringVar(
		&req.Model,
		"model",
		"mistral",
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
