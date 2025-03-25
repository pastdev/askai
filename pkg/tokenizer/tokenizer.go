package tokenizer

import (
	"errors"
	"fmt"

	"github.com/pastdev/askai/pkg/tokenizer/llama"
	"github.com/pkoukk/tiktoken-go"
)

type Tokenizer interface {
	Encode(text string, allowedSpecial []string, disallowedSpecial []string) []int
	Decode(data []int) string
}

func NewTokenizer(model string) (Tokenizer, error) {
	var errs error

	tke, err := tiktoken.EncodingForModel(model)
	if err == nil {
		return tke, nil
	}
	errs = errors.Join(errs, fmt.Errorf("tiktoken: %w", err))

	llama, err := llama.EncodingForModel(model)
	if err == nil {
		return llama, nil
	}
	errs = errors.Join(errs, fmt.Errorf("llama: %w", err))

	return nil, errs
}
