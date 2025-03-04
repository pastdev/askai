package chatcompletion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pastdev/askai/pkg/log"
	"github.com/sashabaranov/go-openai"
)

type PersistentConversation struct {
	name    string
	request openai.ChatCompletionRequest
}

// LoadPersistentConversation will load an existing conversation by the supplied
// name or create it if it does not exist.
func LoadPersistentConversation(
	name string,
	defaults openai.ChatCompletionRequest,
) (PersistentConversation, error) {
	c := PersistentConversation{name: name}

	err := deepCopy(&c.request, &defaults)
	if err != nil {
		return c, fmt.Errorf("deep copy defaults: %w", err)
	}
	log.Trace().Interface("request", c.request).Msg("request after defaults")

	yml, err := os.ReadFile(conversationFile(name))
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

func (c *PersistentConversation) Continue(
	reply openai.ChatCompletionRequest,
) (openai.ChatCompletionRequest, error) {
	//nolint: gocritic // assigned to same slice after deep copy
	messages := append(c.request.Messages, reply.Messages...)
	log.Trace().Interface("messages", messages).Msg("after continue append")
	err := deepCopy(&c.request, &reply)
	if err != nil {
		return openai.ChatCompletionRequest{}, fmt.Errorf("deep copy reply: %w", err)
	}
	log.Trace().Interface("messages", messages).Msg("BEFORE")
	c.request.Messages = messages
	log.Trace().Interface("messages", messages).Msg("AFTER")
	log.Trace().Interface("request", c.request).Msg("continue request")
	return c.request, nil
}

func (c PersistentConversation) UpdateResponse(response string) error {
	c.request.Messages = append(
		c.request.Messages,
		openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: response,
		})

	data, err := json.Marshal(c.request)
	if err != nil {
		return fmt.Errorf("marshal %s: %w", c.name, err)
	}

	err = os.MkdirAll(conversationDir(), 0700)
	if err != nil {
		return fmt.Errorf("mkdir %s: %w", conversationDir(), err)
	}

	err = os.WriteFile(conversationFile(c.name), data, 0600)
	if err != nil {
		return fmt.Errorf("write response: %w", err)
	}

	return nil
}

func conversationDir() string {
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

func conversationFile(name string) string {
	return filepath.Join(conversationDir(), name)
}

// deepCopy will copy all public fields from src into dest recursively
func deepCopy(dest *openai.ChatCompletionRequest, src *openai.ChatCompletionRequest) error {
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
