package codereview

import (
	"context"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/chatcompletion"
	"github.com/pastdev/askai/pkg/git"
	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const (
	SystemPrompt = `You are acting as a senior developer doing a first-pass code review.
Your task is to find issues with code quality that could impact long term maintenance of the code base and provide feedback or suggestions that MUST, SHOULD, or COULD be taken before the code is accepted into the primary branch.
Look for bugs, code smells, security issues and any other issues with the code and determine a level at which they apply.
For example, a potential SQL injection would rate a MUST, but renaming a variable for clarity may be a COULD.
Take the most modern best practices into account for the languages in question.
Also take into account typical best practices that apply to any language.
When commenting, provide the "why" in addition to the "what".

The input format is that of a git diff with an overarching line number prefixed to the entire patch.
Use this prefix line number to indicate where any suggestions are intended.

Response must be a JSON document of the form:
{
  "diff": "...",
  "file": "foo/bar.sh",
  "line_start": 1,
  "line_end": 1,
  "suggestion": "include a shebang to identify the command to run"
}
diff MUST contain the full diff chunk that this suggestion refers to.
line_start and line_end MUST take into account the line range supplied for the diff chunk.
if suggesting addition of a new line of code, the line_start and line_end should have the same value and indicate where the insert should occur.`
)

func New(cfg *config.Config) *cobra.Command {
	var req openai.ChatCompletionRequest

	cmd := cobra.Command{
		Use:   "cr",
		Short: `Ask AI to perform a code review`,
		Args:  cobra.ExactArgs(2),
		Example: `  # review between 2 commits
  askai cr e43ddd4e1848df08dc0141d0abe8eb544b58878a 0a8a46974a721caa2a0275b442980d26e2a94227
`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			base := args[0]
			head := args[1]

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

			codeDiff, err := git.Diff(".", base, head)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}
			codeDiff = git.PrefixDiff(codeDiff)
			// CODE_REVIEW_CATCH_ME: remove this
			fmt.Println(codeDiff)
			return nil
			log.Trace().Str("diff", codeDiff).Msg("git diff")
			req.Messages = []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: codeDiff,
				},
			}

			req.Tools = []openai.Tool{}

			err = mergo.Merge(&req, defaults)
			if err != nil {
				return fmt.Errorf("apply defaults: %w", err)
			}
			req.Messages = append(defaults.Messages, req.Messages...)

			err = chatcompletion.Send(
				ctx,
				client,
				req,
				&chatcompletion.ContentResponseWriter{W: os.Stdout})
			if err != nil {
				return fmt.Errorf("complete chat: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(
		&req.Model,
		"model",
		"",
		"AI model to use")

	return &cmd
}
