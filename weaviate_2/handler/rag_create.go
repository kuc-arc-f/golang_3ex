package handler

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"

    //"github.com/google/uuid"
    "github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
    client "github.com/weaviate/weaviate-go-client/v5/weaviate"
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
const collectionName = "Article"

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
func CreateVector() {
    fmt.Printf("#CreateVector-start\n")
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
    cfg := client.Config{
        Scheme: "http",
        Host:   "localhost:8080", // Weaviate が Docker やクラウドで動いている場所
    }
    weaviateClient, err := client.NewClient(cfg)
    if err != nil {
        panic(err)
    }    

    fmt.Println("Qdrant client connected")    

    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))

    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("con.len=%d\n", len(fileRow.Content))
        result, err := getEmbedding(apiKey, fileRow.Content)
        if err != nil {
                fmt.Printf("エラー: %v\n", err)
                os.Exit(1)
        }        

        // 結果の表示
        fmt.Printf("テキスト: %s\n", fileRow.Content)
        fmt.Printf("埋め込みベクトルの次元数: %d\n", len(result.Embedding.Values))
        fmt.Printf("最初の5要素: %v\n", result.Embedding.Values[:5])
        var embed_value = result.Embedding.Values

        // クラスへ1件登録
        _, err = weaviateClient.Data().Creator().
            WithClassName(collectionName).
            WithProperties(map[string]interface{}{
                "content": fileRow.Content,
            }).
            WithVector(embed_value).
            Do(context.Background())
        if err != nil {
            panic(err)
        }
        fmt.Println("Add ok")        

    }
    fmt.Println("データ挿入完了")
}
