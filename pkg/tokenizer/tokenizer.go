package tokenizer

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

func Tokenize(model string, text string) ([]int, error) {
	tke, err := tiktoken.EncodingForModel(model)
	if err != nil {
		return nil, fmt.Errorf("get encoding: %w", err)
	}

	return tke.Encode(text, nil, nil), nil
}
