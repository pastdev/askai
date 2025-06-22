package image

import (
	"context"
	"fmt"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
)

func Send(
	ctx context.Context,
	client *openai.Client,
	req openai.ImageRequest,
	writer ResponseWriter,
) error {
	log.Debug().Interface("req", req).Msg("the request")
	resp, err := client.CreateImage(ctx, req)
	if err != nil {
		return fmt.Errorf("send create image: %w", err)
	}

	err = writer.Write(resp)
	if err != nil {
		return fmt.Errorf("send write response: %w", err)
	}
	return nil
}
