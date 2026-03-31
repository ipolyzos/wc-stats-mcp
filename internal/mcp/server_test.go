package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"strings"
	"testing"

	"wc-stats-mcp/internal/tools"
)

func TestServeStdio_toolsCall_ping(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register("server.ping", tools.PingHandler)
	var logBuf bytes.Buffer
	srv := &Server{Reg: reg, Log: log.New(&logBuf, "", 0)}

	req := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"server.ping","arguments":{}}}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer
	if err := srv.ServeStdio(context.Background(), in, &out, &logBuf); err != nil {
		t.Fatal(err)
	}

	var resp Response
	dec := json.NewDecoder(bytes.NewReader(out.Bytes()))
	if err := dec.Decode(&resp); err != nil {
		t.Fatalf("decode: %v\nraw=%q", err, out.String())
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", resp.Error)
	}
	var wrap struct {
		StructuredContent json.RawMessage `json:"structuredContent"`
	}
	if err := json.Unmarshal(resp.Result, &wrap); err != nil {
		t.Fatal(err)
	}
	var payload map[string]bool
	if err := json.Unmarshal(wrap.StructuredContent, &payload); err != nil {
		t.Fatal(err)
	}
	if !payload["ok"] {
		t.Fatalf("expected ok true: %v", payload)
	}
}

func TestServeStdio_unknownTool_errorEnvelope(t *testing.T) {
	reg := tools.NewRegistry()
	var logBuf bytes.Buffer
	srv := &Server{Reg: reg, Log: log.New(&logBuf, "", 0)}
	req := `{"jsonrpc":"2.0","id":42,"method":"tools/call","params":{"name":"missing","arguments":{}}}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer
	if err := srv.ServeStdio(context.Background(), in, &out, &logBuf); err != nil {
		t.Fatal(err)
	}
	var resp Response
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil {
		t.Fatal("expected error")
	}
	if resp.Error.Code == 0 || resp.Error.Message == "" {
		t.Fatalf("invalid error: %+v", resp.Error)
	}
	if len(resp.Error.Details) == 0 {
		t.Fatal("expected details")
	}
}

func TestServeStdio_toolsList(t *testing.T) {
	reg := tools.NewRegistry()
	reg.Register("server.ping", tools.PingHandler)
	var logBuf bytes.Buffer
	srv := &Server{Reg: reg, Log: log.New(&logBuf, "", 0)}
	req := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}` + "\n"
	in := strings.NewReader(req)
	var out bytes.Buffer
	if err := srv.ServeStdio(context.Background(), in, &out, &logBuf); err != nil {
		t.Fatal(err)
	}
	var resp Response
	if err := json.NewDecoder(bytes.NewReader(out.Bytes())).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("error: %+v", resp.Error)
	}
	var body struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(resp.Result, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Tools) != 1 || body.Tools[0].Name != "server.ping" {
		t.Fatalf("unexpected tools: %+v", body)
	}
}
