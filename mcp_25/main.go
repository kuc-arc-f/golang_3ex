package main

import (
    "bytes"
    //"context"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "example.com/go-mcp-server-25/handler"

    "github.com/joho/godotenv"
)

var model = flag.String("model", "gemini-2.0-flash", "the model name, e.g. gemini-2.0-flash")
const collectionName = "doc-collection"

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

// ChromaDBの基本URL (v2を使用)
const BaseURL = "http://localhost:8000/api/v2"
const Tenant = "default_tenant"
const Database = "default_database"

// Collection APIレスポンス用構造体
type Collection struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

/**
*
* @param
*
* @return
*/
func sendMessage(input string){
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
};
// 検索用リクエスト構造体
type QueryRequest struct {
	QueryEmbeddings [][]float32 `json:"query_embeddings"`
	NResults        int         `json:"n_results"`
	Include         []string    `json:"include,omitempty"`
}

// 検索結果レスポンス用構造体
type QueryResponse struct {
	IDs       [][]string                 `json:"ids"`
	Distances [][]float64                `json:"distances"`
	Metadatas [][]map[string]interface{} `json:"metadatas"`
	Documents [][]string                 `json:"documents"`
}

/**
*
* @param
*
* @return
*/
func searchQuery(query string){
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

    client := &http.Client{}
    fmt.Println("ChromaDB (API v2) へ接続を開始します...")

    // 1. サーバーの確認 (Heartbeat)
    if err := checkHeartbeat(client); err != nil {
        fmt.Printf("警告: Heartbeatチェック失敗: %v\n", err)
    } else {
        fmt.Println("サーバーは正常に稼働しています。")
    }
    // 2. コレクションの作成 (または取得)
    // v2 API endpoint: POST /tenants/{tenant}/databases/{database}/collections
    collectionID, err := getOrCreateCollection(client, collectionName)
    if err != nil {
        log.Fatalf("コレクション作成エラー: %v\n(docker run -p 8000:8000 chromadb/chroma)", err)
    }
    fmt.Printf("コレクション '%s' (ID: %s) を準備しました。\n", collectionName, collectionID)    

    // 5. 検索
    // v2 API endpoint: POST /collections/{collection_id}/query
    // 検索用Embedding (例: engineeringに近いベクトル)
    queryEmbeddings := embed_value

    payload := QueryRequest{
        QueryEmbeddings: [][]float32{queryEmbeddings},
        NResults:        2,
        Include:         []string{"metadatas", "documents", "distances"},
    }
    reqBody, _ := json.Marshal(payload)

    // v2 APIでのquery endpoint: /tenants/{tenant}/databases/{database}/collections/{collection_id}/query
    url := fmt.Sprintf("%s/tenants/%s/databases/%s/collections/%s/query", BaseURL, Tenant, Database, collectionID)

    resp, err := client.Post(url, "application/json", bytes.NewBuffer(reqBody))
    if err != nil {
        //return err
        return
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        fmt.Printf("ステータスコード %d: %s", resp.StatusCode, string(body))
        return
    }

    var resultResp QueryResponse
    if err := json.NewDecoder(resp.Body).Decode(&resultResp); err != nil {
        return
    }

    var matches string = ""
    fmt.Printf("検索結果 (%d件):\n", len(resultResp.IDs[0]))
    for i, id := range resultResp.IDs[0] {
        fmt.Printf("- ID: %s, Distance: %.4f\n", id, resultResp.Distances[0][i])
        //fmt.Printf("  Document: %s\n", resultResp.Documents[0][i])
        matches += resultResp.Documents[0][i] + "\n"
    }
    //fmt.Printf("matches: %s\n", matches)

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

    sendMessage(input)
    return
}

/**
*
* @param
*
* @return
*/
func checkHeartbeat(client *http.Client) error {
	resp, err := client.Get(BaseURL + "/heartbeat")
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("ステータスコード %d: %s", resp.StatusCode, string(body))
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
    var query = "二十四節季"
    fmt.Println("query:", query)
    fmt.Println("全引数:", os.Args)
    if len(os.Args) > 1 {
        fmt.Println("最初の引数:", os.Args[1])
        var argment = os.Args[1]
        if argment == "search" {
            fmt.Println("query:", query)
            searchQuery(query)
        }
        if argment == "create" {
            fmt.Println("create-start:")
            handler.CreateVector()
        }
    } else {
        fmt.Println("none , arg ")
    }
}
