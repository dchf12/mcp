package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// エラーコード定数（JSON-RPC 2.0仕様に準拠）
const (
	ErrorParseError     = -32700
	ErrorInvalidRequest = -32600
	ErrorMethodNotFound = -32601
	ErrorInvalidParams  = -32602
	ErrorInternalError  = -32603
	ErrorDivideByZero   = -32000 // カスタムエラー
)

// Server はMCPサーバーの状態を管理します
type Server struct {
	name        string
	version     string
	initialized bool
}

// Implementation はMCPの実装情報を表します
type Implementation struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapabilities はクライアントがサポートする機能を表します
type ClientCapabilities struct {
	Experimental map[string]any   `json:"experimental,omitempty"`
	Roots        *RootsCapability `json:"roots,omitempty"`
	Sampling     map[string]any   `json:"sampling,omitempty"`
}

// RootsCapability はルート関連の機能をサポートするかを示します
type RootsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// ServerCapabilities はサーバーがサポートする機能を表します
type ServerCapabilities struct {
	Experimental map[string]any   `json:"experimental,omitempty"`
	Tools        *ToolsCapability `json:"tools,omitempty"`
}

// ToolsCapability はツール関連の機能をサポートするかを示します
type ToolsCapability struct {
	ListChanged bool `json:"listChanged"`
}

// Request は JSON-RPC 2.0 リクエストを表します
type Request struct {
	JsonRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
	ID      any             `json:"id"`
}

// Response は JSON-RPC 2.0 レスポンスを表します
type Response struct {
	JsonRPC string `json:"jsonrpc"`
	Result  any    `json:"result,omitempty"`
	Error   *Error `json:"error,omitempty"`
	ID      any    `json:"id"`
}

// Error は JSON-RPC 2.0 エラーを表します
type Error struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// InitializeParams は初期化リクエストのパラメータを表します
type InitializeParams struct {
	ClientInfo      Implementation     `json:"clientInfo"`
	Capabilities    ClientCapabilities `json:"capabilities"`
	ProtocolVersion string             `json:"protocolVersion"`
}

// InitializeResult は初期化の結果を表します
type InitializeResult struct {
	ServerInfo      Implementation     `json:"serverInfo"`
	Capabilities    ServerCapabilities `json:"capabilities"`
	ProtocolVersion string             `json:"protocolVersion"`
	Instructions    string             `json:"instructions,omitempty"`
}

// Tool はサーバーが提供するツールを表します
type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	InputSchema map[string]any `json:"inputSchema"`
}

// ListToolsResult はツール一覧のレスポンスを表します
type ListToolsResult struct {
	Tools      []Tool `json:"tools"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// CalcParams は計算ツールのパラメータを表します
type CalcParams struct {
	A float64 `json:"a"`
	B float64 `json:"b"`
}

// ToolRequest はツール呼び出しのリクエストを表します
type ToolRequest struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

func main() {
	server := &Server{
		name:        "go-calculator-server",
		version:     "0.0.1",
		initialized: false,
	}
	server.run()
}

// getAvailableTools は利用可能なツールの一覧を返します
func getAvailableTools() []Tool {
	commonSchema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"a": map[string]any{
				"type": "number",
			},
			"b": map[string]any{
				"type": "number",
			},
		},
		"required": []string{"a", "b"},
	}

	return []Tool{
		{
			Name:        "add",
			Description: "Add two numbers",
			InputSchema: commonSchema,
		},
		{
			Name:        "subtract",
			Description: "Subtract two numbers",
			InputSchema: commonSchema,
		},
		{
			Name:        "multiply",
			Description: "Multiply two numbers",
			InputSchema: commonSchema,
		},
		{
			Name:        "divide",
			Description: "Divide first number by second number",
			InputSchema: commonSchema,
		},
	}
}

// handleToolCall はツールの呼び出しを処理します
func (s *Server) handleToolCall(name string, args json.RawMessage) (any, *Error) {
	var params CalcParams
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, &Error{
			Code:    ErrorInvalidParams,
			Message: "Invalid arguments",
			Data:    err.Error(),
		}
	}

	result := map[string]any{
		"content": []map[string]any{
			{
				"type": "text",
			},
		},
	}

	switch name {
	case "add":
		result["content"].([]map[string]any)[0]["text"] = fmt.Sprintf("%f", params.A+params.B)
		return result, nil
	case "subtract":
		result["content"].([]map[string]any)[0]["text"] = fmt.Sprintf("%f", params.A-params.B)
		return result, nil
	case "multiply":
		result["content"].([]map[string]any)[0]["text"] = fmt.Sprintf("%f", params.A*params.B)
		return result, nil
	case "divide":
		if params.B == 0 {
			return nil, &Error{
				Code:    ErrorDivideByZero,
				Message: "Division by zero",
			}
		}
		result["content"].([]map[string]any)[0]["text"] = fmt.Sprintf("%f", params.A/params.B)
		return result, nil
	default:
		return nil, &Error{
			Code:    ErrorMethodNotFound,
			Message: fmt.Sprintf("Tool '%s' not found", name),
		}
	}
}

func (s *Server) run() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var req Request
		if err := decoder.Decode(&req); err != nil {
			fmt.Fprintf(os.Stderr, "Error decoding request: %v\n", err)
			continue
		}

		var resp Response
		resp.JsonRPC = "2.0"
		resp.ID = req.ID

		switch req.Method {
		case "initialize":
			var initParams InitializeParams
			if err := json.Unmarshal(req.Params, &initParams); err != nil {
				resp.Error = &Error{
					Code:    ErrorInvalidParams,
					Message: "Invalid initialize params",
					Data:    err.Error(),
				}
				break
			}

			// 初期化処理
			s.initialized = true

			resp.Result = InitializeResult{
				ServerInfo: Implementation{
					Name:    s.name,
					Version: s.version,
				},
				ProtocolVersion: initParams.ProtocolVersion,
				Capabilities: ServerCapabilities{
					Tools: &ToolsCapability{
						ListChanged: true,
					},
				},
				Instructions: "This server provides basic arithmetic operations through tools.",
			}

		case "shutdown":
			resp.Result = struct{}{}
			if err := encoder.Encode(resp); err != nil {
				fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
			}
			os.Exit(0)

		case "tools/list":
			if !s.initialized {
				resp.Error = &Error{
					Code:    ErrorInternalError,
					Message: "Server not initialized",
				}
				break
			}
			resp.Result = ListToolsResult{
				Tools: getAvailableTools(),
			}

		case "tools/call":
			if !s.initialized {
				resp.Error = &Error{
					Code:    ErrorInternalError,
					Message: "Server not initialized",
				}
				break
			}
			var toolReq ToolRequest
			if err := json.Unmarshal(req.Params, &toolReq); err != nil {
				resp.Error = &Error{
					Code:    ErrorInvalidParams,
					Message: "Invalid params",
					Data:    err.Error(),
				}
				break
			}
			result, err := s.handleToolCall(toolReq.Name, toolReq.Arguments)
			if err != nil {
				resp.Error = err
			} else {
				resp.Result = result
			}

		default:
			resp.Error = &Error{
				Code:    ErrorMethodNotFound,
				Message: fmt.Sprintf("Method '%s' not found", req.Method),
			}
		}

		if err := encoder.Encode(resp); err != nil {
			fmt.Fprintf(os.Stderr, "Error encoding response: %v\n", err)
		}
	}
}
