package image

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/pastdev/askai/pkg/log"
)

func Send(
	ctx context.Context,
	client openai.Client,
	req openai.ImageGenerateParams,
	writer ResponseWriter,
) error {
	log.Debug().Interface("req", req).Msg("the request")
	resp, err := client.Images.Generate(ctx, req)
	if err != nil {
		return fmt.Errorf("send create image: %w", err)
	}

	err = writer.Write(resp)
	if err != nil {
		return fmt.Errorf("send write response: %w", err)
	}
	return nil
}
