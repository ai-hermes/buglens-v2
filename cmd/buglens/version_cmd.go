package main

import (
	"fmt"

	"github.com/ai-hermes/buglens-v2/internal/version"

	"github.com/spf13/cobra"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "print version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("buglens %s\\n", version.String())
		},
	}
}
