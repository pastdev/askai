package llama

import (
	"fmt"
	"path"

	"github.com/pkoukk/tiktoken-go"
)

const (
	numReservedSpecialTokens = 256
	patStr                   = `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`
)

type Encoder interface {
	Encode(text string, allowedSpecial []string, disallowedSpecial []string) []int
}

// based off of [meta-llama implementation]
//
// [meta-llama implementation]: https://github.com/meta-llama/llama-models/blob/ee8ede3deb42dbfc8e151fd303c1cd9acdfe1869/models/llama3/api/tokenizer.py#L47
type llamaEncoder struct {
	bosID         int
	eosID         int
	eotID         int
	eomID         int
	pythonTagID   int
	padID         int
	stopTokens    []int
	model         *tiktoken.Encoding
	nWords        int
	specialTokens map[string]int
}

func (e llamaEncoder) Encode(
	text string,
	allowedSpecial []string,
	disallowedSpecial []string,
) []int {
	e.model.
	return nil
}

// NewLlamaEncoder constructs a new Llama compatible encoder. Can use url to
// model file like this:
//
//	NewLlamaEncoder("https://raw.githubusercontent.com/meta-llama/llama-models/ee8ede3deb42dbfc8e151fd303c1cd9acdfe1869/models/llama3/api/tokenizer.model")
func NewLlamaEncoder(modelPath string) (Encoder, error) {
	enc := llamaEncoder{
		specialTokens: make(map[string]int),
	}

	mergeableRanks, err := tiktoken.
		NewDefaultBpeLoader().
		LoadTiktokenBpe(modelPath)
	if err != nil {
		fmt.Errorf("load tiktoken bpe: %w", err)
	}

	numBaseTokens := len(mergeableRanks)
	specialTokens := []string{
		"<|begin_of_text|>",
		"<|end_of_text|>",
		"<|reserved_special_token_0|>",
		"<|reserved_special_token_1|>",
		"<|finetune_right_pad_id|>",
		"<|step_id|>",
		"<|start_header_id|>",
		"<|end_header_id|>",
		"<|eom_id|>", // end of message
		"<|eot_id|>", // end of turn
		"<|python_tag|>",
		"<|image|>",
	}

	for i := 0; i < (numReservedSpecialTokens - len(specialTokens)); i++ {
		specialTokens = append(
			specialTokens,
			fmt.Sprintf("<|reserved_special_token_%d|>", 2+i))
	}

	for i, token := range specialTokens {
		enc.specialTokens[token] = i
	}

	enc.model = tiktoken.Encoding{
		Name:           path.Base(modelPath),
		PatStr:         patStr,
		MergeableRanks: mergeableRanks,
		SpecialTokens:  enc.specialTokens,
	}

	enc.nWords = numBaseTokens + len(specialTokens)
	// BOS / EOS token IDs
	enc.bosID = enc.specialTokens["<|begin_of_text|>"]
	enc.eosID = enc.specialTokens["<|end_of_text|>"]
	enc.eotID = enc.specialTokens["<|eot_id|>"]
	enc.eomID = enc.specialTokens["<|eom_id|>"]
	enc.pythonTagID = enc.specialTokens["<|python_tag|>"]
	enc.padID = enc.specialTokens["<|finetune_right_pad_id|>"]
	enc.stopTokens = []int{
		enc.eosID,
		enc.specialTokens["<|eom_id|>"],
		enc.specialTokens["<|eot_id|>"],
	}

	return enc, nil
}

func EncodingForModel(modelName string) (Encoder, error) {
	switch modelName {
	case "meta-llama/llama-3.2-11b-vision-instruct",
		"meta-llama/Llama-3.3-70B-Instruct":
		return LlamaEncoder{}, nil
	}
	return nil, fmt.Errorf("no encoding for model %s", modelName)
}
