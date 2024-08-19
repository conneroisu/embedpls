package main

import (
	"context"

	"github.com/conneroisu/embedpls/internal/rpc"
	"github.com/spf13/cobra"
)

// Handler is an interface for handling messages from the client to the server.
type Handler interface {
	Handle(ctx context.Context, msg *rpc.BaseMessage, writer *rpc.Writer) error
}

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
		Short: "EmbedPLS is cli based langauage server for the go std-lib embed package.",
	}
	rootCmd.AddCommand(NewLspCmd())
	rootCmd.AddCommand(NewVersionCmd())
	return rootCmd
}
