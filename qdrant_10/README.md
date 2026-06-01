# qdrant_10

 Version: 0.9.1

 date    : 2026/06/01

***

GoLang RAG Search , Qdrant

* Qdrant database
* embedding: gemini-embedding-001 
* model: Gemma-4-E2B
* llama-server , llama.cpp

***
### Setup

* llama-server start
* port 8090: gemma-4-E2B

```
#gemma-4-E2B

/usr/local/llama-b8642/llama-server -m /var/lm_data/unsloth/gemma-4-E2B-it-Q4_K_S.gguf \
 --chat-template-kwargs '{"enable_thinking": false}' --port 8090 

```
***
### .env

```
GEMINI_API_KEY=
```

***
### related

https://huggingface.co/unsloth/gemma-4-E2B-it-GGUF

***
* data path: ./data

***
* build

```
go mod init example.com/qdrant-10
go mod tidy

go build
```

***
* init, collection add
```
qdrant-10.exe init

```

* vector data add
```
qdrant-10.exe create
```

***
* RAG search

```
qdrant-10.exe search hello
```

***
### blog

https://zenn.dev/knaka0209/scraps/552edd046576b9

***
