package main

import (
    "bytes"
    "context"
    "encoding/json"
    //"flag"
    "fmt"
    "io"
    "log"
    "net/http"
    "os"
    "example.com/weaviate-2/handler"

    "github.com/joho/godotenv"
    client "github.com/weaviate/weaviate-go-client/v5/weaviate"
    "github.com/weaviate/weaviate-go-client/v5/weaviate/graphql"
    "github.com/weaviate/weaviate/entities/models"	
)

//const collectionName = "doc-collection"
const collectionName = "Article"

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
    cfg := client.Config{
        Scheme: "http",
        Host:   "localhost:8080", // Weaviate が Docker やクラウドで動いている場所
    }
    weaviateClient, err := client.NewClient(cfg)
    if err != nil {
        panic(err)
    }

    fmt.Println("Connected to Weaviate")

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
    nearVect := weaviateClient.GraphQL().NearVectorArgBuilder().
        //WithVector([]float32{0.1, 0.2, 0.3}).
        WithVector(embed_value).
        WithCertainty(0.5)

    // 検索時に 'content' フィールドを取得するように指定
    fields := []graphql.Field{
        {Name: "content"},
    }
//        WithLimit(2)

    resp, err := weaviateClient.GraphQL().Get().
        WithClassName(collectionName).
        WithNearVector(nearVect).
        WithFields(fields...).
        WithLimit(2).
        Do(context.Background())
    if err != nil {
        panic(err)
    }

    // 結果の解析と表示
    if resp.Errors != nil && len(resp.Errors) > 0 {
        for _, e := range resp.Errors {
            fmt.Printf("GraphQL Error: %v\n", e.Message)
        }
    }
    var matches string = ""
    if resp.Data != nil {
        getVal, ok := resp.Data["Get"]
        if !ok {
            return
        }

        getMap, ok := getVal.(map[string]interface{})
        if !ok {
            return
        }

        // コレクション名で結果を取得
        var items []interface{}
        if list, found := getMap[collectionName]; found {
            items = list.([]interface{})
        }

        var contents []string
        for _, item := range items {
            itemMap := item.(map[string]interface{})
            if content, ok := itemMap["content"].(string); ok {
                contents = append(contents, content)
                fmt.Printf("content: %v\n", content)
                matches += content + "\n"
            }
        }
        //fmt.Printf("Search Results (Content): %v\n", contents)
    }
    //fmt.Printf("matches=%v\n",matches)
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

const EMBED_SIZE = 3072 

/**
*
* @param
*
* @return
*/
func createCollections(){
    cfg := client.Config{
        Scheme: "http",
        Host:   "localhost:8080", // Weaviate が Docker やクラウドで動いている場所
    }
    weaviateClient, err := client.NewClient(cfg)
    if err != nil {
        panic(err)
    }

    fmt.Println("Connected to Weaviate")

    // クラスを作る (既に存在する場合はエラーになるが、ここでは簡易的に無視または処理続行)
    cls := &models.Class{
        Class:       collectionName,
        Description: "Books data",
        Properties: []*models.Property{
            {
                Name:     "content",
                DataType: []string{"text"}, // 'string' is deprecated in newer Weaviate, 'text' is preferred, but string works. keeping structure.
            },
        },
    }
    // エラーハンドリング: クラスが既にある場合は無視するなど
    _ = weaviateClient.Schema().ClassCreator().WithClass(cls).Do(context.Background())

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
        if argment == "init" {
            fmt.Println("init-start:")
            createCollections()
        }
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
