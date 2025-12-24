# mcp_25

 Version: 0.9.1

 date    : 2025/12/24

***

GoLang RAG Search , ChromaDB

* embedding: gemini-embedding-001
* model: gemma-3-27b

***
### Setup

* .env
```
GEMINI_API_KEY="your-key"
```
***
* table: ./table.sql
***
* data path: ./data

***
* build

```
go mod init example.com/go-mcp-server-25
go mod tidy

go build
```

***
* vector data add
```
go-mcp-server-25.exe create
```
***
* RAG search
```
go-mcp-server-25.exe search
```

***
### blog

https://zenn.dev/knaka0209/scraps/babfac220459ea

***
