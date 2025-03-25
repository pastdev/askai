package tokenizer_test

import (
	"testing"

	"github.com/pastdev/askai/pkg/tokenizer"
	"github.com/stretchr/testify/require"
)

func TestTokenize(t *testing.T) {
	tester := func(
		t *testing.T,
		model string,
		text string,
		expectedTokens []int,
		expectedErr string,
	) {
		tokenizer, err := tokenizer.NewTokenizer(model)
		if expectedErr != "" {
			require.EqualError(t, err, expectedErr)
			return
		}
		require.NoError(t, err)
		tokens := tokenizer.Encode(text, nil, nil)
		require.Equal(t, expectedTokens, tokens)
	}

	t.Run("invalid model", func(t *testing.T) {
		tester(
			t,
			"invalidmodelname",
			"Hello world!",
			nil,
			"tiktoken: no encoding for model invalidmodelname\nllama: no encoding for model invalidmodelname")
	})

	t.Run("gpt-3.5-turbo", func(t *testing.T) {
		tester(
			t,
			"gpt-3.5-turbo",
			"Hello world!",
			[]int{9906, 1917, 0},
			"")
	})

	t.Run("meta-llama/llama-3.2-11b-vision-instruct", func(t *testing.T) {
		tester(
			t,
			"meta-llama/llama-3.2-11b-vision-instruct",
			"Hello world!",
			[]int{9906, 1917, 0},
			"")
	})
}
