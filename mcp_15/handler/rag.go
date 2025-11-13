package handler

import (
	"context"
	"bufio"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	//"time"

	"example.com/go-mcp-server-15/models"
	"example.com/go-mcp-server-15/config"
    _ "github.com/lib/pq"
    "google.golang.org/genai"
    "github.com/joho/godotenv"
)

var model = flag.String("model", "gemini-2.0-flash", "the model name, e.g. gemini-2.0-flash")
const SESSION_ID="sess1"

var db *sql.DB


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
        fmt.Printf("i=%d,  %d\n", i, len(embeddingJSON))
        respEmbed = embeddingJSON
        //fmt.Printf("embed: %s\n", string(embeddingJSON))
    }
    //fmt.Println("All embeddings inserted successfully!")
    return respEmbed    
}
/**
*
* @param
*
* @return
*/
func cosineSimilarity(a []float32, b []float32) (float64, error) {
    if len(a) != len(b) {
        return 0, fmt.Errorf("vectors must have the same length")
    }

    var dotProduct, aMagnitude, bMagnitude float64
    for i := 0; i < len(a); i++ {
        dotProduct += float64(a[i] * b[i])
        aMagnitude += float64(a[i] * a[i])
        bMagnitude += float64(b[i] * b[i])
    }

    if aMagnitude == 0 || bMagnitude == 0 {
        return 0, nil
    }

    return dotProduct / (math.Sqrt(aMagnitude) * math.Sqrt(bMagnitude)), nil
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

func connectDB() (*sql.DB, error) {
	psqlInfo := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Dbname,
	)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	// データベースへの接続を確認
	err = db.Ping()
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

/**
*
* @param
*
* @return
*/
func CheckSimalirity(query string, sess string) string {
    type OutEmbed struct {
        Embed  []byte    `json:"embeddings"`
        Content string   `json:"content"`
        Name string   `json:"name"`
    }

    err := godotenv.Load()
    if err != nil {
      log.Fatalf("Error loading .env file: %s", err)
    }

    db, err := connectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	select_sql := fmt.Sprintf(`SELECT name, content, embeddings FROM embeddings`)
  //fmt.Printf("sql=%s\n", select_sql)
	rows, err := db.Query(select_sql)
	if err != nil {
        log.Fatal(err)
	}
	defer rows.Close()

    var targetByte = EmbedUserQuery(query)
    embedding32s := convertFloat32(targetByte)
    //fmt.Printf("len.embedding32s=%v\n", len(embedding32s))

    var matches string = ""
	var outData []OutEmbed
	for rows.Next() {
		var row OutEmbed
		err := rows.Scan(&row.Name, &row.Content, &row.Embed)
		if err != nil {
          log.Fatal(err)
		}
    float32s := convertFloat32(row.Embed)
    similarity, _ := cosineSimilarity(embedding32s, float32s)
    //fmt.Printf("sim= %v name=%s\n", similarity, row.Name)
    if(similarity > 0.6) {
      matches += row.Content + "\n"
    }

		outData = append(outData, row)

	}
    //fmt.Printf("matches= %v\n", len(matches))
    var outText string = ""
    if (len(matches) > 0){
        outText = `context:` + matches + "\n"
        outText += `user query:` + query + "\n"
    }else{
        outText =`user query:` + query + "\n"
    }
    return outText
}

/**
*
* @param
*
* @return
*/
func RagSearchHnadler(writer *bufio.Writer, req models.JSONRPCRequest) {
  type SearchParams struct {
    InputText string `json:"input_text"`
  }

  var params models.CallToolParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		sendError(writer, req.ID, -32602, "Invalid params")
		return
	}
	var args SearchParams
	if err := json.Unmarshal(params.Arguments, &args); err != nil {
		sendError(writer, req.ID, -32602, "Invalid arguments")
		return
	}
	//log.Printf("arg= %v", args.InputText)
  var query string = args.InputText
	//log.Printf("query= %v", query)

	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}    
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}
	if client.ClientConfig().Backend == genai.BackendVertexAI {
		fmt.Println("Calling VertexAI Backend...")
	} else {
		//fmt.Println("Calling GeminiAPI Backend...")
	}
	var config *genai.GenerateContentConfig = &genai.GenerateContentConfig{Temperature: genai.Ptr[float32](0.5)}

  var input = CheckSimalirity(query, SESSION_ID)
  input = "日本語で、回答して欲しい。\n" + input
  //log.Printf("input=%v",  input)

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
	//fmt.Println(result.Text())
	//log.Printf(" jsonString %s", jsonString)

	toolResult := models.ToolResult{
		Content: []models.Content{
			{
				Type: "text",
				Text: result.Text(),
			},
		},
	}

	sendResponse(writer, req.ID, toolResult)
}

/**
*
* @param
*
* @return
*/
func sendResponse(writer *bufio.Writer, id interface{}, result interface{}) {
	resp := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Result:  result,
	}
	
	data, _ := json.Marshal(resp)
	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}

func sendError(writer *bufio.Writer, id interface{}, code int, message string) {
	resp := models.JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &models.RPCError{
			Code:    code,
			Message: message,
		},
	}
	
	data, _ := json.Marshal(resp)
	writer.Write(data)
	writer.WriteByte('\n')
	writer.Flush()
}
