package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"reflect"

	"github.com/charmbracelet/log"
	"github.com/conneroisu/embedpls/internal/rpc"
	"github.com/conneroisu/embedpls/internal/safe"
	"github.com/conneroisu/embedpls/internal/server"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"go.lsp.dev/uri"
)

// NewLspCmd creates a new lsp command.
func NewLspCmd(
	reader io.Reader,
	writer io.Writer,
	handle func(documents *safe.Map[uri.URI, string]) server.Handler,
) *cobra.Command {
	cmd := cobra.Command{
		Use:   "lsp",
		Short: "Starts the LSP server.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			configPath, err := CreateConfigDir("~/.config/embedpls/")
			if err != nil {
				return fmt.Errorf("failed to create config directory: %w", err)
			}
			logPath := path.Join(configPath, "state.log")
			f, err := os.OpenFile(
				logPath,
				os.O_CREATE|os.O_APPEND|os.O_WRONLY,
				0666,
			)
			if err != nil {
				return fmt.Errorf("failed to create log file: %w", err)
			}
			log.SetOutput(f)
			log.SetLevel(log.DebugLevel)
			scanner := bufio.NewScanner(reader)
			rpcWriter := rpc.NewWriter(writer)
			innerCtx, cancel := context.WithCancel(cmd.Context())
			documents := safe.NewSafeMap[uri.URI, string]()
			handler := handle(documents)
			defer cancel()
			scanner.Split(rpc.Split)
			for scanner.Scan() {
				decoded, err := rpc.DecodeMessage(scanner.Bytes())
				if err != nil {
					return err
				}
				resp, err := handler.Handle(
					innerCtx,
					decoded,
				)
				if err != nil {
					log.Errorf(
						"failed to handle message: %s",
						err,
					)
					continue
				}
				if !isNull(resp) {
					err = rpcWriter.WriteResponse(innerCtx, resp)
					if err != nil {
						log.Errorf(
							"failed to write (%s) response: %s",
							resp.Method(),
							err,
						)
					}
				}
			}
			return nil
		},
	}
	return &cmd
}

// isNull checks if the given interface is nil or points to a nil value
func isNull(i interface{}) bool {
	if i == nil {
		return true
	}
	// Use reflect.ValueOf only if the kind is valid for checking nil
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Chan,
		reflect.Func,
		reflect.Interface,
		reflect.Map,
		reflect.Ptr,
		reflect.Slice:
		return v.IsNil()
	}
	return false
}

// CreateConfigDir creates a new config directory and returns the path.
func CreateConfigDir(dirPath string) (string, error) {
	path, err := homedir.Expand(dirPath)
	if err != nil {
		return "", fmt.Errorf("failed to expand home directory: %w", err)
	}
	if err := os.MkdirAll(path, 0755); err != nil {
		if os.IsExist(err) {
			return path, nil
		}
		return "", fmt.Errorf(
			"failed to create or find config directory: %w",
			err,
		)
	}
	return path, nil
}
