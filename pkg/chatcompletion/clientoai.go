package chatcompletion

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	oai "github.com/openai/openai-go/v3"
	"github.com/pastdev/askai/pkg/log"
)

type ConversationOai interface {
	Continue(oai.ChatCompletionNewParams) (oai.ChatCompletionNewParams, error)
	UpdateResponse(string) error
}

func HandleBufferResponseOai(
	ctx context.Context,
	client oai.Client,
	req oai.ChatCompletionNewParams,
	writer ResponseWriterOai,
) error {
	err := writer.WriteRequest(req)
	if err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	resp, err := client.Chat.Completions.New(ctx, req)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}

	log.Debug().Interface("resp", resp).Msg("before invoking tool")
	if len(resp.Choices[0].Message.ToolCalls) > 0 {
		err = handleToolCallsOai(ctx, client, req, resp, false, writer)
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

func handleToolCallsOai(
	ctx context.Context,
	client oai.Client,
	req oai.ChatCompletionNewParams,
	resp *oai.ChatCompletion,
	stream bool,
	writer ResponseWriterOai,
) error {
	toolCalls := resp.Choices[0].Message.ToolCalls
	toolCallCompletionMessages := make([]oai.ChatCompletionMessageParamUnion, 0, len(toolCalls))

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
			oai.ToolMessage(outBuf.String(), toolCall.ID))
	}

	req.Messages = append(req.Messages, resp.Choices[0].Message.ToParam())
	req.Messages = append(req.Messages, toolCallCompletionMessages...)
	err := SendOai(ctx, client, req, stream, writer)
	if err != nil {
		return fmt.Errorf("tool call completion request: %w", err)
	}

	return nil
}

func HandleStreamResponseOai(
	ctx context.Context,
	client oai.Client,
	req oai.ChatCompletionNewParams,
	writer ResponseWriterOai,
) error {
	err := writer.WriteRequest(req)
	if err != nil {
		return fmt.Errorf("write request: %w", err)
	}

	strm := client.Chat.Completions.NewStreaming(ctx, req)
	defer func() { _ = strm.Close() }()

	acc := oai.ChatCompletionAccumulator{}

	phase := WriteStreamStart
	for strm.Next() {
		chunk := strm.Current()
		acc.AddChunk(chunk)

		if content, ok := acc.JustFinishedContent(); ok {
			phase = WriteStreamFinish
			println("Content stream finished:", content)
		}

		// CODE_REVIEW_CATCH_ME: what do i need to do for these?
		// // if using tool calls
		// if tool, ok := acc.JustFinishedToolCall(); ok {
		// 	println("Tool call stream finished:", tool.Index, tool.Name, tool.Arguments)
		// }

		// if refusal, ok := acc.JustFinishedRefusal(); ok {
		// 	println("Refusal stream finished:", refusal)
		// }

		if len(chunk.Choices) > 0 {
			log.Trace().Interface("chunk", chunk).Msg("recieved stream chunk")
			err = writer.WriteStream(chunk, phase)
			if err != nil {
				return fmt.Errorf("write response: %w", err)
			}
		}

		phase = WriteStreamContinue
	}

	if strm.Err() != nil {
		return fmt.Errorf("create completion stream: %w", strm.Err())
	}

	return nil
}

func SendOai(
	ctx context.Context,
	client oai.Client,
	req oai.ChatCompletionNewParams,
	stream bool,
	writer ResponseWriterOai,
) error {
	var err error
	log.Debug().Bool("stream", stream).Interface("messages", req.Messages).Msg("the messages")
	if stream {
		err = HandleStreamResponseOai(ctx, client, req, writer)
	} else {
		err = HandleBufferResponseOai(ctx, client, req, writer)
	}
	if err != nil {
		return fmt.Errorf("handle response: %w", err)
	}
	return nil
}

func SendReplyOai(
	ctx context.Context,
	client oai.Client,
	conversation ConversationOai,
	reply oai.ChatCompletionNewParams,
	stream bool,
	writer ResponseWriterOai,
) error {
	req, err := conversation.Continue(reply)
	if err != nil {
		return fmt.Errorf("continue: %w", err)
	}

	buf := NewResponseWriterContentBufferOai(writer)
	err = SendOai(ctx, client, req, stream, buf)
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	err = conversation.UpdateResponse(buf.String())
	if err != nil {
		return fmt.Errorf("update response: %w", err)
	}
	return nil
}
