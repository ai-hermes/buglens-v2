package main

import "github.com/spf13/cobra"

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "buglens",
		Short: "buglens Go CLI",
	}

	root.AddCommand(newVersionCmd())
	root.AddCommand(newMCPCmd())
	return root
}
