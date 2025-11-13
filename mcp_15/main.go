package main

import (
	//"context"
	//"database/sql"
	"bufio"
	"encoding/json"
	//"flag"
	//"fmt"
	//"log"
	//"math"

	"os"

	"example.com/go-mcp-server-15/models"
	"example.com/go-mcp-server-15/handler"
	_ "github.com/tursodatabase/libsql-client-go/libsql"
)
// ツールリスト
type ToolsList struct {
	Tools []Tool `json:"tools"`
}

type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema InputSchema `json:"inputSchema"`
}

type InputSchema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

// ツール呼び出しパラメータ
type CallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}


func main() {
	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	for scanner.Scan() {
		line := scanner.Text()
		
		var req models.JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			sendError(writer, nil, -32700, "Parse error")
			continue
		}

		handleRequest(writer, req)
	}
}

func handleRequest(writer *bufio.Writer, req models.JSONRPCRequest) {
	switch req.Method {
	case "initialize":
		handleInitialize(writer, req)
	case "tools/list":
		handleToolsList(writer, req)
	case "tools/call":
		handleToolsCall(writer, req)
	default:
		sendError(writer, req.ID, -32601, "Method not found")
	}
}

func handleInitialize(writer *bufio.Writer, req models.JSONRPCRequest) {
	result := map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"serverInfo": map[string]string{
			"name":    "purchase-server",
			"version": "1.0.0",
		},
		"capabilities": map[string]interface{}{
			"tools": map[string]bool{},
		},
	}
	sendResponse(writer, req.ID, result)
}

func handleToolsList(writer *bufio.Writer, req models.JSONRPCRequest) {
	tools := ToolsList{
		Tools: []Tool{
			{
				Name:        "rag_search",
				Description: "受信した文字から、検索します。",
				InputSchema: InputSchema{
					Type: "object",
					Properties: map[string]Property{
						"input_text": {
							Type:        "string",
							Description: "受信した文字",
						},
					},
					Required: []string{"input_text"},
				},

			},			
		},
	}
	sendResponse(writer, req.ID, tools)
}

/**
*
* @param
*
* @return
*/
func handleToolsCall(writer *bufio.Writer, req models.JSONRPCRequest) {
	var params CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(writer, req.ID, -32602, "Invalid params")
		return
	}
	//if params.Name == "purchase_item" {
	//	handler.PurchaseHnadler(writer, req)
	//	return
	//}
	//if params.Name == "purchase_list" {
	//	handler.PurchaseListHnadler(writer, req)
	//	return
	//}
	if params.Name == "rag_search" {
		handler.RagSearchHnadler(writer, req)
		return
	}
}

/**
*
* @param
*
* @return
*/
func sendResponse(writer *bufio.Writer, id interface{}, result interface{}) {
	resp := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	
	data, _ := json.Marshal(resp)
	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}

func sendError(writer *bufio.Writer, id interface{}, code int, message string) {
	resp := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &models.RPCError{
			Code:    code,
			Message: message,
		},
	}
	
	data, _ := json.Marshal(resp)
	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}
