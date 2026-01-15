package embedding

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	oai "github.com/openai/openai-go/v3"
	"github.com/pastdev/askai/pkg/log"
)

func HandleBufferResponse(
	ctx context.Context,
	client oai.Client,
	req oai.EmbeddingNewParams,
	writer io.Writer,
) error {
	var resp *oai.CreateEmbeddingResponse
	resp, err := client.Embeddings.New(ctx, req)
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
	client oai.Client,
	req oai.EmbeddingNewParams,
	writer io.Writer,
) error {
	log.Debug().Interface("input", req.Input).Msg("the input")
	err := HandleBufferResponse(ctx, client, req, writer)
	if err != nil {
		return fmt.Errorf("handle response: %w", err)
	}
	return nil
}
