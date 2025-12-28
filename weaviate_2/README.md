# weaviate_2

 Version: 0.9.1

 date    : 2025/12/27

***

GoLang RAG Search , Weaviate

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
go mod init example.com/weaviate-2
go mod tidy

go build
```

***
* init, collection add
```
weaviate-2.exe init
```

* vector data add
```
weaviate-2.exe create
```

***
* RAG search
```
weaviate-2.exe search
```

***
