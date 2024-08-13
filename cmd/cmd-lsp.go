package main

import "github.com/spf13/cobra"

// NewLspCmd creates a new lsp command.
func NewLspCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lsp",
		Short: "Starts the LSP server",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement
		},
	}
}
