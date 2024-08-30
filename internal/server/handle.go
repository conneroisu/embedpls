// Package server provides a server for the LSP protocol implementation for the
// embedpls language server.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/charmbracelet/log"
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

func (l *lspHandler) handle(ctx context.Context, msg *rpc.BaseMessage) (rpc.MethodActor, error) {
	switch methods.Method(msg.Method) {
	case methods.MethodInitialize:
		var request lsp.InitializeRequest
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf("decode initialize request (initialize) failed: %w", err)
		}
		return l.handleInitialize(ctx, request)
	case methods.MethodRequestTextDocumentDidOpen:
		var request lsp.NotificationDidOpenTextDocument
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (textDocument/didOpen) request failed: %w",
				err,
			)
		}
		return l.handleOpenDocument(
			ctx,
			&request,
		)

	case methods.MethodRequestTextDocumentDefinition:
		var request lsp.TextDocumentCompletionRequest
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal definition request (textDocument/definition): %w",
				err,
			)
		}
		return l.handleTextDocumentDefinition(
			ctx,
			request,
		)
	case methods.MethodRequestTextDocumentCompletion:
		var request lsp.TextDocumentCompletionRequest
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal completion request (textDocument/completion): %w",
				err,
			)
		}
		return l.handleTextDocumentCompletion(
			ctx,
			request,
		)

	case methods.MethodRequestTextDocumentHover:
		var request lsp.HoverRequest
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed unmarshal of hover request (): %w",
				err,
			)
		}
		return l.handleTextDocumentHover(
			ctx,
			request,
		)

	case methods.MethodRequestTextDocumentCodeAction:
		var request lsp.TextDocumentCodeActionRequest
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal of codeAction request (textDocument/codeAction): %w",
				err,
			)
		}
		return l.handleTextDocumentCodeAction(
			ctx,
			request,
		)

	case methods.MethodShutdown:
		var request lsp.ShutdownRequest
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (shutdown) request failed: %w",
				err,
			)
		}
		for _, cancel := range l.cancelMap.Values() {
			cancel()
		}
		return lsp.NewShutdownResponse(request, nil)

	case methods.MethodCancelRequest:
		var request lsp.CancelRequest
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to unmarshal cancel request ($/cancelRequest): %w",
				err,
			)
		}
		id, err := lsp.ParseCancelParams(request.Params)
		if err != nil {
			return nil, fmt.Errorf(
				"failed to parse cancel params: %w",
				err,
			)
		}
		log.Debugf("canceling request: %d", request.Params.ID)
		c, ok := l.cancelMap.Get(int(request.Params.ID.(float64)))
		if ok {
			(*c)()
		}
		return lsp.CancelResponse{
			RPC: lsp.RPCVersion,
			ID:  id,
		}, nil

	case methods.MethodNotificationInitialized:
		var request lsp.InitializedParamsRequest
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (initialized) request failed: %w",
				err,
			)
		}
		return nil, nil

	case methods.MethodNotificationExit:
		for _, cancel := range l.cancelMap.Values() {
			cancel()
		}
		os.Exit(0)
		return nil, nil

	case methods.MethodNotificationTextDocumentWillSave:
		return nil, nil

	case methods.MethodNotificationTextDocumentDidSave:
		var request lsp.DidSaveTextDocumentNotification
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (didSave) request failed: %w",
				err,
			)
		}
		read, err := os.ReadFile(request.Params.TextDocument.URI.Filename())
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}
		l.documents.Set(request.Params.TextDocument.URI, string(read))
		return nil, nil

	case methods.NotificationTextDocumentDidClose:
		var request lsp.DidCloseTextDocumentParamsNotification
		err := json.Unmarshal([]byte(msg.Content), &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (didClose) request failed: %w",
				err,
			)
		}
		return l.handleTextDocumentDidClose(ctx, request)

	case methods.NotificationMethodTextDocumentDidChange:
		var request lsp.TextDocumentDidChangeNotification
		err := json.Unmarshal(msg.Content, &request)
		if err != nil {
			return nil, fmt.Errorf(
				"decode (textDocument/didChange) request failed: %w",
				err,
			)
		}
		return l.handleTextDocumentDidChange(
			ctx,
			request,
		)

	default:
		return nil, fmt.Errorf("unknown method: %s", msg.Method)
	}
}

//

func (l *lspHandler) handleInitialize(
	_ context.Context,
	request lsp.InitializeRequest,
) (rpc.MethodActor, error) {
	return lsp.NewInitializeResponse(&request), nil
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

//

func (l *lspHandler) handleTextDocumentDidChange(
	_ context.Context,
	request lsp.TextDocumentDidChangeNotification,
) (rpc.MethodActor, error) {
	l.documents.Set(request.Params.TextDocument.URI, string(request.Params.ContentChanges[0].Text))
	return nil, nil
}

//

func (l *lspHandler) handleTextDocumentDidClose(
	_ context.Context,
	request lsp.DidCloseTextDocumentParamsNotification,
) (rpc.MethodActor, error) {
	l.documents.Delete(request.Params.TextDocument.URI)
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
