package chatcompletion

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/sashabaranov/go-openai"
)

var _ ResponseWriter = &ContentResponseWriter{}

type ContentResponseWriter struct {
	W io.Writer
}

type ResponseWriter interface {
	Write(openai.ChatCompletionResponse) error
	WriteRequest(openai.ChatCompletionRequest) error
	WriteStream(openai.ChatCompletionStreamResponse) error
}

type RawResponseWriter struct {
	W io.Writer
}

type RecapResponseWriter struct {
	streamWriteCount int
	W                io.Writer
}

func (b *ContentResponseWriter) Write(res openai.ChatCompletionResponse) error {
	if len(res.Choices) < 1 {
		return nil
	}

	_, err := b.W.Write([]byte(res.Choices[0].Message.Content))
	if err != nil {
		return fmt.Errorf("contentresponsewriter write: %w", err)
	}
	return nil
}

func (b *ContentResponseWriter) WriteRequest(_ openai.ChatCompletionRequest) error {
	return nil
}

func (b *ContentResponseWriter) WriteStream(res openai.ChatCompletionStreamResponse) error {
	if len(res.Choices) < 1 {
		return nil
	}

	_, err := b.W.Write([]byte(res.Choices[0].Delta.Content))
	if err != nil {
		return fmt.Errorf("contentresponsewriter writestream: %w", err)
	}
	return nil
}

func (b *RawResponseWriter) Write(res openai.ChatCompletionResponse) error {
	err := json.NewEncoder(b.W).Encode(res)
	if err != nil {
		return fmt.Errorf("rawresponsewriter write: %w", err)
	}
	return nil
}

func (b *RawResponseWriter) WriteRequest(_ openai.ChatCompletionRequest) error {
	return nil
}

func (b *RawResponseWriter) WriteStream(res openai.ChatCompletionStreamResponse) error {
	err := json.NewEncoder(b.W).Encode(res)
	if err != nil {
		return fmt.Errorf("rawresponsewriter writestream: %w", err)
	}
	return nil
}

func (b *RecapResponseWriter) Write(res openai.ChatCompletionResponse) error {
	if len(res.Choices) < 1 {
		return nil
	}

	_, err := fmt.Fprintf(
		b.W,
		"%s: %s\n\n",
		res.Choices[0].Message.Role,
		res.Choices[0].Message.Content)
	if err != nil {
		return fmt.Errorf("contentresponsewriter write: %w", err)
	}
	return nil
}

func (b *RecapResponseWriter) WriteRequest(req openai.ChatCompletionRequest) error {
	var err error
	for _, message := range req.Messages {
		switch message.Role {
		case openai.ChatMessageRoleAssistant,
			openai.ChatMessageRoleSystem,
			openai.ChatMessageRoleUser:
			_, ierr := fmt.Fprintf(b.W, "%s: %s\n\n", message.Role, message.Content)
			err = errors.Join(ierr)
		}
	}
	if err != nil {
		return fmt.Errorf("contentresponsewriter write request: %w", err)
	}
	return nil
}

func (b *RecapResponseWriter) WriteStream(res openai.ChatCompletionStreamResponse) error {
	if len(res.Choices) < 1 {
		return nil
	}

	if b.streamWriteCount == 0 {
		_, err := fmt.Fprintf(b.W, "%s: ", res.Choices[0].Delta.Role)
		if err != nil {
			return fmt.Errorf("recapresponsewriter writestream start: %w", err)
		}
	}

	_, err := fmt.Fprint(b.W, res.Choices[0].Delta.Role)
	if err != nil {
		return fmt.Errorf("recapresponsewriter writestream: %w", err)
	}

	if res.Choices[0].FinishReason != openai.FinishReasonNull {
		_, err := fmt.Fprintf(b.W, "%s: ", res.Choices[0].Delta.Role)
		if err != nil {
			return fmt.Errorf("recapresponsewriter writestream end: %w", err)
		}
	}

	return nil
}

type ResponseWriterContentBuffer struct {
	w   ResponseWriter
	buf strings.Builder
}

func (b *ResponseWriterContentBuffer) String() string {
	return b.buf.String()
}

func (b *ResponseWriterContentBuffer) Write(res openai.ChatCompletionResponse) error {
	err := b.w.Write(res)
	if err != nil {
		return fmt.Errorf("pass-thru write: %w", err)
	}

	_, _ = b.buf.Write([]byte(res.Choices[0].Message.Content))
	return nil
}

func (b *ResponseWriterContentBuffer) WriteRequest(req openai.ChatCompletionRequest) error {
	err := b.w.WriteRequest(req)
	if err != nil {
		return fmt.Errorf("pass-thru write request: %w", err)
	}
	return nil
}

func (b *ResponseWriterContentBuffer) WriteStream(res openai.ChatCompletionStreamResponse) error {
	err := b.w.WriteStream(res)
	if err != nil {
		return fmt.Errorf("pass-thru write stream: %w", err)
	}

	_, _ = b.buf.Write([]byte(res.Choices[0].Delta.Content))
	return nil
}

func NewResponseWriterContentBuffer(w ResponseWriter) *ResponseWriterContentBuffer {
	return &ResponseWriterContentBuffer{w: w}
}
