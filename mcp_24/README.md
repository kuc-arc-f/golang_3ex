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
```

***
* build

```
go mod init example.com/go-mcp-server-24
go mod tidy

go build
```

***
* vector data add
```
go-mcp-server-24.exe create
```

* search
```
go-mcp-server-24.exe search
```
***
