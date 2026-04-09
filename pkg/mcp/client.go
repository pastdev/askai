package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/packages/param"
	"github.com/openai/openai-go/v3/shared"
	"github.com/pastdev/askai/pkg/log"
	"github.com/pastdev/askai/pkg/version"
)

type Client struct {
	autoConfirm       bool
	confirmDefaultYes bool
	name              string
	session           *mcp.ClientSession
}

type Option func(*Client)

func (c *Client) ChatCompletionToolsCall(
	ctx context.Context,
	call openai.ChatCompletionMessageToolCallUnion,
) (openai.ChatCompletionMessageParamUnion, error) {
	if !c.Supports(call.Function.Name) {
		return openai.ChatCompletionMessageParamUnion{},
			fmt.Errorf("mcp %s does not support %s", c.name, c.FunctionName(call))
	}

	err := c.confirm(call)
	if err != nil {
		return openai.ChatCompletionMessageParamUnion{}, err
	}

	log.Info().
		Str("Tool", c.name).
		Str("Function", c.FunctionName(call)).
		Str("Arguments", call.Function.Arguments).
		Msg("invoking mcp tool")

	var argsObj any
	err = json.Unmarshal([]byte(call.Function.Arguments), &argsObj)
	if err != nil {
		return openai.ChatCompletionMessageParamUnion{},
			fmt.Errorf("mcp tool parse args: %w", err)
	}

	toolCallParams := &mcp.CallToolParams{
		Name:      c.FunctionName(call),
		Arguments: argsObj,
	}
	log.Trace().Interface("toolCallParams", toolCallParams).Msg("execute mcp tool call")
	resp, err := c.session.CallTool(ctx, toolCallParams)
	if err != nil {
		return openai.ChatCompletionMessageParamUnion{}, fmt.Errorf("mcp tools call: %w", err)
	}

	var data []byte
	if resp.StructuredContent != nil {
		data, err = json.Marshal(resp.StructuredContent)
	} else {
		data, err = json.Marshal(resp.Content)
	}
	if err != nil {
		return openai.ChatCompletionMessageParamUnion{},
			fmt.Errorf("mcp tools call marshal: %w", err)
	}

	return openai.ToolMessage(string(data), call.ID), nil
}

func (c *Client) ChatCompletionToolsList(
	ctx context.Context,
) ([]openai.ChatCompletionToolUnionParam, error) {
	tools, err := c.session.ListTools(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("mcp tools list: %w", err)
	}

	ccTools := []openai.ChatCompletionToolUnionParam{}
	for _, tool := range tools.Tools {
		params, ok := tool.InputSchema.(map[string]any)
		if !ok {
			// skip
			continue
		}

		ccTools = append(
			ccTools,
			openai.ChatCompletionFunctionTool(
				shared.FunctionDefinitionParam{
					Name:        fmt.Sprintf("%s__%s", c.name, tool.Name),
					Strict:      param.NewOpt(true),
					Description: param.NewOpt(tool.Description),
					Parameters:  params,
				}))
	}

	return ccTools, nil
}

func (c Client) confirm(call openai.ChatCompletionMessageToolCallUnion) error {
	if c.autoConfirm {
		return nil
	}

	r := bufio.NewReader(os.Stdin)

	suffix := "[y/N]"
	if c.confirmDefaultYes {
		suffix = "[Y/n]"
	}

	for {
		fmt.Fprintf(
			os.Stderr,
			"\n%s would like to invoke %s(%s)\nIs this ok %s:",
			c.name,
			c.FunctionName(call),
			call.Function.Arguments,
			suffix)

		line, err := r.ReadString('\n')
		if err != nil {
			return fmt.Errorf("read tool confirmation: %w", err)
		}

		s := strings.ToLower(strings.TrimSpace(line))
		switch {
		case s == "n" || s == "no" || (s == "" && !c.confirmDefaultYes):
			return errors.New("tool call denied by user")
		case s == "y" || s == "yes" || (s == "" && c.confirmDefaultYes):
			return nil
		}
	}
}

func (c *Client) FunctionName(call openai.ChatCompletionMessageToolCallUnion) string {
	return call.Function.Name[len(c.prefix()):]
}

func (c *Client) Name() string {
	return c.name
}

func (c *Client) prefix() string {
	return fmt.Sprintf("%s__", c.name)
}

func (c *Client) Supports(toolName string) bool {
	return strings.HasPrefix(toolName, c.prefix())
}

func NewClient(
	ctx context.Context,
	name string,
	endpointURL string,
	opts ...Option,
) (*Client, error) {
	session, err := mcp.
		NewClient(
			&mcp.Implementation{
				Name:    "askai",
				Version: version.Version()},
			nil).
		Connect(
			ctx,
			&mcp.StreamableClientTransport{
				Endpoint: endpointURL,
			},
			nil)
	if err != nil {
		return nil, fmt.Errorf("mcp connect: %w", err)
	}

	c := &Client{
		name:    name,
		session: session,
	}
	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

func WithAutoConfirm(autoConfirm bool) Option {
	return func(c *Client) {
		c.autoConfirm = autoConfirm
	}
}

func WithConfirmDefaultYes(confirmDefaultYes bool) Option {
	return func(c *Client) {
		c.confirmDefaultYes = confirmDefaultYes
	}
}
