package handler

import (
    "bytes"
    "context"
    "bufio"
    //"database/sql"
    "encoding/json"
    //"flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    //"math"
    //"time"
    "example.com/go-qdrant-6/models"
    "example.com/go-qdrant-6/config"
    //"google.golang.org/genai"
    //"github.com/joho/godotenv"
    qdrant "github.com/qdrant/go-client/qdrant"
    "google.golang.org/grpc"    
)
const SESSION_ID="sess1"
// リクエストの構造体
type EmbeddingRequest struct {
    Model   string  `json:"model"`
    Content EmbeddingContent `json:"content"`
}

type EmbeddingContent struct {
    Parts []EmbeddingPart `json:"parts"`
}

type EmbeddingPart struct {
    Text string `json:"text"`
}
// レスポンスの構造体
type EmbeddingResponse struct {
    Embedding Embedding `json:"embedding"`
}
type Embedding struct {
    Values []float32 `json:"values"`
}

func getEmbedding(apiKey, text string) (*EmbeddingResponse, error) {
    url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent"

    reqBody := EmbeddingRequest{
        Model: "models/gemini-embedding-001",
        Content: EmbeddingContent{
            Parts: []EmbeddingPart{
                {Text: text},
            },
        },
    }

    // JSONにエンコード
    jsonData, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("JSONエンコードエラー: %w", err)
    }

    // HTTPリクエストの作成
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, fmt.Errorf("リクエスト作成エラー: %w", err)
    }

    // ヘッダーの設定
    req.Header.Set("x-goog-api-key", apiKey)
    req.Header.Set("Content-Type", "application/json")

    // HTTPクライアントでリクエスト送信
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("リクエスト送信エラー: %w", err)
    }
    defer resp.Body.Close()

    // レスポンスボディの読み取り
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("レスポンス読み取りエラー: %w", err)
    }

    // ステータスコードのチェック
    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("APIエラー (ステータス: %d): %s", resp.StatusCode, string(body))
    }

    // JSONデコード
    var embeddingResp EmbeddingResponse
    if err := json.Unmarshal(body, &embeddingResp); err != nil {
        return nil, fmt.Errorf("JSONデコードエラー: %w", err)
    }

    return &embeddingResp, nil
}

const collectionName = "document-2"

/**
*
* @param
*
* @return
*/
func CheckSimalirity(query string, sess string) string {
    type OutEmbed struct {
        Embed  []byte    `json:"embeddings"`
        Content string   `json:"content"`
        Name string   `json:"name"`
    }
    var apiKey = config.API_KEY
    //fmt.Printf("apiKey: %s\n", apiKey)

    result, err := getEmbedding(apiKey, query)
    if err != nil {
        fmt.Printf("エラー: %v\n", err)
        os.Exit(1)
    }    
    fmt.Printf("テキスト: %s\n", query)
    fmt.Printf("埋め込みベクトルの次元数: %d\n", len(result.Embedding.Values))
    fmt.Printf("最初の5要素: %v\n", result.Embedding.Values[:5])
    var embed_value = result.Embedding.Values    
    conn, err := grpc.Dial(
        "localhost:6334",
        grpc.WithInsecure(),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    client := qdrant.NewPointsClient(conn)
    fmt.Println("Qdrant client connected")
    var queryVector = embed_value

    resp, err := client.Search(context.Background(), &qdrant.SearchPoints{
        CollectionName: collectionName,
        Vector:         queryVector,
        Limit:          2,
        WithPayload: &qdrant.WithPayloadSelector{
            SelectorOptions: &qdrant.WithPayloadSelector_Enable{
                Enable: true,
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }   
    var matches string = ""
    for _, p := range resp.Result {
        var contentStr = p.Payload["content"].GetStringValue()
        matches += contentStr + "\n"
    }   
    //fmt.Printf("matches=%v\n", matches)

    var outText string = ""
    if (len(matches) > 0){
        outText = `context:` + matches + "\n"
        outText += `user query:` + query + "\n"
    }else{
        outText =`user query:` + query + "\n"
    }
    return outText
}

/**
*
* @param
*
* @return
*/
func RagSearchHnadler(writer *bufio.Writer, req models.JSONRPCRequest) {
    type SearchParams struct {
        Query string `json:"query"`
    }

    var params models.CallToolParams
    if err := json.Unmarshal(req.Params, &params); err != nil {
        sendError(writer, req.ID, -32602, "Invalid params")
        return
    }
    var args SearchParams
    if err := json.Unmarshal(params.Arguments, &args); err != nil {
        sendError(writer, req.ID, -32602, "Invalid arguments")
        return
    }
    var query string = args.Query
    log.Printf("query= %v", query)

    /*
    err := godotenv.Load()
    if err != nil {
    log.Fatalf("Error loading .env file: %s", err)
    }    
    */

    var input = CheckSimalirity(query, SESSION_ID)
    input = "日本語で、回答して欲しい。\n" + input
    log.Printf("input=%v",  input)
    toolResult := models.ToolResult{
        Content: []models.Content{
            {
                Type: "text",
                Text: input,
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
