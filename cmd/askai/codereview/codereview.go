package codereview

import (
	"context"
	"fmt"
	"os"

	"dario.cat/mergo"
	"github.com/pastdev/askai/cmd/askai/config"
	"github.com/pastdev/askai/pkg/chatcompletion"
	"github.com/pastdev/askai/pkg/git"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/cobra"
)

const (
	SystemPrompt = `# Code Review Prompt
You are performing a code review as a senior developer of the attached diff. If accepted it will be merged into production code. Your feedback should be pragmatic, focusing on what is most important for a healthy, secure, and maintainable codebase.

## Your primary objectives are to:

* Identify Critical Flaws: Pinpoint any actual bugs, security vulnerabilities (like SQL injection, XSS, insecure dependency usage, etc.), or significant logical errors that could cause outages or incorrect behavior.
* Assess Production Readiness: Evaluate whether the code is robust enough for a production environment. Consider error handling, edge cases, potential performance bottlenecks, and scalability concerns.
* Ensure Long-Term Maintainability: Check for clarity, readability, and adherence to idiomatic language conventions. Is the code's intent clear? Could a new developer understand it quickly? If not, are comments present to explain the "why," and not "what"?
* Suggest High-Impact Refinements: Propose concrete improvements or refactoring only if the long-term benefits (e.g., significant simplification, performance gains, or improved readability) clearly outweigh the effort required to implement them.

## What to ignore:

* Style & Linting: Do not comment on anything a linter or code formatter would automatically flag (e.g., whitespace, line length, naming conventions unless they are actively misleading).
* Project Configuration: Ignore boilerplate, dependency versions, or project setup files unless they introduce a direct conflict or security risk.
* Vague Generalities: Avoid generic advice like "consider performance." Instead, point to a specific line or block and explain why it might be a performance issue.
* Explicitly ignored findings with justification: Code may be annotated with a label indicating a finding should be ignored (ie: 'nolint: errcheck'). This indicates the finding has already been evaluated and does not need further comment.

## Output format:

Provide your review in JSON format following the schema:

~~~json
{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "$id": "https://example.com/product.schema.json",
  "description": "a code review",
  "properties": {
    "issues": {
      "description": "a list of all issues found in the code review. if no issues are found, return an empty list.",
      "items": [
        "type": "object"
        "properties": {
          "comment": {
            "description": "a detailed comment explaining the issue. structure your comment to clearly state the problem, its impact, and a recommended solution.",
            "examples": [
              "Problem: The code uses a traditional for-loop to iterate over an array.\nImpact: While functional, this is not the most idiomatic or readable approach in Go. It is more verbose and slightly more prone to off-by-one errors.\nSolution: Use a ` + "`" + `for ... range` + "`" + ` loop for simpler, more declarative iteration.",
            ],
            "type": "string"
          },
          "corrected_code": {
            "description": "a corrected version of the code snippet. provide this if it's the clearest way to communicate the suggested change.",
            "examples": [
              "for _, item := range items {\n\tprocessItem(item);\n}"
            ],
            "type": "string"
          },
          "hunk_headers": {
            "description": "the diff hunk header shown prior to the block of code that is being commented on. should not contain the body of the diff hunk",
            "examples": [
              "diff --git a/cmd/askai/codereview/codereview.go b/cmd/askai/codereview/codereview.go\nnew file mode 100644\nindex 0000000..b9e09f4\n--- /dev/null\n+++ b/cmd/askai/codereview/codereview.go\n@@ -0,0 +1,140 @@"
            ],
            "type": "string"
          },
          "severity": {
            "description": "the severity of the issue.\n- blocker: Must be fixed before merge (e.g., bugs, security flaws).\n- suggestion: Recommended improvement (e.g., refactoring for clarity).\n- nitpick: Minor, non-critical feedback.",
            "enum": [
              "blocker",
              "suggestion",
              "nitpick"
            ],
            "type": "string"
          },
          "snippet": {
            "description": "an exact, possibly multi-line snippet of the code for which this note applies.",
            "examples": [
              "for (int i = 0; i < items.len; i++) {\n\tprocessItem(items[i]);\n}"
            ],
            "type": "string"
          }
        }
        "required": [ "comment", "hunk_headers", "severity", "snippet" ]
      ],
      "type": "array"
    }
  },
  "title": "code_review",
  "type": "object"
}
~~~

Do not include fences (ie: triple backtick) around the json, just output pure json as it will be parsed directly.
`
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

			req.Tools = []openai.Tool{}

			err = mergo.Merge(&req, defaults)
			if err != nil {
				return fmt.Errorf("apply defaults: %w", err)
			}

			codeDiff, err := git.Diff(".", base, head)
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
