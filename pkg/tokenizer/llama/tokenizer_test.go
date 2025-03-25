package llama

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDecode(t *testing.T) {
	tester := func(t *testing.T, data []int, expected string) {
		tok, err := NewTokenizer()
		require.Nil(t, err)
		require.EqualValues(t, expected, tok.Decode(data))
	}

	t.Run("decode", func(t *testing.T) {
		tester(
			t,
			[]int{128000, 2028, 374, 264, 1296, 11914, 13, 128001},
			"<|begin_of_text|>This is a test sentence.<|end_of_text|>")
	})
}

func TestEncode(t *testing.T) {
	tester := func(t *testing.T, s string, expected []int, opts ...EncodeOption) {
		tok, err := NewTokenizer()
		require.Nil(t, err)
		require.EqualValues(t, expected, tok.EncodeWithOptions(s, nil, nil, opts...))
	}

	t.Run("encode", func(t *testing.T) {
		tester(
			t,
			"This is a test sentence.",
			[]int{128000, 2028, 374, 264, 1296, 11914, 13, 128001},
			WithBos(),
			WithEos())
	})
}

func TestSplitWhitespaceOrNonWhitespace(t *testing.T) {
	tester := func(t *testing.T, s string, maxConsecutiveSliceLen int, expected []string) {
		require.EqualValues(
			t,
			expected,
			slices.Collect(splitWhitespaceOrNonWhitespace(s, maxConsecutiveSliceLen)))
	}

	// (.venv)
	// ltheisen@ltserver ~/git/github.com/meta-llama/llama-models
	// $ python3 -c "import models.llama3.api; print([i for i in models.llama3.api.Tokenizer('models/llama3/api/tokenizer.model')._split_whitespaces_or_nonwhitespaces('foo bar baz', 3)]);"
	// ['foo bar baz']
	t.Run("no split", func(t *testing.T) {
		tester(t, "foo bar baz", 3, []string{"foo bar baz"})
	})

	// (.venv)
	// ltheisen@ltserver ~/git/github.com/meta-llama/llama-models
	// $ python3 -c "import models.llama3.api; print([i for i in models.llama3.api.Tokenizer('models/llama3/api/tokenizer.model')._split_whitespaces_or_nonwhitespaces('foo bar baz', 1)]);"
	// ['f', 'o', 'o b', 'a', 'r b', 'a', 'z']
	t.Run("split every 1", func(t *testing.T) {
		tester(t, "foo bar baz", 1, []string{"f", "o", "o b", "a", "r b", "a", "z"})
	})

	// (.venv)
	// ltheisen@ltserver ~/git/github.com/meta-llama/llama-models
	// $ python3 -c "import models.llama3.api; print([i for i in models.llama3.api.Tokenizer('models/llama3/api/tokenizer.model')._split_whitespaces_or_nonwhitespaces('foo bar baz', 2)]);"
	// ['fo', 'o ba', 'r ba', 'z']
	t.Run("split every 2", func(t *testing.T) {
		tester(t, "foo bar baz", 2, []string{"fo", "o ba", "r ba", "z"})
	})

	// (.venv)
	// ltheisen@ltserver ~/git/github.com/meta-llama/llama-models
	// $ python3 -c "import models.llama3.api; print([i for i in models.llama3.api.Tokenizer('models/llama3/api/tokenizer.model')._split_whitespaces_or_nonwhitespaces('foo bar baz hiphophapzipzopzap', 2)]);"
	// ['fo', 'o ba', 'r ba', 'z hi', 'ph', 'op', 'ha', 'pz', 'ip', 'zo', 'pz', 'ap']
	t.Run("split every 2 with long sequence", func(t *testing.T) {
		tester(
			t,
			"foo bar baz hiphophapzipzopzap",
			2,
			[]string{"fo", "o ba", "r ba", "z hi", "ph", "op", "ha", "pz", "ip", "zo", "pz", "ap"})
	})

	// (.venv)
	// ltheisen@ltserver ~/git/github.com/meta-llama/llama-models
	// $ python3 -c "import models.llama3.api; print([i for i in models.llama3.api.Tokenizer('models/llama3/api/tokenizer.model')._split_whitespaces_or_nonwhitespaces('foo bar baz hiphophapzipzopzap', 3)]);"
	// ['foo bar baz hip', 'hop', 'hap', 'zip', 'zop', 'zap']
	t.Run("split every 3 with long sequence", func(t *testing.T) {
		tester(
			t,
			"foo bar baz hiphophapzipzopzap",
			3,
			[]string{"foo bar baz hip", "hop", "hap", "zip", "zop", "zap"})
	})
}
