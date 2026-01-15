package chatcompletion

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/openai/openai-go/v3"
)

const (
	WriteStreamStart WriteStreamPhase = iota
	WriteStreamContinue
	WriteStreamFinish
)

var _ ResponseWriter = &ContentResponseWriter{}
var _ ResponseWriter = &RawResponseWriter{}
var _ ResponseWriter = &RecapResponseWriter{}
var _ ResponseWriter = &ResponseWriterContentBuffer{}

type ContentResponseWriter struct {
	W io.Writer
}

type ResponseWriter interface {
	Write(*openai.ChatCompletion) error
	WriteRequest(openai.ChatCompletionNewParams) error
	WriteStream(openai.ChatCompletionChunk, WriteStreamPhase) error
}

type RawResponseWriter struct {
	W io.Writer
}

type RecapResponseWriter struct {
	W io.Writer
}

type WriteStreamPhase int

func (b *ContentResponseWriter) Write(res *openai.ChatCompletion) error {
	if len(res.Choices) < 1 {
		return nil
	}

	_, err := b.W.Write([]byte(res.Choices[0].Message.Content))
	if err != nil {
		return fmt.Errorf("contentresponsewriter write: %w", err)
	}
	return nil
}

func (b *ContentResponseWriter) WriteRequest(_ openai.ChatCompletionNewParams) error {
	return nil
}

func (b *ContentResponseWriter) WriteStream(
	res openai.ChatCompletionChunk,
	_ WriteStreamPhase,
) error {
	if len(res.Choices) < 1 {
		return nil
	}

	_, err := b.W.Write([]byte(res.Choices[0].Delta.Content))
	if err != nil {
		return fmt.Errorf("contentresponsewriter writestream: %w", err)
	}
	return nil
}

func (b *RawResponseWriter) Write(res *openai.ChatCompletion) error {
	err := json.NewEncoder(b.W).Encode(res)
	if err != nil {
		return fmt.Errorf("rawresponsewriter write: %w", err)
	}
	return nil
}

func (b *RawResponseWriter) WriteRequest(_ openai.ChatCompletionNewParams) error {
	return nil
}

func (b *RawResponseWriter) WriteStream(
	res openai.ChatCompletionChunk,
	_ WriteStreamPhase,
) error {
	err := json.NewEncoder(b.W).Encode(res)
	if err != nil {
		return fmt.Errorf("rawresponsewriter writestream: %w", err)
	}
	return nil
}

func (b *RecapResponseWriter) Write(res *openai.ChatCompletion) error {
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

func (b *RecapResponseWriter) WriteRequest(req openai.ChatCompletionNewParams) error {
	var err error
	for _, message := range req.Messages {
		switch {
		case message.OfAssistant != nil,
			message.OfSystem != nil,
			message.OfUser != nil:
			_, ierr := fmt.Fprintf(b.W, "%s: %s\n\n", *(message.GetRole()), message.GetContent())
			err = errors.Join(ierr)
		}
	}
	if err != nil {
		return fmt.Errorf("contentresponsewriter write request: %w", err)
	}
	return nil
}

func (b *RecapResponseWriter) WriteStream(
	res openai.ChatCompletionChunk,
	phase WriteStreamPhase,
) error {
	if len(res.Choices) < 1 {
		return nil
	}

	if phase == WriteStreamStart {
		_, err := fmt.Fprintf(b.W, "%s: ", res.Choices[0].Delta.Role)
		if err != nil {
			return fmt.Errorf("recapresponsewriter writestream start: %w", err)
		}
	}

	_, err := fmt.Fprint(b.W, res.Choices[0].Delta.Content)
	if err != nil {
		return fmt.Errorf("recapresponsewriter writestream: %w", err)
	}

	if phase == WriteStreamFinish {
		_, err := fmt.Fprint(b.W, "\n")
		if err != nil {
			return fmt.Errorf("recapresponsewriter writestream finish: %w", err)
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

func (b *ResponseWriterContentBuffer) Write(res *openai.ChatCompletion) error {
	err := b.w.Write(res)
	if err != nil {
		return fmt.Errorf("pass-thru write: %w", err)
	}

	_, _ = b.buf.Write([]byte(res.Choices[0].Message.Content))
	return nil
}

func (b *ResponseWriterContentBuffer) WriteRequest(req openai.ChatCompletionNewParams) error {
	err := b.w.WriteRequest(req)
	if err != nil {
		return fmt.Errorf("pass-thru write request: %w", err)
	}
	return nil
}

func (b *ResponseWriterContentBuffer) WriteStream(
	res openai.ChatCompletionChunk,
	phase WriteStreamPhase,
) error {
	err := b.w.WriteStream(res, phase)
	if err != nil {
		return fmt.Errorf("pass-thru write stream: %w", err)
	}

	_, _ = b.buf.Write([]byte(res.Choices[0].Delta.Content))
	return nil
}

func NewResponseWriterContentBuffer(w ResponseWriter) *ResponseWriterContentBuffer {
	return &ResponseWriterContentBuffer{w: w}
}
