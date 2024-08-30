// Package server provides a server for the LSP protocol implementation for the
// embedpls language server.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/conneroisu/embedpls/internal/lsp"
	"github.com/conneroisu/embedpls/internal/lsp/methods"
	"github.com/conneroisu/embedpls/internal/rpc"
	"github.com/conneroisu/embedpls/internal/safe"
	"go.lsp.dev/uri"
)

// Handler is an interface for handling messages from the client to the server.
type Handler interface {
	Handle(
		ctx context.Context,
		msg *rpc.BaseMessage,
	) (rpc.MethodActor, error)
}

// NewLSPHandler creates a new LSPHandler.
func NewLSPHandler(documents *safe.Map[uri.URI, string]) Handler {
	return &lspHandler{documents: documents}
}

type lspHandler struct {
	documents *safe.Map[uri.URI, string]
	cancelMap *safe.Map[int, context.CancelFunc]
}

// Handle handles a message from the client to the server.
func (l *lspHandler) Handle(
	ctx context.Context,
	msg *rpc.BaseMessage,
) (rpc.MethodActor, error) {
	errCh := make(chan error)
	resultCh := make(chan rpc.MethodActor)
	go func() {
		result, err := l.handle(ctx, msg)
		if err == nil {
			resultCh <- result
			return
		}
		errCh <- err
	}()
	select {
	case err := <-errCh:
		return nil, err
	case result := <-resultCh:
		return result, nil
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}
}

func decode[
	T lsp.InitializeRequest | lsp.NotificationDidOpenTextDocument | lsp.TextDocumentCompletionRequest | lsp.HoverRequest | lsp.TextDocumentCodeActionRequest | lsp.ShutdownRequest | lsp.CancelRequest | lsp.DidSaveTextDocumentNotification | lsp.DidCloseTextDocumentParamsNotification | lsp.TextDocumentDidChangeNotification,
](msg *rpc.BaseMessage) (T, error) {
	var request T
	err := json.Unmarshal([]byte(msg.Content), &request)
	if err != nil {
		return request, fmt.Errorf("decode (%s) failed: %w", msg.Method, err)
	}
	return request, nil
}

func (l *lspHandler) handle(ctx context.Context, msg *rpc.BaseMessage) (rpc.MethodActor, error) {
	switch methods.Method(msg.Method) {
	case methods.MethodCancelRequest:
		request, err := decode[lsp.CancelRequest](msg)
		if err != nil {
			return nil, err
		}
		id, err := lsp.ParseCancelParams(request.Params)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to parse cancel params: %w",
				err,
			)
		}
		c, ok := l.cancelMap.Get(int(request.Params.ID.(float64)))
		if ok {
			(*c)()
		}
		return lsp.CancelResponse{
			RPC: lsp.RPCVersion,
			ID:  id,
		}, nil

	case methods.MethodNotificationExit:
		for _, cancel := range l.cancelMap.Values() {
			cancel()
		}
		os.Exit(0)
		return nil, nil

	case methods.NotificationTextDocumentDidClose:
		var request lsp.DidCloseTextDocumentParamsNotification
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (%s) failed: %w",
				msg.Method,
				err,
			)
		}
		l.documents.Delete(request.Params.TextDocument.URI)
		return nil, nil

	case methods.MethodNotificationInitialized:
		return nil, nil

	case methods.MethodNotificationTextDocumentWillSave:
		return nil, nil

	case methods.MethodNotificationTextDocumentDidSave:
		request, err := decode[lsp.DidSaveTextDocumentNotification](msg)
		if err != nil {
			return nil, err
		}
		read, err := os.ReadFile(request.Params.TextDocument.URI.Filename())
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		l.documents.Set(request.Params.TextDocument.URI, string(read))
		return nil, nil

	case methods.MethodShutdown:
		request, err := decode[lsp.ShutdownRequest](msg)
		if err != nil {
			return nil, err
		}
		for _, cancel := range l.cancelMap.Values() {
			cancel()
		}
		return lsp.NewShutdownResponse(request, nil)

	case methods.NotificationMethodTextDocumentDidChange:
		var request lsp.TextDocumentDidChangeNotification
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (%s) failed: %w",
				msg.Method,
				err,
			)
		}
		l.documents.Set(request.Params.TextDocument.URI, string(request.Params.ContentChanges[0].Text))
		return nil, nil

	case methods.MethodInitialize:
		request, err := decode[lsp.InitializeRequest](msg)
		if err != nil {
			return nil, err
		}
		return lsp.NewInitializeResponse(&request), nil

	case methods.MethodRequestTextDocumentDidOpen:
		request, err := decode[lsp.NotificationDidOpenTextDocument](msg)
		if err != nil {
			return nil, err
		}
		return l.handleOpenDocument(
			ctx,
			&request,
		)

	case methods.MethodRequestTextDocumentDefinition:
		request, err := decode[lsp.TextDocumentCompletionRequest](msg)
		if err != nil {
			return nil, err
		}
		return l.handleTextDocumentDefinition(
			ctx,
			request,
		)

	case methods.MethodRequestTextDocumentCompletion:
		request, err := decode[lsp.TextDocumentCompletionRequest](msg)
		if err != nil {
			return nil, err
		}
		return l.handleTextDocumentCompletion(
			ctx,
			request,
		)

	case methods.MethodRequestTextDocumentHover:
		request, err := decode[lsp.HoverRequest](msg)
		if err != nil {
			return nil, err
		}
		return l.handleTextDocumentHover(
			ctx,
			request,
		)

	case methods.MethodRequestTextDocumentCodeAction:
		request, err := decode[lsp.TextDocumentCodeActionRequest](msg)
		if err != nil {
			return nil, err
		}
		return l.handleTextDocumentCodeAction(
			ctx,
			request,
		)

	default:
		return nil, fmt.Errorf("unknown method: %s", msg.Method)
	}
}

//

func (l *lspHandler) handleOpenDocument(
	_ context.Context,
	request *lsp.NotificationDidOpenTextDocument,
) (rpc.MethodActor, error) {
	if !strings.HasSuffix(
		string(request.Params.TextDocument.URI),
		".go",
	) {
		return nil, nil
	}
	l.documents.Set(request.Params.TextDocument.URI, string(request.Params.TextDocument.Text))
	return nil, nil
}

// TODO: Implement Below This Line

func (l *lspHandler) handleTextDocumentCompletion(
	ctx context.Context,
	request lsp.TextDocumentCompletionRequest,
) (rpc.MethodActor, error) {
	return nil, nil
}

//

func (l *lspHandler) handleTextDocumentHover(
	ctx context.Context,
	request lsp.HoverRequest,
) (rpc.MethodActor, error) {
	return &lsp.HoverResponse{
		Response: lsp.Response{
			RPC: lsp.RPCVersion,
			ID:  request.ID,
		},
		Result: lsp.HoverResult{
			Contents: "Hello, world!",
		},
	}, nil
}

//

func (l *lspHandler) handleTextDocumentDefinition(
	ctx context.Context,
	request lsp.TextDocumentCompletionRequest,
) (rpc.MethodActor, error) {
	return nil, nil
}

//

func (l *lspHandler) handleTextDocumentCodeAction(
	ctx context.Context,
	request lsp.TextDocumentCodeActionRequest,
) (rpc.MethodActor, error) {
	return nil, nil
}
