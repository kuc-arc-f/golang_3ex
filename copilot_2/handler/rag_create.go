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

    "github.com/google/uuid"
    //"github.com/joho/godotenv"
    "github.com/tmc/langchaingo/textsplitter"
	qdrant "github.com/qdrant/go-client/qdrant"
	"google.golang.org/grpc"    
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500
const collectionName = "document-3"

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
}
// リクエストの構造体
type EmbeddingRequest struct {
    Model  string `json:"model"`
    Prompt string `json:"prompt"`
}

// レスポンスの構造体
type EmbeddingResponse struct {
    Embedding []float32 `json:"embedding"`
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
	modelName := "qwen3-embedding:0.6b"

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

    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))

    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("con.len=%d\n", len(fileRow.Content))
        var embed_value = EmbedUserQuery(fileRow.Content)
        fmt.Printf("テキスト: %s\n", fileRow.Content)
        fmt.Printf("埋め込みベクトルの次元数: %d\n", len(embed_value))
        fmt.Printf("最初の5要素: %v\n", embed_value[:5])

        newID := uuid.New().String()

        points := []*qdrant.PointStruct{
            {
                Id: &qdrant.PointId{
                    PointIdOptions: &qdrant.PointId_Uuid{
                        Uuid : newID,
                    },
                },
                Vectors: &qdrant.Vectors{
                    VectorsOptions: &qdrant.Vectors_Vector{
                        Vector: &qdrant.Vector{
                            Data: embed_value,
                        },
                    },
                },
                Payload: map[string]*qdrant.Value{
                    "content": {
                            Kind: &qdrant.Value_StringValue{StringValue: fileRow.Content},
                    },
                },
            },
        }
        _, err = client.Upsert(context.Background(), &qdrant.UpsertPoints{
                CollectionName: collectionName,
                Points:         points,
        })
        if err != nil {
                log.Fatal(err)
        }				

    }
    fmt.Println("データ挿入完了")
}
