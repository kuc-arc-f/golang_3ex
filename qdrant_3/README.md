# qdrant_3

 Version: 0.9.1

 date    : 2025/12/27

***

GoLang RAG Search , Qdrant

* embedding: gemini-embedding-001
* model: gemma-3-27b

***
### Setup

* .env
```
GEMINI_API_KEY=your-key
```
***
* data path: ./data

***
* build

```
go mod init example.com/qdrant-3
go mod tidy

go build
```

***
* init, collection add
```
qdrant-3.exe init
```

* vector data add
```
qdrant-3.exe create
```

***
* RAG search
```
qdrant-3.exe search
```

***
### blog

https://zenn.dev/knaka0209/scraps/ef3b5fc2f6f916

***
