package main

import (
    //"bytes"
    "context"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "example.com/copilot_2/handler"

    copilot "github.com/github/copilot-sdk/go"
    qdrant "github.com/qdrant/go-client/qdrant"
    "google.golang.org/grpc"    
)

const collectionName = "document-3"

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
	client := copilot.NewClient(nil)
	if err := client.Start(); err != nil {
		log.Fatal(err)
	}
	defer client.Stop()

	session, err := client.CreateSession(&copilot.SessionConfig{Model: "gpt-4.1"})
	if err != nil {
		log.Fatal(err)
	}

	response, err := session.SendAndWait(copilot.MessageOptions{Prompt: input}, 0)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(*response.Data.Content)
	os.Exit(0)  
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
    // テキストの埋め込みを取得
    text := query
    var embed_value = handler.EmbedUserQuery(text)

    fmt.Printf("テキスト: %s\n", text)
    fmt.Printf("埋め込みベクトルの次元数: %d\n", len(embed_value))
    fmt.Printf("最初の5要素: %v\n", embed_value[:5])
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
        Limit:          1,
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
        //fmt.Printf("ID=%v score=%.4f \npayload=%v\n",
        //	p.Id, p.Score, contentStr)
    }   
    fmt.Printf("matches=%v\n",matches)
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

const EMBED_SIZE = 1024

/**
*
* @param
*
* @return
*/
func createCollections(){
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
    collectionsClient := qdrant.NewCollectionsClient(conn)
    _, err = collectionsClient.Create(context.Background(), &qdrant.CreateCollection{
        CollectionName: collectionName,
        VectorsConfig: &qdrant.VectorsConfig{
            Config: &qdrant.VectorsConfig_Params{
                Params: &qdrant.VectorParams{
                    Size:     EMBED_SIZE,
                    Distance: qdrant.Distance_Cosine,
                },
            },
        },
    })
    if err != nil {
        log.Fatal(err)
    }

    _ = client
}
/**
*
* @param
*
* @return
*/
func main() {
    var query = ""
    fmt.Println("全引数:", os.Args)
    fmt.Println("全引数.count:", len(os.Args))
    if len(os.Args) > 1 {
        fmt.Println("最初の引数:", os.Args[1])
        var argment = os.Args[1]
        if argment == "init" {
            fmt.Println("init-start:")
            createCollections()
        }
        if argment == "search" {
            if len(os.Args) < 3 {
                fmt.Println("error: argment, ex: copilot_2.exe search hello")
                return
            }
            query = os.Args[2]
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
