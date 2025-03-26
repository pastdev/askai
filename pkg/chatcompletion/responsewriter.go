package chatcompletion

import (
	"encoding/json"
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
	WriteStream(openai.ChatCompletionStreamResponse) error
}

type RawResponseWriter struct {
	W io.Writer
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

func (b *RawResponseWriter) WriteStream(res openai.ChatCompletionStreamResponse) error {
	err := json.NewEncoder(b.W).Encode(res)
	if err != nil {
		return fmt.Errorf("rawresponsewriter writestream: %w", err)
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
