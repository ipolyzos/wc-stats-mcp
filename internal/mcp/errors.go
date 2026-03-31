package mcp

import (
	"encoding/json"
	"fmt"
)

// JSON-RPC–style codes used on the wire.
const (
	CodeParseError     = -32700
	CodeInvalidRequest = -32600
	CodeMethodNotFound = -32601
	CodeInvalidParams  = -32602
	CodeInternal       = -32603
	CodeToolError      = -32000
)

func errEnvelope(code int, msg string, details any) *ErrorEnvelope {
	var raw json.RawMessage
	if details != nil {
		b, err := json.Marshal(details)
		if err == nil {
			raw = b
		}
	}
	return &ErrorEnvelope{
		Code:    code,
		Message: msg,
		Details: raw,
	}
}

func toolCallParamsError(err error) *ErrorEnvelope {
	return errEnvelope(CodeInvalidParams, "invalid tools/call params", map[string]string{
		"reason": err.Error(),
	})
}

func unknownMethod(method string) *ErrorEnvelope {
	return errEnvelope(CodeMethodNotFound, fmt.Sprintf("unknown method %q", method), map[string]string{
		"method": method,
	})
}

// ToolExecutionFailed is the standard error for tools/call handler failures (code -32000).
func ToolExecutionFailed(tool string, err error) *ErrorEnvelope {
	if err == nil {
		return errEnvelope(CodeToolError, "tool execution failed", map[string]string{"tool": tool, "reason": "unknown error"})
	}
	return errEnvelope(CodeToolError, "tool execution failed", map[string]string{
		"tool":   tool,
		"reason": err.Error(),
	})
}
