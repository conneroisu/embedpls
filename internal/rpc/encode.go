package rpc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/charmbracelet/log"
	"github.com/conneroisu/embedpls/internal/lsp"
	"github.com/conneroisu/embedpls/internal/lsp/methods"
)

// MethodActor is a type for responding to a request
type MethodActor interface {
	Method() methods.Method
}

// Encode encodes a message into a string
//
// It uses the json library to encode the message
// and returns a string representation of the encoded message with
// a Content-Length header.
//
// It also returns an error if there is an error while encoding the message.
func Encode(
	ctx context.Context,
	msg MethodActor,
) (string, error) {
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		buffer := &bytes.Buffer{}
		encoder := json.NewEncoder(buffer)
		encoder.SetEscapeHTML(false)
		err := encoder.Encode(msg)
		if err != nil {
			return "", err
		}
		body := buffer.Bytes()
		// body := bytes.TrimSuffix(
		//         buffer.Bytes(),
		//         []byte("\n"),
		// )
		log.Debugf(
			"wrote msg [%d] (%s): %s",
			len(body),
			msg.Method(),
			string(body),
		)
		result := fmt.Sprintf(
			"Content-Length: %d\r\n\r\n%s",
			len(body),
			string(body),
		)
		return result, nil
	}
}

// Decode decodes a message into lsp request.
func Decode[
	T lsp.InitializeRequest | lsp.NotificationDidOpenTextDocument | lsp.TextDocumentCompletionRequest | lsp.HoverRequest | lsp.TextDocumentCodeActionRequest | lsp.ShutdownRequest | lsp.CancelRequest | lsp.DidSaveTextDocumentNotification | lsp.DidCloseTextDocumentParamsNotification | lsp.TextDocumentDidChangeNotification,
](msg *BaseMessage) (T, error) {
	var request T
	err := json.Unmarshal([]byte(msg.Content), &request)
	if err != nil {
		return request, fmt.Errorf("decode (%s) failed: %w", msg.Method, err)
	}
	return request, nil
}
