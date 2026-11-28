package prowl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// MCPServer serves Prowl's cited SQLite code index (and optional router) over MCP stdio.
type MCPServer struct {
	idx    *Index
	router *Router
}

// NewMCPServer constructs an MCP stdio server backed by idx and optional router.
func NewMCPServer(idx *Index, router *Router) *MCPServer {
	return &MCPServer{idx: idx, router: router}
}

type rpcRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type rpcResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  any             `json:"result,omitempty"`
	Error   *rpcError       `json:"error,omitempty"`
}

type rpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// Serve reads newline-delimited JSON-RPC 2.0 requests from in and writes responses to out.
func (s *MCPServer) Serve(ctx context.Context, in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	enc := json.NewEncoder(out)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var req rpcRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			_ = enc.Encode(rpcResponse{
				JSONRPC: "2.0",
				ID:      json.RawMessage("null"),
				Error:   &rpcError{Code: -32700, Message: "parse error"},
			})
			continue
		}
		// Notifications have no ID and must not receive a response.
		if len(req.ID) == 0 || string(req.ID) == "null" {
			continue
		}
		resp := s.handle(ctx, req)
		if err := enc.Encode(resp); err != nil {
			return err
		}
	}
	return scanner.Err()
}

func (s *MCPServer) handle(ctx context.Context, req rpcRequest) rpcResponse {
	switch req.Method {
	case "initialize":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"protocolVersion": "2024-11-05",
				"capabilities": map[string]any{
					"tools": map[string]any{},
				},
				"serverInfo": map[string]any{
					"name":    "prowl",
					"version": "0.1.0",
				},
			},
		}
	case "ping":
		return rpcResponse{JSONRPC: "2.0", ID: req.ID, Result: map[string]any{}}
	case "tools/list":
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result:  map[string]any{"tools": s.tools()},
		}
	case "tools/call":
		var params struct {
			Name      string          `json:"name"`
			Arguments json.RawMessage `json:"arguments"`
		}
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return rpcResponse{JSONRPC: "2.0", ID: req.ID, Error: &rpcError{Code: -32602, Message: "invalid params"}}
		}
		text, isErr := s.callTool(ctx, params.Name, params.Arguments)
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Result: map[string]any{
				"content": []map[string]any{
					{"type": "text", "text": text},
				},
				"isError": isErr,
			},
		}
	default:
		return rpcResponse{
			JSONRPC: "2.0",
			ID:      req.ID,
			Error:   &rpcError{Code: -32601, Message: fmt.Sprintf("method not found: %s", req.Method)},
		}
	}
}

func (s *MCPServer) tools() []mcpTool {
	tools := []mcpTool{
		{
			Name:        "prowl_search",
			Description: "Search the secret-masked SQLite code index and return cited file:line-range matches.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{"type": "string", "description": "Search keywords or symbol name"},
					"limit": map[string]any{"type": "integer", "description": "Max results (default 10, max 50)"},
				},
				"required": []string{"query"},
			},
		},
		{
			Name:        "prowl_read",
			Description: "Read a cited line range from a workspace file with automatic secret masking.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"path":       map[string]any{"type": "string", "description": "Workspace-relative file path"},
					"start_line": map[string]any{"type": "integer", "description": "1-based start line"},
					"end_line":   map[string]any{"type": "integer", "description": "1-based end line (0 for EOF)"},
				},
				"required": []string{"path"},
			},
		},
		{
			Name:        "prowl_index",
			Description: "Incrementally index the workspace into SQLite, masking unambiguous vendor secrets before write.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
		{
			Name:        "prowl_stats",
			Description: "Return counts of indexed files, chunks, and masked secrets.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
		},
	}
	if s.router != nil {
		tools = append(tools, mcpTool{
			Name:        "prowl_route",
			Description: "Run the bounded 24 KiB Zero-LLM classifier and return the selected model from the active pool.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"prompt": map[string]any{"type": "string", "description": "Latest user turn"},
				},
				"required": []string{"prompt"},
			},
		})
	}
	return tools
}

func (s *MCPServer) callTool(ctx context.Context, name string, rawArgs json.RawMessage) (string, bool) {
	switch name {
	case "prowl_search":
		var args struct {
			Query string `json:"query"`
			Limit int    `json:"limit"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return err.Error(), true
		}
		hits, err := s.idx.Search(ctx, args.Query, args.Limit)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(hits, "", "  ")
		return string(b), false

	case "prowl_read":
		var args struct {
			Path      string `json:"path"`
			StartLine int    `json:"start_line"`
			EndLine   int    `json:"end_line"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return err.Error(), true
		}
		txt, err := s.idx.ReadRange(args.Path, args.StartLine, args.EndLine)
		if err != nil {
			return err.Error(), true
		}
		return txt, false

	case "prowl_index":
		res, err := s.idx.IndexWorkspace(ctx)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		return string(b), false

	case "prowl_stats":
		st, err := s.idx.Stats(ctx)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(st, "", "  ")
		return string(b), false

	case "prowl_route":
		if s.router == nil {
			return "router not configured", true
		}
		var args struct {
			Prompt string `json:"prompt"`
		}
		if err := json.Unmarshal(rawArgs, &args); err != nil {
			return err.Error(), true
		}
		dec, _, err := s.router.RouteTurn(args.Prompt)
		if err != nil {
			return err.Error(), true
		}
		b, _ := json.MarshalIndent(dec, "", "  ")
		return string(b), false

	default:
		return fmt.Sprintf("unknown tool: %s", name), true
	}
}
