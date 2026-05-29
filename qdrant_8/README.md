# qdrant_8

 Version: 0.9.1

 date    : 2026/05/28

***

GoLang RAG Search , Qdrant

* Qdrant database
* embedding: Qwen3-Embedding-0.6B-Q8_0.gguf
* model: Gemma-4-E2B
* llama-server , llama.cpp

***
### Setup

* llama-server start
* port 8080: Qwen3-Embedding-0.6B
* port 8090: gemma-4-E2B

```
#Qwen3-Embedding-0.6B

/home/user123/llama-server -m /var/lm_data/Qwen3-Embedding-0.6B-Q8_0.gguf --embedding  -c 1024 --port 8080

#gemma-4-E2B

/usr/local/llama-b8642/llama-server -m /var/lm_data/unsloth/gemma-4-E2B-it-Q4_K_S.gguf \
 --chat-template-kwargs '{"enable_thinking": false}' --port 8090 

```

***
### related

https://huggingface.co/unsloth/gemma-4-E2B-it-GGUF

https://huggingface.co/Qwen/Qwen3-Embedding-0.6B-GGUF

***
* data path: ./data

***
* build

```
go mod init example.com/qdrant-8
go mod tidy

go build
```

***
* init, collection add
```
qdrant-8.exe init
```

* vector data add
```
qdrant-8.exe create
```

***
* RAG search

```
qdrant-8.exe search
```


***
