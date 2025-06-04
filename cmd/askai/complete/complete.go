package complete

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"dario.cat/mergo"
	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/chatcompletion"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

func encodeAttachments(attachements []string) (string, error) {
	var data strings.Builder
	for _, attachment := range attachements {
		parts := strings.Split(attachment, ":")
		var filename string
		var path string
		switch len(parts) {
		case 1:
			path = parts[0]
			filename = filepath.Base(parts[0])
		default:
			filename = parts[0]
			path = parts[1]
		}

		info, err := os.Stat(path)
		if err != nil {
			return "", fmt.Errorf("stat attachment: %w", err)
		}
		if info.IsDir() {
			err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() {
					content, err := os.ReadFile(path)
					if err != nil {
						return fmt.Errorf("read attachment: %w", err)
					}

					data.WriteString(
						fmt.Sprintf(
							"%s: %s\n",
							path,
							base64.StdEncoding.EncodeToString(content)))
				}
				return nil
			})
			if err != nil {
				return "", fmt.Errorf("attachment walk: %w", err)
			}
		} else {
			content, err := os.ReadFile(path)
			if err != nil {
				return "", fmt.Errorf("read attachment: %w", err)
			}

			data.WriteString(
				fmt.Sprintf("%s: %s\n", filename, base64.StdEncoding.EncodeToString(content)))
		}
	}
	return data.String(), nil
}

func New(cfg *config.Config) *cobra.Command {
	var req openai.ChatCompletionRequest
	var conversation string
	var logItBias string
	var output string
	var attachments []string

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

			if req.TopLogProbs > 0 {
				req.LogProbs = true
			}

			if req.LogProbs {
				// obviously cant use "content" for output or you wouldn't see
				// the log probs you explicitly asked for
				output = "raw"
			}

			var writer chatcompletion.ResponseWriter
			switch output {
			case "content":
				writer = &chatcompletion.ContentResponseWriter{W: os.Stdout}
			case "raw":
				writer = &chatcompletion.RawResponseWriter{W: os.Stdout}
			case "recap":
				writer = &chatcompletion.RecapResponseWriter{W: os.Stdout}
			}

			if len(attachments) > 0 {
				for i := len(req.Messages) - 1; true; i-- {
					if i < 0 {
						return errors.New("no user message to append attements to")
					}
					if req.Messages[i].Role == openai.ChatMessageRoleUser {
						data, err := encodeAttachments(attachments)
						if err != nil {
							return err
						}
						req.Messages[i].Content = fmt.Sprintf(
							"%s\n\n########## base64 encoded attachments ##########\n%s",
							req.Messages[i].Content,
							data)
						break
					}
				}
			}

			if conversation == "" {
				err := mergo.Merge(&req, defaults)
				if err != nil {
					return fmt.Errorf("apply defaults: %w", err)
				}
				req.Messages = append(defaults.Messages, req.Messages...)

				err = chatcompletion.Send(ctx, client, req, writer)
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
					writer)
				if err != nil {
					return fmt.Errorf("complete chat: %w", err)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringArrayVar(
		&attachments,
		"attach",
		[]string{},
		""+
			"An attachment to add to the user message, these attachments will be base64 encoded and appended to the last user message. "+
			"The format of the attachment argument is [alias:]path where alias is optional and if not supplied the basename of path will be used. "+
			"If path is a directory, the directory will be recursively walked and all files encountered will be included.")
	cmd.Flags().StringVar(
		&conversation,
		"conversation",
		"",
		"A named conversation to start or continue")
	cmd.Flags().StringVar(
		&logItBias,
		"logit-bias",
		"",
		"A json map of string to int where they key is the token (can be obtained using: `askai tokens encode`) and the value is a bias between -100 (prohibit) and 100 (encourage)")
	cmd.Flags().BoolVar(
		&req.LogProbs,
		"logprobs",
		false,
		"Returns the log probabilities of each output token returned in the content of message")
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
	cmd.Flags().StringVar(
		&output,
		"output",
		"content",
		"Format of output, one of: content, raw, recap")
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
	cmd.Flags().IntVar(
		&req.TopLogProbs,
		"top-logprobs",
		0,
		""+
			"An integer between 0 and 5 specifying the number of most likely tokens to return at each token position, each with an associated log probability. "+
			"Implies --logprobs")

	return &cmd
}
