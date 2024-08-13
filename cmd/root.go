package main

import "github.com/spf13/cobra"

// main is the entry point for the application.
func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}

// run is the main function for the application.
func run() error {
	rootCmd := NewRootCmd()
	err := rootCmd.Execute()
	if err != nil {
		return err
	}
	return nil
}

// NewRootCmd creates a new root command.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "embedpls",
		Short: "EmbedPLS is a tool for embedding PL/SQL code into a Go application",
	}
	rootCmd.AddCommand(NewLspCmd())
	rootCmd.AddCommand(NewVersionCmd())
	return rootCmd
}
