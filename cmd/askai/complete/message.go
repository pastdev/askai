package complete

import (
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go/v3"
	oai "github.com/openai/openai-go/v3"
	"github.com/pastdev/askai/pkg/log"
	"github.com/spf13/pflag"
)

type messageArrayValue struct {
	msgs    *[]openai.ChatCompletionMessageParamUnion
	factory func(string) oai.ChatCompletionMessageParamUnion
}

func newMessageArrayValue(
	val []openai.ChatCompletionMessageParamUnion,
	p *[]openai.ChatCompletionMessageParamUnion,
	factory func(string) oai.ChatCompletionMessageParamUnion,
) *messageArrayValue {
	mav := new(messageArrayValue)
	mav.msgs = p
	*mav.msgs = val
	mav.factory = factory
	return mav
}

func (m *messageArrayValue) String() string {
	// error ignored in upstream StringArrayVar as well
	msgs, _ := json.Marshal(m.msgs)
	return string(msgs)
}

func (m *messageArrayValue) Set(v string) error {
	var msg openai.ChatCompletionMessageParamUnion
	if m.factory == nil {
		err := json.Unmarshal([]byte(v), &msg)
		if err != nil {
			return fmt.Errorf("unmarshal message: %w", err)
		}
	} else {
		msg = m.factory(v)
	}

	if len(*m.msgs) > 0 {
		log.Trace().Interface("message", msg).Msg("adding message")
		*m.msgs = append(*m.msgs, msg)
	} else {
		log.Trace().Interface("message", msg).Msg("initial message")
		*m.msgs = []openai.ChatCompletionMessageParamUnion{msg}
	}

	return nil
}

func (*messageArrayValue) Type() string {
	return "messages"
}

func MessageArrayVar(
	f *pflag.FlagSet,
	factory func(string) oai.ChatCompletionMessageParamUnion,
	p *[]openai.ChatCompletionMessageParamUnion,
	name string,
	value []openai.ChatCompletionMessageParamUnion,
	usage string,
) {
	MessageArrayVarP(f, factory, p, name, "", value, usage)
}

func MessageArrayVarP(
	f *pflag.FlagSet,
	factory func(string) oai.ChatCompletionMessageParamUnion,
	p *[]oai.ChatCompletionMessageParamUnion,
	name string,
	shorthand string,
	value []oai.ChatCompletionMessageParamUnion,
	usage string,
) {
	f.VarP(newMessageArrayValue(value, p, factory), name, shorthand, usage)
}
