package handler

import (
    "bufio"
    "bytes"
    "context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    //"os"
    "example.com/go-mcp-server-18/models"

    "github.com/jackc/pgx/v5"
    "github.com/pgvector/pgvector-go"    
    //"github.com/joho/godotenv"
)

const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
const MODEL_EMBED = "qwen3-embedding:0.6b"
var model = flag.String("model", "gemini-2.0-flash", "the model name, e.g. gemini-2.0-flash")

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
}
// EmbeddingRequest represents the request body for the Ollama embedding API.
type EmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbeddingResponse represents the response body from the Ollama embedding API.
type EmbeddingResponse struct {
	Embedding []float32 `json:"embedding"`
}
// RequestPayload represents the JSON payload for the Ollama API
type RequestPayload struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}
// ResponsePayload represents the JSON response from the Ollama API
type ResponsePayload struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

/**
*
* @param
*
* @return
*/
func EmbedUserQuery(query string)[]float32{
	// Configuration
	ollamaURL := "http://localhost:11434/api/embeddings"
	modelName := MODEL_EMBED

	// Create request body
	reqBody := EmbeddingRequest{
		Model:  modelName,
		Prompt: query,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", ollamaURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error sending request: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Fatalf("API request failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	var embeddingResp EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embeddingResp); err != nil {
		log.Fatalf("Error decoding response: %v", err)
	}

	// Output results
	fmt.Printf("Model: %s\n", modelName)
	fmt.Printf("Prompt: %s\n", query)
	fmt.Printf("Embedding Vector Length: %d\n", len(embeddingResp.Embedding))
	
	// Print first few dimensions to verify
	if len(embeddingResp.Embedding) > 5 {
		fmt.Printf("First 5 dimensions: %v\n", embeddingResp.Embedding[:5])
	} else {
		fmt.Printf("Embedding: %v\n", embeddingResp.Embedding)
	}
	
    return embeddingResp.Embedding   
}

/**
*
* @param
*
* @return
*/
func searchQuery(query string, pgConectStr string) string{
    ctx := context.Background()
    var embed_value = EmbedUserQuery(query)
    fmt.Printf("Embedding Vector Length: %d\n", len(embed_value))

    fmt.Printf("connStr=%s\n", pgConectStr) 
    conn, err := pgx.Connect(ctx, pgConectStr)
    if err != nil {
      log.Fatal(err)
    }
    defer conn.Close(ctx)


    // --- 2. 類似検索（Nearest Neighbor Search） ---
	// クエリベクトルに近い順に5件取得
	// <-> はユークリッド距離、 <=> はコサイン距離
    queryVec := pgvector.NewVector(embed_value)
    rows, err := conn.Query(ctx, "SELECT id, content, embedding FROM documents ORDER BY embedding <-> $1 LIMIT 5", queryVec)
    if err != nil {
      log.Fatal(err)
    }
    defer rows.Close()

    var matches string = ""
    for rows.Next() {
        var id int64
        var content string
        var embedding pgvector.Vector
        if err := rows.Scan(&id, &content, &embedding); err != nil {
            log.Fatal(err)
        }
        matches += content + "\n"
        fmt.Printf("ID: %d, cont.len: %d\n", id, len(content))
    }
    var outText string = ""
    if (len(matches) > 0){
        outText = `context:` + matches + "\n"
        outText += `user query:` + query + "\n"
    }else{
        outText =`user query:` + query + "\n"
    }    
    var input string = ""
    input = "日本語で、回答して欲しい。\n" + outText
    fmt.Printf("input:\n%s", input)
    return input
}

/**
*
* @param
*
* @return
*/
func SearchHnadler(writer *bufio.Writer, req models.JSONRPCRequest) {
    var params models.CallToolParams
    if err := json.Unmarshal(req.Params, &params); err != nil {
      sendError(writer, req.ID, -32602, "Invalid params")
      return
    }
    var args models.SearchParams
    if err := json.Unmarshal(params.Arguments, &args); err != nil {
      sendError(writer, req.ID, -32602, "Invalid arguments")
      return
    }
    log.Printf("q= %s", args.Query)
    log.Printf("conn_str= %s", args.PgConectStr)

    var outStr = searchQuery(args.Query, args.PgConectStr)
	
    toolResult := models.ToolResult{
      Content: []models.Content{
        {
          Type: "text",
          Text: outStr,
        },
      },
    }
    sendResponse(writer, req.ID, toolResult)    
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
