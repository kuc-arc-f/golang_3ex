# chroma_2

 Version: 0.9.1

 date    : 2025/12/28

***

GoLang RAG Search , MCP Server + ChromaDB

* embedding: gemini-embedding-001
* model: gemma-3-27b
* GEMINI-CLI use

***
* vector data add

https://github.com/kuc-arc-f/golang_3ex/tree/main/mcp_25

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
go mod init example.com/chroma_2
go mod tidy

go build
```

***
* settings.json , GEMINI-CLI

```
    "my-local-tool": {
      "command": "/path/chroma_2/chroma_2.exe",
      "args": [
        ""
      ],
      "env": {
        "hoge": ""
      }
    }
```


***
