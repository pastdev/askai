package llama

import (
	"fmt"
	"iter"
	"path"
	"strings"
	"unicode"

	"github.com/pkoukk/tiktoken-go"
)

const (
	numReservedSpecialTokens = 256
	patStr                   = `(?i:'s|'t|'re|'ve|'m|'ll|'d)|[^\r\n\p{L}\p{N}]?\p{L}+|\p{N}{1,3}| ?[^\s\p{L}\p{N}]+[\r\n]*|\s*[\r\n]+|\s+(?!\S)|\s+`
	// maxNoWhitespaceChars may not be necessary but the
	// [upstream python implementation does it], and the reasoning it gives is
	// related to [this issue] which could reasonably be assumed to apply to the
	// go implementation as it is logic based not necessarily language based.
	//
	// [upstream python implementation does it]: https://github.com/meta-llama/llama-models/blob/364ed4ec2bdcc13ab3acb08e364e843b6322b9ae/models/llama3/api/tokenizer.py#L38-L41
	// [this issue]: https://github.com/openai/tiktoken/issues/195
	maxNoWhitespacesChars = 25_000
)

// based off of [meta-llama implementation]
//
// [meta-llama implementation]: https://github.com/meta-llama/llama-models/blob/ee8ede3deb42dbfc8e151fd303c1cd9acdfe1869/models/llama3/api/tokenizer.py#L47
type Tokenizer struct {
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
	tiktoken      *tiktoken.Tiktoken
}

type EncodeOption func(*EncodeOptions)

type EncodeOptions struct {
	bos bool
	eos bool
}

func (e Tokenizer) Decode(tokens []int) string {
	return e.tiktoken.Decode(tokens)
}

func (e Tokenizer) Encode(
	text string,
	allowedSpecial []string,
	disallowedSpecial []string,
) []int {
	return e.EncodeWithOptions(text, allowedSpecial, disallowedSpecial)
}

func (e Tokenizer) EncodeWithOptions(
	text string,
	allowedSpecial []string,
	disallowedSpecial []string,
	opts ...EncodeOption,
) []int {
	encodeOpts := EncodeOptions{}
	for _, opt := range opts {
		opt(&encodeOpts)
	}

	t := make([]int, 0)
	for segment := range splitWhitespaceOrNonWhitespace(text, maxNoWhitespacesChars) {
		t = append(t, e.tiktoken.Encode(segment, allowedSpecial, disallowedSpecial)...)
	}

	if encodeOpts.bos {
		t = append([]int{e.bosID}, t...)
	}
	if encodeOpts.eos {
		t = append(t, e.eosID)
	}

	return t
}

func EncodingForModel(modelName string) (*Tokenizer, error) {
	switch strings.ToLower(modelName) {
	case "meta-llama/llama-3.2-11b-vision-instruct",
		"meta-llama/llama-3.3-70b-instruct":
		return NewTokenizer()
	default:
		return nil, fmt.Errorf("no encoding for model %s", modelName)
	}
}

// NewTokenizer constructs a new Llama compatible tokenizer.
func NewTokenizer() (*Tokenizer, error) {
	modelPath := "https://raw.githubusercontent.com/meta-llama/llama-models/v0.1.4/models/llama3/api/tokenizer.model"
	tokenizer := Tokenizer{
		specialTokens: make(map[string]int),
	}

	mergeableRanks, err := tiktoken.
		NewDefaultBpeLoader().
		LoadTiktokenBpe(modelPath)
	if err != nil {
		return nil, fmt.Errorf("load tiktoken bpe: %w", err)
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
		tokenizer.specialTokens[token] = numBaseTokens + i
	}

	tokenizer.model = &tiktoken.Encoding{
		Name:           path.Base(modelPath),
		PatStr:         patStr,
		MergeableRanks: mergeableRanks,
		SpecialTokens:  tokenizer.specialTokens,
	}

	tokenizer.nWords = numBaseTokens + len(specialTokens)
	// BOS / EOS token IDs
	tokenizer.bosID = tokenizer.specialTokens["<|begin_of_text|>"]
	tokenizer.eosID = tokenizer.specialTokens["<|end_of_text|>"]
	tokenizer.eotID = tokenizer.specialTokens["<|eot_id|>"]
	tokenizer.eomID = tokenizer.specialTokens["<|eom_id|>"]
	tokenizer.pythonTagID = tokenizer.specialTokens["<|python_tag|>"]
	tokenizer.padID = tokenizer.specialTokens["<|finetune_right_pad_id|>"]
	tokenizer.stopTokens = []int{
		tokenizer.eosID,
		tokenizer.specialTokens["<|eom_id|>"],
		tokenizer.specialTokens["<|eot_id|>"],
	}

	pbe, err := tiktoken.NewCoreBPE(mergeableRanks, tokenizer.specialTokens, patStr)
	if err != nil {
		return nil, fmt.Errorf("newcorebpe: %w", err)
	}

	specialTokensSet := map[string]any{}
	for k := range tokenizer.specialTokens {
		specialTokensSet[k] = true
	}

	tokenizer.tiktoken = tiktoken.NewTiktoken(
		pbe,
		tokenizer.model,
		specialTokensSet)

	return &tokenizer, nil
}

// splitWhitespaceOrNonWhtespace will splits the string s so that each substring
// contains no more than maxConsecutiveSliceLen characters. This is a direct
// port of the [upstream python code].
//
// [upstream python code]: https://github.com/meta-llama/llama-models/blob/364ed4ec2bdcc13ab3acb08e364e843b6322b9ae/models/llama3/api/tokenizer.py#L187-L209
func splitWhitespaceOrNonWhitespace(s string, maxConsecutiveSliceLen int) iter.Seq[string] {
	return func(yield func(string) bool) {
		currentSliceLen := 0
		currentSliceIsSpace := false
		if len(s) == 0 {
			currentSliceIsSpace = unicode.IsSpace(rune(s[0]))
		}
		sliceStart := 0

		for i, v := range s {
			isNowSpace := unicode.IsSpace(rune(v))
			if currentSliceIsSpace != isNowSpace {
				currentSliceLen = 1
				currentSliceIsSpace = isNowSpace
			} else {
				currentSliceLen++
				if currentSliceLen > maxConsecutiveSliceLen {
					if !yield(s[sliceStart:i]) {
						return
					}
					sliceStart = i
					currentSliceLen = 1
				}
			}
		}
		yield(s[sliceStart:])
	}
}

func WithBos() EncodeOption {
	return func(eo *EncodeOptions) {
		eo.bos = true
	}
}

func WithEos() EncodeOption {
	return func(eo *EncodeOptions) {
		eo.eos = true
	}
}
