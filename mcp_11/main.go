package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "os"
    "path/filepath"

    "google.golang.org/genai"
    "github.com/joho/godotenv"
    _ "github.com/mattn/go-sqlite3"
    "github.com/google/uuid"
    "github.com/tmc/langchaingo/textsplitter"
)
const DATA_DIR = "./data"
const CHUNK_SIZE_MAX = 500

type ReadParam struct {
    Content  string    `json:"content"`
    Name     string    `json:"name"`
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
    type OutEmbed struct {
        Embed  []byte    `json:"embeddings"`
    }

    err := godotenv.Load()
    if err != nil {
      log.Fatalf("Error loading .env file: %s", err)
    }	    
    ctx := context.Background()
    client, err := genai.NewClient(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }
    // SQLiteデータベース接続
    db, err := sql.Open("sqlite3", "./embeddings.db")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    contents := []*genai.Content{}
    var fileItems []ReadParam = readTextData()
    fmt.Printf("len=%v\n", len(fileItems))
    
    for i, fileRow := range fileItems {
        fmt.Printf("i=%d, name=%v\n", i, fileRow.Name)
        fmt.Printf("%v\n", i, fileRow.Content)
        contents = append(contents, genai.NewContentFromText(fileRow.Content, genai.RoleUser))
    }

    result, err := client.Models.EmbedContent(ctx,
        "gemini-embedding-001",
        contents,
        nil,
    )
    if err != nil {
        log.Fatal(err)
    }
    var sessionId = ""
    // Embeddingsをデータベースに登録
    for i, embedding := range result.Embeddings {
        embeddingJSON, err := json.Marshal(embedding.Values)
        if err != nil {
            log.Fatal(err)
        }
        fmt.Printf("ベクトルの次元数: i=%d,  %d\n", i, len(embeddingJSON))
        //fmt.Printf("embed: %s\n", string(embeddingJSON))
        fmt.Printf("nm.i=%d,  %v\n", i, fileItems[i].Name)

        newID := uuid.New().String()
        _, err = db.Exec(
            "INSERT INTO embeddings (id, sessid, name, content, embeddings) VALUES (?, ?, ? , ? , ?)",
            newID,
            sessionId,
            fileItems[i].Name,
            fileItems[i].Content,
            embeddingJSON,
        )
        if err != nil {
            log.Fatal(err)
        }

        fmt.Printf("Inserted embedding %d for content: %s\n", i+1, fileItems[i].Name)
    }
    fmt.Println("All embeddings inserted successfully!")
}