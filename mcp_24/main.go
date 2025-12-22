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
    "path/filepath"

    "github.com/jackc/pgx/v5"
    "github.com/pgvector/pgvector-go"    
    "google.golang.org/genai"
    "github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
)

const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
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
	//Embedding []float64 `json:"embedding"`
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
func EmbedUserQuery(query string) []byte{
    err := godotenv.Load()
    if err != nil {
      log.Fatalf("Error loading .env file: %s", err)
    }	    
    ctx := context.Background()
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    contents := []*genai.Content{
        genai.NewContentFromText(query, genai.RoleUser),
    }
    result, err := client.Models.EmbedContent(ctx,
        "gemini-embedding-001",
        contents,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }   
    var respEmbed []byte 
    for i, embedding := range result.Embeddings {
        embeddingJSON, err := json.Marshal(embedding.Values)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("ベクトルの次元数: i=%d,  %d\n", i, len(embeddingJSON))
        respEmbed = embeddingJSON
        //fmt.Printf("embed: %s\n", string(embeddingJSON))
    }
    fmt.Println("All embeddings inserted successfully!")
    return respEmbed    
}

/**
*
* @param
*
* @return
*/
func searchQuery(query string){
    ctx := context.Background()
    err := godotenv.Load()
    if err != nil {
        log.Fatal("Error loading .env file")
    }    
	// APIキーを環境変数から取得
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Println("エラー: GEMINI_API_KEY環境変数を設定してください")
		return
	}

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

    dbHost := os.Getenv("PG_HOST")
    dbUser := os.Getenv("PG_USER")
    dbDatabase := os.Getenv("PG_DATABASE")
    dbPass := os.Getenv("PG_PASSWORD")
    dbPort := os.Getenv("PG_PORT")
    connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
        dbUser, dbPass, dbHost, dbPort, dbDatabase,
    ) 
    fmt.Printf("connStr=%s\n", connStr) 
    conn, err := pgx.Connect(ctx, connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(ctx)

    var embed_value = EmbedUserQuery(query)
    embedding32s := convertFloat32(embed_value)
    fmt.Printf("Embedding Vector Length: %d\n", len(embedding32s))

    // --- 2. 類似検索（Nearest Neighbor Search） ---
    // クエリベクトルに近い順に n件取得
    // <-> はユークリッド距離、 <=> はコサイン距離
    queryVec := pgvector.NewVector(embedding32s)
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
    //fmt.Printf("outText= %s\n", outText)
    var input string = ""
    input = "日本語で、回答して欲しい。\n" + outText
    fmt.Printf("input:\n%s", input)

    //generate answer
 	// APIエンドポイント
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemma-3-27b-it:generateContent?key=%s", apiKey)

	// リクエストボディの作成
	reqBody := Request{
		Contents: []Content{
			{
				Parts: []Part{
					{Text: input},
				},
			},
		},
	}

	// JSONにエンコード
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("JSONエンコードエラー: %v\n", err)
		return
	}

	// POSTリクエストの送信
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("リクエスト送信エラー: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// レスポンスボディの読み取り
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("レスポンス読み取りエラー: %v\n", err)
		return
	}

	// ステータスコードの確認
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("エラー: ステータスコード %d\n", resp.StatusCode)
		fmt.Printf("レスポンス: %s\n", string(body))
		return
	}

	// レスポンスのパース
	var apiResp Response
	if err := json.Unmarshal(body, &apiResp); err != nil {
		fmt.Printf("JSONパースエラー: %v\n", err)
		fmt.Printf("生レスポンス: %s\n", string(body))
		return
	}

	// 結果の表示
	fmt.Println("=== API レスポンス ===")
	if len(apiResp.Candidates) > 0 && len(apiResp.Candidates[0].Content.Parts) > 0 {
		fmt.Println(apiResp.Candidates[0].Content.Parts[0].Text)
	} else {
		fmt.Println("レスポンスが空です")
		fmt.Printf("生レスポンス: %s\n", string(body))
	}

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
func readTextData() []ReadParam{
    fileItem := []ReadParam{}

	entries, err := os.ReadDir(DATA_DIR)
	if err != nil {
		fmt.Println("フォルダ読み込みエラー:", err)
		return nil
	}
    // textsplitter Setting
	splitter := textsplitter.NewRecursiveCharacter(
		textsplitter.WithChunkSize(CHUNK_SIZE_MAX),
		textsplitter.WithChunkOverlap(10),
	)        

    var row ReadParam
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".txt" {
			continue
		}

		path := filepath.Join(DATA_DIR, entry.Name())
        row.Name = entry.Name()

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Println("ファイル読み込みエラー:", err)
			continue
		}
        row.Content = string(data)
        // chunks add
        chunks, err := splitter.SplitText(row.Content)
        if err != nil {
            log.Fatal(err)
        }

        for i, chunk := range chunks {
            fmt.Printf("Chunk %d:\n%s\n---\n", i+1, chunk)
            row.Content = chunk
            fileItem = append(fileItem, row)
        }    
		//fmt.Printf("=== %s ===\n%s\n\n", entry.Name(), string(data))
	}
    return fileItem
}
/**
*
* @param
*
* @return
*/
func main() {
    var query = "二十四節季"
    fmt.Println("query:", query)
    searchQuery(query)
}
