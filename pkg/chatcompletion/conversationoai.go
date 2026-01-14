package chatcompletion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	oai "github.com/openai/openai-go/v3"
	"github.com/pastdev/askai/pkg/log"
)

var _ ConversationOai = &PersistentConversationOai{}

type PersistentConversationOai struct {
	name    string
	request oai.ChatCompletionNewParams
}

// LoadPersistentConversation will load an existing conversation by the supplied
// name or create it if it does not exist.
func LoadPersistentConversationOai(
	name string,
	defaults oai.ChatCompletionNewParams,
) (PersistentConversationOai, error) {
	c := PersistentConversationOai{name: name}

	err := deepCopyOai(&c.request, &defaults)
	if err != nil {
		return c, fmt.Errorf("deep copy defaults: %w", err)
	}
	log.Trace().Interface("request", c.request).Msg("request after defaults")

	yml, err := os.ReadFile(conversationFileOai(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c, nil
		}

		return c, fmt.Errorf("read %s: %w", name, err)
	}

	err = json.Unmarshal(yml, &c.request)
	if err != nil {
		return c, fmt.Errorf("unmarshal %s: %w", c.name, err)
	}
	log.Trace().Interface("request", c.request).Msg("request after load")

	return c, nil
}

func (c *PersistentConversationOai) Continue(
	reply oai.ChatCompletionNewParams,
) (oai.ChatCompletionNewParams, error) {
	// originally:
	//   messages := append(c.request.Messages, reply.Messages...)
	// but append actually modifies and returns the first argument so it was
	// a reference to the c.request.Message that then got modified by the
	// deepCopy call causing the reply to replace the first message (system):
	//   https://github.com/pastdev/askai/issues/1
	// so we need to create a new array to avoid this
	messages := make(
		[]oai.ChatCompletionMessageParamUnion,
		0,
		len(c.request.Messages)+len(reply.Messages))
	messages = append(messages, c.request.Messages...)
	messages = append(messages, reply.Messages...)

	err := deepCopyOai(&c.request, &reply)
	if err != nil {
		return oai.ChatCompletionNewParams{}, fmt.Errorf("deep copy reply: %w", err)
	}
	c.request.Messages = messages
	return c.request, nil
}

func (c PersistentConversationOai) UpdateResponse(response string) error {
	c.request.Messages = append(
		c.request.Messages,
		oai.AssistantMessage(response))

	data, err := json.Marshal(c.request)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", c.name, err)
	}

	err = os.MkdirAll(conversationDir(), 0700)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", conversationDir(), err)
	}

	err = os.WriteFile(conversationFileOai(c.name), data, 0600)
	if err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func conversationDirOai() string {
	dir, ok := os.LookupEnv("XDG_DATA_HOME")
	if ok {
		return filepath.Join(dir, "askai")
	}

	dir, err := os.UserHomeDir()
	if err == nil {
		// default value of XDG_DATA_HOME:
		//   https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html#variables
		return filepath.Join(dir, ".local", "share", "askai")
	}

	return filepath.Join(os.TempDir(), "askai")
}

func conversationFileOai(name string) string {
	return filepath.Join(conversationDirOai(), name)
}

// deepCopy will copy all public fields from src into dest recursively
func deepCopyOai(dest *oai.ChatCompletionNewParams, src *oai.ChatCompletionNewParams) error {
	// Model is not _omitempty_ and we want to preserve the value from the existing
	// if it is not _explicitly changed_. So we store here and set it after if
	// needed.
	destModel := dest.Model

	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("deepCopy marshal: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("deepCopy marshal: %w", err)
	}

	if src.Model == "" {
		dest.Model = destModel
	}

	return nil
}
