package version

import (
	"fmt"

	"github.com/pastdev/askai/pkg/version"
	"github.com/spf13/cobra"
)

func New() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Display askai version",
		//nolint: revive // command part of the implementation signature
		Run: func(command *cobra.Command, args []string) {
			fmt.Printf("Version: %s\n", version.Version())
		},
	}
}
