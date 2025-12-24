package handler

import (
    "bytes"
    //"context"
    "encoding/json"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "path/filepath"

    "github.com/google/uuid"
    "github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
const collectionName = "doc-collection"

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

const BaseURL = "http://localhost:8000/api/v2"
const Tenant = "default_tenant"
const Database = "default_database"

type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type AddRequest struct {
	IDs        []string                 `json:"ids"`
	Embeddings [][]float32              `json:"embeddings"`
	Metadatas  []map[string]interface{} `json:"metadatas"`
	Documents  []string                 `json:"documents"`
}
func addData(client *http.Client, collectionID string, ids []string, embeddings [][]float32, metadatas []map[string]interface{}, documents []string) error {
	payload := AddRequest{
		IDs:        ids,
		Embeddings: embeddings,
		Metadatas:  metadatas,
		Documents:  documents,
	}
	reqBody, _ := json.Marshal(payload)

	// v2 APIでのadd endpoint: /tenants/{tenant}/databases/{database}/collections/{collection_id}/add
	url := fmt.Sprintf("%s/tenants/%s/databases/%s/collections/%s/add", BaseURL, Tenant, Database, collectionID)

	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		// 失敗した場合、APIパスの問題の可能性があるため詳細を表示
		return fmt.Errorf("ステータスコード %d: %s (URL: %s)", resp.StatusCode, string(body), url)
	}
	return nil
}
/**
*
* @param
*
* @return
 */
func getOrCreateCollection(client *http.Client, name string) (string, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"name":          name,
		"get_or_create": true,
	})

	url := fmt.Sprintf("%s/tenants/%s/databases/%s/collections", BaseURL, Tenant, Database)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("URL: %s, ステータスコード %d: %s", url, resp.StatusCode, string(body))
	}

	var col Collection
	if err := json.NewDecoder(resp.Body).Decode(&col); err != nil {
		return "", err
	}
	return col.ID, nil
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
	client := &http.Client{}
	fmt.Println("ChromaDB (API v2) へ接続を開始します...")

	collectionID, err := getOrCreateCollection(client, collectionName)
	if err != nil {
		log.Fatalf("コレクション作成エラー: %v\n(docker run -p 8000:8000 chromadb/chroma)", err)
	}
	fmt.Printf("コレクション '%s' (ID: %s) を準備しました。\n", collectionName, collectionID)

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

        documents := []string{fileRow.Content}
        newID := uuid.New().String()
        ids := []string{newID}
        metadatas := []map[string]interface{}{ {"tag": "none"} }
        embeddings := [][]float32{embed_value}

        // v2 API endpoint: POST /collections/{collection_id}/add
        if err := addData(client, collectionID, ids, embeddings, metadatas, documents); err != nil {
            log.Fatalf("データ追加エラー: %v", err)
        }        
    }
    fmt.Println("データ挿入完了")
}
