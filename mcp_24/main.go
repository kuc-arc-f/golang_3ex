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
    //"google.golang.org/genai"
    "github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
)

const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
var model = flag.String("model", "gemini-2.0-flash", "the model name, e.g. gemini-2.0-flash")
const PG_CONNECT_STR = "postgres://root:admin@localhost:5432/mydb"

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
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

    connStr := PG_CONNECT_STR
    fmt.Printf("connStr=%s\n", connStr) 
    conn, err := pgx.Connect(ctx, connStr)
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close(ctx)

    // テキストの埋め込みを取得
    text := query
    result, err := getEmbedding(apiKey, text)
    if err != nil {
        fmt.Printf("エラー: %v\n", err)
        os.Exit(1)
    }

	// 結果の表示
	fmt.Printf("テキスト: %s\n", text)
	fmt.Printf("埋め込みベクトルの次元数: %d\n", len(result.Embedding.Values))
	fmt.Printf("最初の5要素: %v\n", result.Embedding.Values[:5])
    var embed_value = result.Embedding.Values

    //embedding32s := convertFloat32(embed_value)
    fmt.Printf("Embedding Vector Length: %d\n", len(embed_value))
    queryVec := pgvector.NewVector(embed_value)
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
    //var query = "立春 春分"
    fmt.Println("query:", query)
    searchQuery(query)
}
