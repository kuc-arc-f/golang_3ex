# mcp_17

 Version: 0.9.1

 date    : 2025/11/26 

***

GoLang RAG , pgvector

* embedding: qwen3-embedding:0.6b , ollama
* model: gemini-2.0-flash

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
* data file path: ./data

***
* table: ./table.sql

```
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
  id SERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  embedding vector(1024)
);
```
***
* build

```
go mod init example.com/go-mcp-server-17
go mod tidy

go build
```

***

* vertor add

```
go-mcp-server-17.exe create
```

* RAG search
```
go-mcp-server-17.exe search
```
***
### blog

https://zenn.dev/knaka0209/scraps/8f1fab8ffe9062

***
