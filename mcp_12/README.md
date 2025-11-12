# mcp_12

 Version: 0.9.1

 date    : 2025/11/11

***

GoLang RAG , Search

* model-embed: gemini-embedding-001
* model: gemini-2.5-flash
* SQLite database use
***

### Setup

* .env
```
GEMINI_API_KEY="your-key"
```

* dbname: embeddings.db

***
* build

```
go mod init example.com/go-mcp-server-12
go mod tidy

go build
go run .
```

