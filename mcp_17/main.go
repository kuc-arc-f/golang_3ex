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
func createVector(){
    ctx := context.Background()

    err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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
    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))
    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("con.len=%d\n", len(fileRow.Content))
        var embed_value = EmbedUserQuery(fileRow.Content)

        // 3次元のベクトルデータを作成
        vec := pgvector.NewVector(embed_value)

        _, err = conn.Exec(ctx, "INSERT INTO documents (content, embedding) VALUES ($1, $2)", fileRow.Content, vec)
        if err != nil {
            log.Fatal(err)
        }
    }
    fmt.Println("データ挿入完了")   
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
    fmt.Printf("Embedding Vector Length: %d\n", len(embed_value))

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
    
   
    //generate answer
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if client.ClientConfig().Backend == genai.BackendVertexAI {
		fmt.Println("Calling VertexAI Backend...")
	} else {
		fmt.Println("Calling GeminiAPI Backend...")
	}
	var config *genai.GenerateContentConfig = &genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0.5)}
    var input string = ""
    input = "日本語で、回答して欲しい。\n" + outText
    fmt.Printf("input:\n%s", input)

	// Create a new Chat.
	chat, err := client.Chats.Create(ctx, *model, config, nil)
	if err != nil {
		log.Fatal(err)
	}

	// Send first chat message.
	result, err := chat.SendMessage(ctx, genai.Part{Text: input})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result.Text())    
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
    var query = "磨製石器の発達"
    fmt.Println("全引数:", os.Args)
    if len(os.Args) > 1 {
        // 最初の引数（ユーザーが渡したもの）
        fmt.Println("最初の引数:", os.Args[1])
        var argment = os.Args[1]
        if argment == "create" {
            createVector()
        }
        if argment == "search" {
            fmt.Println("query:", query)
            searchQuery(query)
        }
    }else{
        fmt.Println("none , arg ")
    }  
}
