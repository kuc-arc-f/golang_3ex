package models

import (
	"encoding/json"
)
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// ツール呼び出しパラメータ
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}
// ツールパラメータ
type PurchaseParams struct {
	Name string `json:"name"`
	Price    int    `json:"price"`
}
type SearchParams struct {
	Query string `json:"query"`
	PgConectStr string `json:"pg_conect_str"`
}

type Content struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ツール実行結果
type ToolResult struct {
	Content []Content `json:"content"`
}

type Item struct {
	ID          int     `json:"id"`
	Data        string  `json:"data"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
