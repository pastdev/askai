package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var version = "0.0.0"

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display askai version",
		Run: func(command *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version)
		},
	}
}
