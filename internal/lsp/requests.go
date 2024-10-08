package lsp

import (
	"github.com/conneroisu/embedpls/internal/lsp/methods"
	"go.lsp.dev/protocol"
)

// TextDocumentCompletionRequest is a request for a completion to the language server
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
type TextDocumentCompletionRequest struct {
	// CompletionRequest embeds the Request struct
	Request
	// Params are the parameters for the completion request
	Params protocol.CompletionParams `json:"params"`
}

// Method returns the method for the completion request
func (r TextDocumentCompletionRequest) Method() methods.Method {
	return methods.MethodRequestTextDocumentCompletion
}

// TextDocumentCompletionResponse is a response for a completion to the language server
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_completion
type TextDocumentCompletionResponse struct {
	// CompletionResponse embeds the Response struct
	Response
	// Result is the result of the completion request
	Result []protocol.CompletionItem `json:"result"`
}

// Method returns the method for the completion response
func (r TextDocumentCompletionResponse) Method() methods.Method {
	return methods.MethodRequestTextDocumentCompletion
}

// TextDocumentCodeActionRequest is a request for a code action to the language server.
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_codeAction
type TextDocumentCodeActionRequest struct {
	// CodeActionRequest embeds the Request struct
	Request
	// Params are the parameters for the code action request.
	Params protocol.CodeActionParams `json:"params"`
}

// Method returns the method for the code action request
func (r TextDocumentCodeActionRequest) Method() methods.Method {
	return methods.MethodRequestTextDocumentCodeAction
}

// HoverRequest is sent from the client to the server to request hover
// information.
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#textDocument_hover
type HoverRequest struct {
	// HoverRequest embeeds the request struct.
	Request
	// Params are the parameters for the hover request.
	Params protocol.HoverParams `json:"params"`
}

// Method returns the method for the hover request
func (r HoverRequest) Method() methods.Method {
	return methods.MethodRequestTextDocumentHover
}

// InitializeRequest is a struct for the initialize request.
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#initialize
type InitializeRequest struct {
	// InitializeRequest embeds the Request struct
	Request
	// Params are the parameters for the initialize request.
	Params protocol.InitializeParams `json:"params"`
}

// Method returns the method for the initialize request.
func (r InitializeRequest) Method() methods.Method {
	return methods.MethodInitialize
}

// InitializedParamsRequest is a struct for the initialized params.
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#initialized
type InitializedParamsRequest struct {
	// InitializedParamsRequest embeds the Request struct
	Response
}

// Method returns the method for the initialized params request.
func (r InitializedParamsRequest) Method() methods.Method {
	return methods.MethodNotificationInitialized
}

// CancelRequest is sent from the client to the server to cancel a request.
type CancelRequest struct {
	// CancelRequest embeds the Request struct
	Request
	// ID is the id of the request to be cancelled.
	ID int `json:"id"`
	// Params are the parameters for the request to be cancelled.
	Params protocol.CancelParams `json:"params"`
}

// Method returns the method for the cancel request
func (r CancelRequest) Method() methods.Method {
	return methods.MethodCancelRequest
}

// ShutdownRequest is the request
//
// Microsoft LSP Docs:
// https://microsoft.github.io/language-server-protocol/specifications/specification-current/#shutdown
type ShutdownRequest struct {
	Request
}

// Method returns the method for the shutdown request
func (r ShutdownRequest) Method() methods.Method {
	return methods.MethodShutdown
}
