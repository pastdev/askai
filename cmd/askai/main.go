package main

import (
	"os"

	"github.com/pastdev/askai/cmd/askai/root"
)

func main() {
	if err := root.New().Execute(); err != nil {
		os.Exit(1)
	}
}
