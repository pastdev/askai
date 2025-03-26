package chatcompletion

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os/exec"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
)

type Conversation interface {
	Continue(openai.ChatCompletionRequest) (openai.ChatCompletionRequest, error)
	UpdateResponse(string) error
}

func HandleBufferResponse(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer ResponseWriter,
) error {
	resp, err := client.CreateChatCompletion(ctx, req)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}

	log.Debug().Interface("resp", resp).Msg("before invoking tool")
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		err = handleToolCalls(ctx, client, req, resp, writer)
		if err != nil {
			return fmt.Errorf("handle tool calls: %w", err)
		}
	}

	err = writer.Write(resp)
	if err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func handleToolCalls(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	resp openai.ChatCompletionResponse,
	writer ResponseWriter,
) error {
	toolCalls := resp.Choices[0].Message.ToolCalls
	toolCallCompletionMessages := make([]openai.ChatCompletionMessage, 0, len(toolCalls))

	for _, toolCall := range toolCalls {
		log.Debug().Interface("toolCall", toolCall).Msg("invoking tool")

		args := []string{}
		if toolCall.Function.Arguments != "" {
			args = append(args, toolCall.Function.Arguments)
		}
		// prolly wanna have a whitelist here. not sure if openai api has any
		// safety guarantees, prolly not. ai _could_ just respond with a function
		// not in the list like rm --rf /. for now though, i am going to ignore
		// this and revisit when i have a more concrete case for using tools
		//   https://github.com/pastdev/askai/issues/4
		//nolint: gosec
		cmd := exec.CommandContext(ctx, toolCall.Function.Name, args...)
		outBuf := &bytes.Buffer{}
		errBuf := &bytes.Buffer{}
		cmd.Stdout = outBuf
		if log.Trace().Enabled() {
			// may need to loop over lines writing to log to avoid large buffer, but
			// for now, lets just do the _easy_ thing
			cmd.Stderr = errBuf
		}
		err := cmd.Run()
		log.Trace().
			Err(err).
			Str("stderr", errBuf.String()).
			Str("stdout", outBuf.String()).
			Msg("tool call complete")
		if err != nil {
			return fmt.Errorf("tool_call: %w", err)
		}

		toolCallCompletionMessages = append(
			toolCallCompletionMessages,
			openai.ChatCompletionMessage{
				Content: outBuf.String(),
				// appears from ollama example, that name is used instead of
				// tool_call_id to match:
				//   https://github.com/ollama/ollama-python/blob/aec125c77345b30d53309f5726226b5473159219/examples/tools.py#L77
				// this bug seems to confirm that:
				//   https://github.com/ollama/ollama/issues/7510
				Name:       toolCall.Function.Name,
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
			})
	}

	req.Messages = append(req.Messages, resp.Choices[0].Message)
	req.Messages = append(req.Messages, toolCallCompletionMessages...)
	err := Send(ctx, client, req, writer)
	if err != nil {
		return fmt.Errorf("tool call completion request: %w", err)
	}

	return nil
}

func HandleStreamResponse(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer ResponseWriter,
) error {
	strm, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("create completion stream: %w", err)
	}
	defer func() { _ = strm.Close() }()

	for {
		res, err := strm.Recv()
		if errors.Is(err, io.EOF) {
			log.Trace().Err(err).Msg("reached end of streaming response")
			return nil
		} else if err != nil {
			return fmt.Errorf("stream response: %w", err)
		}

		log.Trace().Interface("res", res).Msg("recieved stream chunk")
		err = writer.WriteStream(res)
		if err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}
}

func Send(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer ResponseWriter,
) error {
	var err error
	log.Debug().Bool("stream", req.Stream).Interface("messages", req.Messages).Msg("the messages")
	if req.Stream {
		err = HandleStreamResponse(ctx, client, req, writer)
	} else {
		err = HandleBufferResponse(ctx, client, req, writer)
	}
	if err != nil {
		return fmt.Errorf("handle response: %w", err)
	}
	return nil
}

func SendReply(
	ctx context.Context,
	client *openai.Client,
	conversation Conversation,
	reply openai.ChatCompletionRequest,
	writer ResponseWriter,
) error {
	req, err := conversation.Continue(reply)
	if err != nil {
		return fmt.Errorf("continue: %w", err)
	}

	buf := NewResponseWriterContentBuffer(writer)
	err = Send(ctx, client, req, buf)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	err = conversation.UpdateResponse(buf.String())
	if err != nil {
		return fmt.Errorf("update response: %w", err)
	}
	return nil
}
