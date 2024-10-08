package lsp

import (
	"github.com/conneroisu/embedpls/internal/lsp/methods"
	"go.lsp.dev/protocol"
)

// Request is the request to a LSP
type Request struct {
	// RPC is the rpc method for the request
	RPC string `json:"jsonrpc"`
	// ID is the id of the request
	ID int `json:"id,omitempty"`
	// Method is the method for the request
	Method string `json:"method"`
}

// Response is the response of a LSP
type Response struct {
	// RPC is the rpc method for the response
	RPC string `json:"jsonrpc"`
	// ID is the id of the response
	ID int `json:"id,omitempty"`
	// Result string `json:"result"`
	// Error  string `json:"error"`
}

// Method returns the method for the response
func (r Response) Method() methods.Method {
	return protocol.MethodInitialize
}

// Notification is a notification from a LSP
type Notification struct {
	// RPC is the rpc method for the notification.
	RPC string `json:"jsonrpc"`
	// Method is the method for the notification.
	Method string `json:"method"`
}
