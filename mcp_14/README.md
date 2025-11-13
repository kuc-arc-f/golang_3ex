# mcp_14

 Version: 0.9.1

 date    : 2025/11/11

***

GoLang RAG , Vector search

* Postgres use
* model-embed: gemini-embedding-001
* model: gemini-2.5-flash
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
* build

```
go mod init example.com/go-mcp-server-14
go mod tidy

go build
go run .
```

