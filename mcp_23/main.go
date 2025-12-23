package main

import (
    "bytes"
    "context"
    "encoding/json"
    "flag"
	"fmt"
	"io"
    "log"
	"net/http"
    "os"
    "example.com/go-mcp-server-23/config"

    "github.com/jackc/pgx/v5"
    "github.com/pgvector/pgvector-go"    
    //"google.golang.org/genai"
    //"github.com/joho/godotenv"
    "github.com/modelcontextprotocol/go-sdk/mcp"
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
var model = flag.String("model", "gemini-2.0-flash", "the model name, e.g. gemini-2.0-flash")

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
}
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

/**
*
* @param
*
* @return
*/
func searchQuery(query string) string{
    ctx := context.Background()

    // リクエストの構造体
    type Part struct {
        Text string `json:"text"`
    }

    type Content struct {
        Parts []Part `json:"parts"`
    }

    type Request struct {
        Contents []Content `json:"contents"`
    }

    // レスポンスの構造体
    type Candidate struct {
        Content struct {
            Parts []Part `json:"parts"`
        } `json:"content"`
    }    
    type Response struct {
        Candidates []Candidate `json:"candidates"`
    }
    connStr := config.PG_CONNECT_STR
    fmt.Printf("connStr=%s\n", connStr) 
    conn, err := pgx.Connect(ctx, connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(ctx)

	// 環境変数からAPIキーを取得
	apiKey := config.API_KEY

	result, err := getEmbedding(apiKey, query)
	if err != nil {
		fmt.Printf("エラー: %v\n", err)
		os.Exit(1)
	}
	// 結果の表示
	fmt.Printf("テキスト: %s\n", query)
	fmt.Printf("埋め込みベクトルの次元数: %d\n", len(result.Embedding.Values))
	fmt.Printf("最初の5要素: %v\n", result.Embedding.Values[:5])
    var embed_value = result.Embedding.Values

    fmt.Printf("Embedding Vector Length: %d\n", len(embed_value))
    queryVec := pgvector.NewVector(embed_value)

    // --- 2. 類似検索（Nearest Neighbor Search） ---
    // クエリベクトルに近い順に n件取得
    // <-> はユークリッド距離、 <=> はコサイン距離
    rows, err := conn.Query(ctx, "SELECT id, content, embedding FROM documents ORDER BY embedding <-> $1 LIMIT 2", queryVec)
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
func convertFloat32(value []byte) []float32 {
    var float64s []float64
    if err := json.Unmarshal(value, &float64s); err != nil {
        panic(err)
    }        
    float32s := make([]float32, len(float64s))
    for i, v := range float64s {
        float32s[i] = float32(v)
    }        
    //fmt.Printf("float32s.len= %v\n", len(float32s))
    return float32s
}

/**
*
* @param
*
* @return
*/
func main() {
	server := mcp.NewServer(&mcp.Implementation{Name: "greeter"}, nil)

	// The schema considers 'json' and 'jsonschema' struct tags to get argument
	// names and descriptions.
	type args struct {
		Name string `json:"name" jsonschema:"the person to greet"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "greet",
		Description: "say hi",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args args) (*mcp.CallToolResult, any, error) {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Hi " + args.Name},
			},
		}, nil, nil
	})

	type ragArgs struct {
		Query string `json:"query" jsonschema:"the search query"`
	}
	mcp.AddTool(server, &mcp.Tool{
		Name:        "rag_search",
		Description: "search with rag",
	}, func(ctx context.Context, req *mcp.CallToolRequest, args ragArgs) (*mcp.CallToolResult, any, error) {
        var outStr = searchQuery(args.Query)
        //&mcp.TextContent{Text: "Fixed search result for: " + args.Query},
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: outStr},
			},
		}, nil, nil
	})

	// In this case, the server communicates over stdin/stdout.
	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Printf("Server failed: %v", err)
	}
}
