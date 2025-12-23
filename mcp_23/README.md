# mcp_23

 Version: 0.9.1

 date    : 2025/12/22

***

GoLang MCP Server , RAG Search 

* pgvector use
* embedding: gemini-embedding-001

***
### Setup

* config/config.go
```
const API_KEY = "your-key"
const PG_CONNECT_STR = "postgres://root:admin@localhost:5432/mydb"
```

***
* build

```
go mod init example.com/go-mcp-server-23
go mod tidy

go build
go run .
```

***
