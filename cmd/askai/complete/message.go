package complete

import (
	"encoding/json"
	"fmt"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
	"github.com/spf13/pflag"
)

type messageArrayValue struct {
	msgs *[]openai.ChatCompletionMessage
	role string
}

func newMessageArrayValue(
	val []openai.ChatCompletionMessage,
	p *[]openai.ChatCompletionMessage,
	role string,
) *messageArrayValue {
	mav := new(messageArrayValue)
	mav.msgs = p
	*mav.msgs = val
	mav.role = role
	return mav
}

func (m *messageArrayValue) String() string {
	// error ignored in upstream StringArrayVar as well
	msgs, _ := json.Marshal(m.msgs)
	return string(msgs)
}

func (m *messageArrayValue) Set(v string) error {
	var msg openai.ChatCompletionMessage
	if m.role == "" {
		err := json.Unmarshal([]byte(v), &msg)
		if err != nil {
			return fmt.Errorf("unmarshal message: %w", err)
		}
	} else {
		msg = openai.ChatCompletionMessage{
			Role:    m.role,
			Content: v,
		}
	}

	if len(*m.msgs) > 0 {
		log.Trace().Interface("message", msg).Msg("adding message")
		*m.msgs = append(*m.msgs, msg)
	} else {
		log.Trace().Interface("message", msg).Msg("initial message")
		*m.msgs = []openai.ChatCompletionMessage{msg}
	}

	return nil
}

func (*messageArrayValue) Type() string {
	return "messages"
}

func MessageArrayVar(
	f *pflag.FlagSet,
	role string,
	p *[]openai.ChatCompletionMessage,
	name string,
	value []openai.ChatCompletionMessage,
	usage string,
) {
	MessageArrayVarP(f, role, p, name, "", value, usage)
}

func MessageArrayVarP(
	f *pflag.FlagSet,
	role string,
	p *[]openai.ChatCompletionMessage,
	name string,
	shorthand string,
	value []openai.ChatCompletionMessage,
	usage string,
) {
	f.VarP(newMessageArrayValue(value, p, role), name, shorthand, usage)
}
