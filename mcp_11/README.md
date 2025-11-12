# mcp_11

 Version: 0.9.1

 date    : 2025/11/11

***

GoLang RAG , Vector data add

* model: gemini-embedding-001
* SQLite database use
***
### Setup

* .env
```
GEMINI_API_KEY="your-key"
```

* dbname: embeddings.db

***
* data file path: ./data
***
* table
```
CREATE TABLE IF NOT EXISTS embeddings (
  id TEXT PRIMARY KEY,
  sessid TEXT,
  name TEXT,
  content TEXT,
  embeddings BLOB
);

PRAGMA journal_mode = WAL;
```
***
* build

```
go mod init example.com/go-mcp-server-11
go mod tidy

go get github.com/google/uuid
go get github.com/mattn/go-sqlite3
go get github.com/joho/godotenv

go build
go run .
```

