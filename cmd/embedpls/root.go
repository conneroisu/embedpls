package main

import (
	"context"
	"fmt"
	"os"

	"github.com/conneroisu/embedpls/internal/server"
	"github.com/spf13/cobra"
)

// main is the entry point for the application.
func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var (
	rootCmd = &cobra.Command{
		Use:   "embedpls",
		Short: "EmbedPLS is cli based langauage server for the go std-lib embed package.",
	}
)

func init() {
	rootCmd.AddCommand(NewLspCmd(
		os.Stdin,
		os.Stdout,
		server.NewLSPHandler,
	))
	rootCmd.AddCommand(NewVersionCmd())
}

// run is the main function for the application.
func run() error {
	rootCmd := NewRootCmd()
	err := rootCmd.ExecuteContext(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// NewRootCmd creates a new root command.
func NewRootCmd() *cobra.Command {
	return rootCmd
}
