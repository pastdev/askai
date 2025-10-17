package codereview

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

const (
	SystemPrompt = `you are performing a code review. user input will be in the form of a git
diff between the base and head commits that define a merge request.
if you require additional context to make sense of any of the diff
segments, use the supplied tools to request additional context.

response must be a json document of the form:
{
  "file": "foo/bar.sh",
  "line_start": 1,
  "line_end": 1,
  "suggestion": "include a shebang to identify the command to run"
}`
)

func diff(base, head string) (string, error) {
	return `diff --git a/duplicate_function..sh b/duplicate_function.sh
index 3f88e36..13553f0 100644
--- a/duplicate_function.sh
+++ b/duplicate_function.sh
@@ -0,0 +1,16 @@
+#!/bin/bash
+
+function foo {
+  printf "%q " "$@" | sed 's/ $//'
+}
+
+function bar {
+  printf "%q " "$@" | sed 's/ $//'
+}
+
+function main {
+  foo hip hop
+  bar hip hop
+}
+
+main
`, nil
}

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

			codeDiff, err := diff(base, head)
			if err != nil {
				return fmt.Errorf("diff: %w", err)
			}
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
