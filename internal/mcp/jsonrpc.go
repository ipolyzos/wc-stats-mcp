package mcp

import "encoding/json"

const jsonRPCVersion = "2.0"

// Request is a single JSON-RPC 2.0 request object (batching not supported).
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response is a JSON-RPC 2.0 response object.
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *ErrorEnvelope  `json:"error,omitempty"`
}

// ErrorEnvelope is the stable application error shape (also used as JSON-RPC error object).
type ErrorEnvelope struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Details json.RawMessage `json:"details,omitempty"`
}
