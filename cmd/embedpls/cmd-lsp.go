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
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"go.lsp.dev/uri"
)

// LSPHandler is a struct for the LSP server
type LSPHandler func(
	ctx context.Context,
	cancel *context.CancelFunc, // cancel is a pointer to the cancel function to avoid copying
	msg *rpc.BaseMessage, // msg is a pointer to the message to avoid copying
	documents *safe.Map[uri.URI, string],
) (rpc.MethodActor, error)

// NewLspCmd creates a new lsp command.
func NewLspCmd(
	ctx context.Context,
	reader io.Reader,
	writer io.Writer,
	handle LSPHandler,
) *cobra.Command {
	cmd := cobra.Command{
		Use:   "lsp",
		Short: "Starts the LSP server.",
		RunE: func(cmd *cobra.Command, args []string) error {
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
			documents := safe.NewSafeMap[uri.URI, string]()
			scanner := bufio.NewScanner(reader)
			scanner.Split(rpc.Split)
			rpcWriter := rpc.NewWriter(writer)
			innerCtx, cancel := context.WithCancel(cmd.Context())
			for scanner.Scan() {
				decoded, err := rpc.DecodeMessage(scanner.Bytes())
				if err != nil {
					return err
				}
				resp, err := handle(
					innerCtx,
					&cancel,
					decoded,
					documents,
				)
				if err != nil {
					log.Errorf(
						"failed to decode message: %s",
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
	cmd.SetContext(ctx)
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
