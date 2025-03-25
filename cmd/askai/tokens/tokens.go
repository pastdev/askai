package tokens

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/pastdev/askai/pkg/tokenizer"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	cmd := cobra.Command{
		Use:   "tokens",
		Short: `Encode and decode text to/from BPE`,
	}

	cmd.AddCommand(NewDecode())
	cmd.AddCommand(NewEncode())

	return &cmd
}

func NewDecode() *cobra.Command {
	var model string

	cmd := cobra.Command{
		Use:   "decode",
		Short: `Decode text from BPE`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			tkzr, err := tokenizer.NewTokenizer(model)
			if err != nil {
				return fmt.Errorf("new tokenizer: %w", err)
			}

			encoded := make([]int, 0)
			if len(args) == 0 {
				err = json.NewDecoder(os.Stdin).Decode(&encoded)
			} else {
				err = json.Unmarshal([]byte(args[0]), &encoded)
			}
			if err != nil {
				return fmt.Errorf("json unmarshal: %w", err)
			}

			fmt.Print(tkzr.Decode(encoded))

			return nil
		},
	}

	cmd.Flags().StringVar(
		&model,
		"model",
		"gpt-4",
		"the model to use")

	return &cmd
}

func NewEncode() *cobra.Command {
	var model string

	cmd := cobra.Command{
		Use:   "encode",
		Short: `Encode text to BPE`,
		//nolint: revive // required to match upstream signature
		RunE: func(cmd *cobra.Command, args []string) error {
			tkzr, err := tokenizer.NewTokenizer(model)
			if err != nil {
				return fmt.Errorf("new tokenizer: %w", err)
			}

			var text string
			if len(args) == 0 {
				data, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("read stdin: %w", err)
				}
				text = string(data)
			} else {
				text = args[0]
			}

			encoded := tkzr.Encode(text, nil, nil)

			err = json.NewEncoder(os.Stdout).Encode(encoded)
			if err != nil {
				return fmt.Errorf("json encode: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(
		&model,
		"model",
		"gpt-4",
		"the model to use")

	return &cmd
}
