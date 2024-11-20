package chatcompletion

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sashabaranov/go-openai"
)

type PersistentConversation struct {
	name    string
	request openai.ChatCompletionRequest
}

// LoadPersistentConversation will load an existing conversation by the supplied
// name or create it if it does not exist. If a new one is created, the returned
// bool will be true.
func LoadPersistentConversation(
	name string,
	defaults openai.ChatCompletionRequest,
) (PersistentConversation, bool, error) {
	c := PersistentConversation{name: name}

	err := deepCopy(&c.request, &defaults)
	if err != nil {
		return c, false, fmt.Errorf("deep copy defaults: %w", err)
	}

	yml, err := os.ReadFile(conversationFile(name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return c, true, nil
		}

		return c, false, fmt.Errorf("read %s: %w", name, err)
	}

	err = json.Unmarshal(yml, &c.request)
	if err != nil {
		return c, false, fmt.Errorf("unmarshal %s: %w", c.name, err)
	}

	return c, false, nil
}

func (c *PersistentConversation) SetModel(model string) {
	c.request.Model = model
}

func (c *PersistentConversation) Continue(
	msgs ...openai.ChatCompletionMessage,
) openai.ChatCompletionRequest {
	c.request.Messages = append(c.request.Messages, msgs...)
	return c.request
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
func deepCopy(dest any, src any) error {
	data, err := json.Marshal(src)
	if err != nil {
		return fmt.Errorf("deepCopy marshal: %w", err)
	}

	err = json.Unmarshal(data, dest)
	if err != nil {
		return fmt.Errorf("deepCopy marshal: %w", err)
	}

	return nil
}
