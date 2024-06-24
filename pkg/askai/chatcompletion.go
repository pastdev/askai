package askai

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
)

type Conversation interface {
	Continue(...openai.ChatCompletionMessage) openai.ChatCompletionRequest
	UpdateResponse(string) error
}

func SendReply(
	ctx context.Context,
	client *openai.Client,
	conversation Conversation,
	messages []openai.ChatCompletionMessage,
	stream bool,
	writer io.Writer,
) error {
	var buf strings.Builder

	req := conversation.Continue(messages...)
	req.Stream = stream

	err := Send(ctx, client, req, io.MultiWriter(writer, &buf))
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	return conversation.UpdateResponse(buf.String())
}

func Send(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer io.Writer,
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

func HandleBufferResponse(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer io.Writer,
) error {
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return fmt.Errorf("chat completion: %w", err)
	}

	fmt.Fprintln(writer, resp.Choices[0].Message.Content)
	return nil
}

func HandleStreamResponse(
	ctx context.Context,
	client *openai.Client,
	req openai.ChatCompletionRequest,
	writer io.Writer,
) error {
	strm, err := client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return fmt.Errorf("create completion stream: %w", err)
	}
	defer func() { _ = strm.Close() }()

	for {
		res, err := strm.Recv()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			return fmt.Errorf("stream response: %w", err)
		}

		fmt.Fprintf(writer, res.Choices[0].Delta.Content)
	}
}
