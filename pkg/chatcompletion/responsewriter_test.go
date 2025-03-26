package chatcompletion_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/pastdev/askai/pkg/chatcompletion"
	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/require"
)

func TestResponseWriter(t *testing.T) {
	tester := func(
		t *testing.T,
		wFac func(io.Writer) chatcompletion.ResponseWriter,
		res *openai.ChatCompletionResponse,
		sres []openai.ChatCompletionStreamResponse,
		validator func(*testing.T, string),
	) {
		var buf strings.Builder
		w := wFac(&buf)

		if res != nil {
			err := w.Write(*res)
			require.NoError(t, err)
		}

		for _, r := range sres {
			err := w.WriteStream(r)
			require.NoError(t, err)
		}

		validator(t, buf.String())
	}

	contentFac := func(w io.Writer) chatcompletion.ResponseWriter {
		return &chatcompletion.ContentResponseWriter{W: w}
	}

	contentValidator := func(expected string) func(*testing.T, string) {
		return func(t *testing.T, actual string) {
			require.Equal(t, expected, actual)
		}
	}

	rawFac := func(w io.Writer) chatcompletion.ResponseWriter {
		return &chatcompletion.RawResponseWriter{W: w}
	}

	rawValidator := func(expected string) func(*testing.T, string) {
		return func(t *testing.T, actual string) {
			expectedLines := strings.Split(strings.Trim(expected, "\n"), "\n")
			actualLines := strings.Split(strings.Trim(actual, "\n"), "\n")
			require.Equal(t, len(expectedLines), len(actualLines))
			for i := range expectedLines {
				require.JSONEq(t, expectedLines[i], actualLines[i])
			}
		}
	}

	resFac := func(t *testing.T, raw string) (*openai.ChatCompletionResponse, string) {
		var res openai.ChatCompletionResponse
		err := json.Unmarshal([]byte(raw), &res)
		require.NoError(t, err)
		fullRaw, err := json.Marshal(res)
		require.NoError(t, err)
		return &res, string(fullRaw)
	}

	resStreamFac := func(t *testing.T, raw string) ([]openai.ChatCompletionStreamResponse, string) {
		var res []openai.ChatCompletionStreamResponse
		err := json.Unmarshal([]byte(raw), &res)
		require.NoError(t, err)

		var builder strings.Builder
		enc := json.NewEncoder(&builder)
		for _, v := range res {
			err := enc.Encode(v)
			require.NoError(t, err)
		}
		return res, builder.String()
	}

	t.Run("empty choices", func(t *testing.T) {
		res, raw := resFac(
			t,
			`{
  "choices": []
}`)
		log.Trace().Interface("res", res).Str("raw", raw).Msg("resFac output")

		t.Run("raw", func(t *testing.T) {
			tester(
				t,
				rawFac,
				res,
				nil,
				rawValidator(raw))
		})

		t.Run("content", func(t *testing.T) {
			tester(
				t,
				contentFac,
				res,
				nil,
				contentValidator(""))
		})
	})

	t.Run("buffered", func(t *testing.T) {
		content := "Foo was a silent monk. He lived in a temple at the foot of a mountain. Foo spent his days meditating and tending to a garden. One day, a traveler arrived at the temple seeking wisdom. Foo served the traveler tea and they sat in silence. After a time, the traveler asked Foo a question. Foo responded with a gesture, pointing to a blooming flower. The traveler understood."
		res, raw := resFac(
			t,
			fmt.Sprintf(`{
  "choices": [
    {
      "content_filter_results": {},
      "finish_reason": "stop",
      "index": 0,
      "message": {
        "content": "%s",
        "role": "assistant"
      }
    }
  ],
  "created": 1743012340,
  "id": "chatcmpl-ec172374-b70a-4d2f-9445-dbf2dd29bdda",
  "model": "meta-llama/Llama-3.3-70B-Instruct",
  "object": "chat.completion",
  "system_fingerprint": "",
  "usage": {
    "completion_tokens": 83,
    "completion_tokens_details": null,
    "prompt_tokens": 58,
    "prompt_tokens_details": null,
    "total_tokens": 141
  }
}`,
				content))
		log.Trace().Interface("res", res).Str("raw", raw).Msg("resFac output")

		t.Run("raw", func(t *testing.T) {
			tester(
				t,
				rawFac,
				res,
				nil,
				rawValidator(raw))
		})

		t.Run("content", func(t *testing.T) {
			tester(
				t,
				contentFac,
				res,
				nil,
				contentValidator(content))
		})
	})

	t.Run("stream", func(t *testing.T) {
		content := []any{
			"B",
			"arn",
			"acle",
			" Black",
			"be",
			"ak",
			" Buster",
		}
		res, raw := resStreamFac(
			t,
			fmt.Sprintf(`[
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "role": "assistant"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {
          "content": "%s"
        },
        "finish_reason": null,
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  },
  {
    "choices": [
      {
        "content_filter_results": {},
        "delta": {},
        "finish_reason": "stop",
        "index": 0
      }
    ],
    "created": 1743014636,
    "id": "chatcmpl-30140901-b144-4a53-9914-c2558292d9e3",
    "model": "meta-llama/Llama-3.3-70B-Instruct",
    "object": "chat.completion.chunk",
    "system_fingerprint": ""
  }
]`,
				content...))
		log.Trace().Interface("res", res).Str("raw", raw).Msg("resFac output")

		t.Run("raw", func(t *testing.T) {
			tester(
				t,
				rawFac,
				nil,
				res,
				rawValidator(raw))
		})

		t.Run("content", func(t *testing.T) {
			tester(
				t,
				contentFac,
				nil,
				res,
				contentValidator(fmt.Sprintf(strings.Repeat("%s", len(content)), content...)))
		})
	})
}
