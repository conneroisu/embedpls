package server

import (
	"context"

	"github.com/conneroisu/embedpls/internal/rpc"
)

// HandleMessage handles a message from the client to the server.
func HandleMessage(
	ctx context.Context,
	msg *rpc.BaseMessage,
	writer *rpc.Writer,
) error {
	return nil
}
