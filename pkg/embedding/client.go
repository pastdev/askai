package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
)

func HandleBufferResponse(
	ctx context.Context,
	client *openai.Client,
	req openai.EmbeddingRequest,
	writer io.Writer,
) error {
	var resp openai.EmbeddingResponse
	resp, err := client.CreateEmbeddings(ctx, req)
	if err != nil {
		return fmt.Errorf("embeddings: %w", err)
	}

	// cannot use yaml because of:
	//   https://github.com/go-yaml/yaml/issues/463
	out, err := json.Marshal(resp)
	if err != nil {
		return fmt.Errorf("marshal: %w", err)
	}

	_, err = fmt.Fprintf(writer, "%s\n", out)
	if err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func Send(
	ctx context.Context,
	client *openai.Client,
	req openai.EmbeddingRequest,
	writer io.Writer,
) error {
	log.Debug().Interface("input", req.Input).Msg("the input")
	err := HandleBufferResponse(ctx, client, req, writer)
	if err != nil {
		return fmt.Errorf("handle response: %w", err)
	}
	return nil
}
