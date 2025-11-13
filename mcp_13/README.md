# mcp_13

 Version: 0.9.1

 date    : 2025/11/12

***

GoLang RAG , Vector data Add

* Postgres use
* model: gemini-embedding-001

***
### Setup

* .env
```
GEMINI_API_KEY="your-key"
```

***
* config/config.go
* postgres set

```
const (
	Host     = "localhost"
	Port     = 5432
	User     = "user1"
	Password = "pass"
	Dbname   = "postgres"
)
```
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
  embeddings BYTEA
);
```
***
* build

```
go mod init example.com/go-mcp-server-13
go mod tidy

go get github.com/lib/pq
go get github.com/google/uuid
go get github.com/joho/godotenv

go build
go run .
```

