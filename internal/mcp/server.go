package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"wc-stats-mcp/internal/tools"
)

// Server handles JSON-RPC lines on stdio: responses only on out; logs on logOut.
type Server struct {
	Reg *tools.Registry
	Log *log.Logger
}

// ServeStdio reads one JSON-RPC request per line from in, writes one response per line to out.
func (s *Server) ServeStdio(ctx context.Context, in io.Reader, out io.Writer, logOut io.Writer) error {
	if s.Reg == nil {
		s.Reg = tools.NewRegistry()
	}
	if s.Log == nil {
		s.Log = log.New(logOut, "", log.LstdFlags|log.Lmicroseconds)
	}

	sc := bufio.NewScanner(in)
	buf := make([]byte, 0, 64*1024)
	const maxLine = 4 << 20
	sc.Buffer(buf, maxLine)

	enc := json.NewEncoder(out)
	enc.SetEscapeHTML(false)

	for sc.Scan() {
		line := sc.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}
		resp := s.handleLine(ctx, line)
		if resp == nil {
			continue
		}
		if err := enc.Encode(resp); err != nil {
			s.Log.Printf("encode response: %v", err)
			return err
		}
	}
	return sc.Err()
}

func (s *Server) handleLine(ctx context.Context, line []byte) *Response {
	var req Request
	if err := json.Unmarshal(line, &req); err != nil {
		return &Response{
			JSONRPC: jsonRPCVersion,
			Error:   errEnvelope(CodeParseError, "parse error", map[string]string{"reason": err.Error()}),
		}
	}
	if req.JSONRPC != jsonRPCVersion {
		return replyInvalid(req.ID, errEnvelope(CodeInvalidRequest, "jsonrpc must be \"2.0\"", nil))
	}
	if req.Method == "" {
		return replyInvalid(req.ID, errEnvelope(CodeInvalidRequest, "method required", nil))
	}

	// Notifications have no id — no response on stdout.
	if len(req.ID) == 0 || string(req.ID) == "null" {
		s.Log.Printf("notification method=%s", req.Method)
		return nil
	}

	switch req.Method {
	case "initialize":
		return s.handleInitialize(req.ID)
	case "tools/list":
		return s.handleToolsList(req.ID)
	case "tools/call":
		return s.handleToolsCall(ctx, req)
	default:
		return &Response{JSONRPC: jsonRPCVersion, ID: req.ID, Error: unknownMethod(req.Method)}
	}
}

func replyInvalid(id json.RawMessage, env *ErrorEnvelope) *Response {
	return &Response{JSONRPC: jsonRPCVersion, ID: id, Error: env}
}

func (s *Server) handleInitialize(id json.RawMessage) *Response {
	res := map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]any{
			"tools": map[string]any{},
		},
		"serverInfo": map[string]string{
			"name":    "wc-stats-mcp",
			"version": "0.0.0-dev",
		},
	}
	raw, _ := json.Marshal(res)
	return &Response{JSONRPC: jsonRPCVersion, ID: id, Result: raw}
}

func (s *Server) handleToolsList(id json.RawMessage) *Response {
	names := s.Reg.Names()
	list := make([]map[string]any, 0, len(names))
	for _, n := range names {
		desc, schema := tools.ToolMeta(n)
		list = append(list, map[string]any{
			"name":        n,
			"description": desc,
			"inputSchema": schema,
		})
	}
	res, _ := json.Marshal(map[string]any{"tools": list})
	return &Response{JSONRPC: jsonRPCVersion, ID: id, Result: res}
}

type toolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func (s *Server) handleToolsCall(ctx context.Context, req Request) *Response {
	var p toolsCallParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return &Response{JSONRPC: jsonRPCVersion, ID: req.ID, Error: toolCallParamsError(err)}
		}
	}
	if p.Name == "" {
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Error:   errEnvelope(CodeInvalidParams, "tool name required", map[string]string{"field": "name"}),
		}
	}
	if p.Arguments == nil {
		p.Arguments = json.RawMessage(`{}`)
	}

	out, err := s.Reg.Call(ctx, p.Name, p.Arguments)
	if err != nil {
		return &Response{JSONRPC: jsonRPCVersion, ID: req.ID, Error: ToolExecutionFailed(p.Name, err)}
	}

	// MCP-shaped result; structuredContent carries the tool JSON.
	inner := map[string]json.RawMessage{
		"structuredContent": out,
	}
	wrapped, err := json.Marshal(inner)
	if err != nil {
		return &Response{
			JSONRPC: jsonRPCVersion,
			ID:      req.ID,
			Error:   errEnvelope(CodeInternal, fmt.Sprintf("marshal result: %v", err), nil),
		}
	}
	return &Response{JSONRPC: jsonRPCVersion, ID: req.ID, Result: wrapped}
}
