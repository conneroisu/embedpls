package main

import "github.com/spf13/cobra"

// NewVersionCmd creates a new version command.
func NewVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Prints the version of the tool",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: Implement
		},
	}
}
