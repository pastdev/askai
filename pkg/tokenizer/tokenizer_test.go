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
		tokens, err := tokenizer.Tokenize(model, text)
		if expectedErr != "" {
			require.EqualError(t, err, expectedErr)
		}
		require.Equal(t, expectedTokens, tokens)
	}

	t.Run("invalid model", func(t *testing.T) {
		tester(
			t,
			"invalidmodelname",
			"Hello world!",
			nil,
			"get encoding: no encoding for model invalidmodelname")
	})

	t.Run("gpt-3.5-turbo", func(t *testing.T) {
		tester(
			t,
			"gpt-3.5-turbo",
			"Hello world!",
			[]int{9906, 1917, 0},
			"")
	})
}
