# weaviate_3

 Version: 0.9.1

 date    : 2025/12/27

***

GoLang MCP Server , RAG Search 

* Weaviate use
* embedding: gemini-embedding-001
* GEMINI-CLI

***
* vector data add

https://github.com/kuc-arc-f/golang_3ex/tree/main/weaviate_2

***
### Setup

* config/config.go
* API_KEY: GEMINI API KEY
```
const API_KEY = "your-key"
```

***
* build

```
go mod init example.com/weaviate-3
go mod tidy

go build
```

***
* settings.json , GEMINI-CLI

```
  "mcpServers": {
    "my-local-tool": {
      "command": "/path/weaviate_3/weaviate-3.exe",
      "args": [
        ""
      ],
      "env": {
        "hoge": ""
      }
    }
  },
```
***
