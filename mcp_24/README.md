# mcp_24

 Version: 0.9.1

 date    : 2025/12/22

***

GoLang RAG , Search pgvector

* embedding: gemini-embedding-001
* model: gemma-3-27b

***
### Setup

* .env
```
GEMINI_API_KEY="your-key"

PG_USER=root
PG_HOST=localhost
PG_DATABASE=mydb
PG_PASSWORD=admin
PG_PORT=5432
```

***
* build

```
go mod init example.com/go-mcp-server-24
go mod tidy

go build
go run .
```

***
