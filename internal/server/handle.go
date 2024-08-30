// Package server provides a server for the LSP protocol implementation for the
// embedpls language server.
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/conneroisu/embedpls/internal/lsp"
	"github.com/conneroisu/embedpls/internal/lsp/methods"
	"github.com/conneroisu/embedpls/internal/parsers"
	"github.com/conneroisu/embedpls/internal/rpc"
	"github.com/conneroisu/embedpls/internal/safe"
	"go.lsp.dev/protocol"
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
	ctx, cancel := context.WithTimeout(ctx, time.Second*1)
	defer cancel()
	go func() {
		result, err := l.handle(ctx, msg)
		if err == nil {
			resultCh <- result
			return
		}
		errCh <- err
	}()
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	case err := <-errCh:
		return nil, err
	case result := <-resultCh:
		return result, nil
	}
}

func (l *lspHandler) handle(ctx context.Context, msg *rpc.BaseMessage) (rpc.MethodActor, error) {
	switch methods.Method(msg.Method) {
	case methods.MethodCancelRequest:
		request, err := rpc.Decode[lsp.CancelRequest](msg)
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
		c, ok := l.cancelMap.Get(id)
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
		request, err := rpc.Decode[lsp.DidSaveTextDocumentNotification](msg)
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
		request, err := rpc.Decode[lsp.ShutdownRequest](msg)
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
		request, err := rpc.Decode[lsp.InitializeRequest](msg)
		if err != nil {
			return nil, err
		}
		return lsp.NewInitializeResponse(&request), nil

	case methods.MethodRequestTextDocumentDidOpen:
		request, err := rpc.Decode[lsp.NotificationDidOpenTextDocument](msg)
		if err != nil {
			return nil, err
		}
		if !strings.HasSuffix(
			string(request.Params.TextDocument.URI),
			".go",
		) {
			return nil, nil
		}
		l.documents.Set(request.Params.TextDocument.URI, string(request.Params.TextDocument.Text))
		return nil, nil

	case methods.MethodRequestTextDocumentDefinition:
		request, err := rpc.Decode[lsp.TextDocumentCompletionRequest](msg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*1)
		defer cancel()
		ans, err := l.handleTextDocumentDefinition(
			ctx,
			request,
		)
		return ans, err

	case methods.MethodRequestTextDocumentCompletion:
		request, err := rpc.Decode[lsp.TextDocumentCompletionRequest](msg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*1)
		defer cancel()
		ans, err := l.handleTextDocumentCompletion(
			ctx,
			request,
		)
		return ans, err

	case methods.MethodRequestTextDocumentHover:
		request, err := rpc.Decode[lsp.HoverRequest](msg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*1)
		defer cancel()
		ans, err := l.handleTextDocumentHover(
			ctx,
			request,
		)
		return ans, err

	case methods.MethodRequestTextDocumentCodeAction:
		request, err := rpc.Decode[lsp.TextDocumentCodeActionRequest](msg)
		if err != nil {
			return nil, err
		}
		ctx, cancel := context.WithTimeout(ctx, time.Second*1)
		defer cancel()
		ans, err := l.handleTextDocumentCodeAction(
			ctx,
			request,
		)
		return ans, err

	default:
		return nil, fmt.Errorf("unknown method: %s", msg.Method)
	}
}

// TODO: Implement Below This Line

func (l *lspHandler) handleTextDocumentCompletion(
	ctx context.Context,
	request lsp.TextDocumentCompletionRequest,
) (rpc.MethodActor, error) {
	doc, ok := l.documents.Get(request.Params.TextDocument.URI)
	if !ok {
		return nil, fmt.Errorf("document not found")
	}
	curVal, state, err := parsers.ParseSourcePosition(
		doc,
		request.Params.Position,
	)
	if err != nil {
		return nil, err
	}
	if state == parsers.StateUnknown {
		log.Debugf("unknown state")
		return nil, nil
	}
	errCh := make(chan error)
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	case embeds := <-getEmbbeddables(request.Params.TextDocument.URI, curVal, errCh):
		resp := &lsp.TextDocumentCompletionResponse{
			Response: lsp.Response{
				RPC: lsp.RPCVersion,
				ID:  request.ID,
			},
			Result: []protocol.CompletionItem{},
		}
		for _, embed := range embeds.embeddables {
			resp.Result = append(resp.Result, protocol.CompletionItem{
				Label:         embed.name,
				Detail:        embed.name,
				Documentation: embed.name,
				Kind:          protocol.CompletionItemKindFile,
			})
		}
	case err := <-errCh:
		return nil, err
	}
	return nil, nil
}

//

func (l *lspHandler) handleTextDocumentHover(
	ctx context.Context,
	request lsp.HoverRequest,
) (rpc.MethodActor, error) {
	resp := lsp.HoverResponse{
		Response: lsp.Response{
			RPC: lsp.RPCVersion,
			ID:  request.ID,
		}}
	errCh := make(chan error)
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	case res := <-l.getHoverResp(request, errCh):
		resp.Result = res
		return resp, nil
	case err := <-errCh:
		return nil, err
	}
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
