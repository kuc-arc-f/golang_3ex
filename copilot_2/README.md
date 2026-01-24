# copilot_2

 Version: 0.9.1

 date    : 2026/01/24

***

GoLang RAG Search , Qdrant database

* GitHub Copilot SDK
* model: gpt-4.1
* embedding: qwen3-embedding:0.6b

***
### Setup

***
* data path: ./data

***
* build

```
go mod init example.com/copilot_2
go mod tidy

go build
```

***
* init, collection add
```
copilot_2.exe init
```

* vector data add
```
copilot_2.exe create
```

***
* RAG search
```
copilot_2.exe search hello
```


***
