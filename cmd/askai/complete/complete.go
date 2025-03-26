package complete

import (
	"context"
	"encoding/json"
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
	var logItBias string

	cmd := cobra.Command{
		Use:   "complete",
		Short: `Ask AI to complete a chat`,
		Example: `  # ask a simple question
  askai complete --user "what color is the sky?"

  # engage in a conversation about life, the world, and everything
  askai complete \
    --conversation life_the_world_and_everything \
    --user "what is the meaning of life, the world, and everything?"

  # get your answer from a frat boy
  askai complete \
    --system "you are a frat boy during peak frat" \
    --user "whats the best beer for a night at the sorority?"

  # create a short story about foo without using "foo"
  askai complete \
    --logit-bias "$(
      printf "{%s}" \
      "$(
        for i in "foo" " foo" "Foo" " Foo"; do
        printf '"%s":-100,' \
          "$(askai tokens encode "$i" | clconf --pipe getv /0)"
      done \
        | sed 's/,$//')")" \
    --user "tell me a short story about foo"`,
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

			if logItBias != "" {
				err := json.Unmarshal([]byte(logItBias), &req.LogitBias)
				if err != nil {
					return fmt.Errorf("logit bias: %w", err)
				}
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
		"A named conversation to start or continue")
	// may want to mention in the help somewher that this value could be
	// generated with the following approach:
	//  printf "{%s}" "$(for i in "foo" " foo" "Foo" " Foo"; do printf '"%s":-100,' "$(askai tokens encode "$i" | clconf --pipe getv /0)"; done | sed 's/,$//')"
	cmd.Flags().StringVar(
		&logItBias,
		"logit-bias",
		"",
		"A json map of string to int where they key is the token (can be obtained using: `askai tokens encode`) and the value is a bias between -100 (prohibit) and 100 (encourage)")
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
		"AI model to use")
	MessageArrayVarP(
		cmd.Flags(),
		"",
		&req.Messages,
		"message",
		"m",
		nil,
		"One or more complete json messages")
	MessageArrayVarP(
		cmd.Flags(),
		"user",
		&req.Messages,
		"user",
		"u",
		nil,
		"One or more user content messages")
	MessageArrayVarP(
		cmd.Flags(),
		"system",
		&req.Messages,
		"system",
		"s",
		nil,
		"One or more system content messages")
	MessageArrayVarP(
		cmd.Flags(),
		"assistant",
		&req.Messages,
		"assistant",
		"a",
		nil,
		"One or more assistant content messages")
	cmd.Flags().BoolVar(
		&req.Stream,
		"stream",
		false,
		"Stream the response")
	cmd.Flags().Float32VarP(
		&req.Temperature,
		"temperature",
		"t",
		0,
		"Temperature, zero is not set, so if you want zero, use 0.0000001 or similar")

	return &cmd
}
