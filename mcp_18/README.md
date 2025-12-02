# mcp_18

 Version: 0.9.1

 date    : 2025/12/01

***

GoLang MCP Server , RAG Search 

* pgvector use
* embedding: qwen3-embedding:0.6b , ollama

***
* table: ./table.sql

```
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE documents (
  id SERIAL PRIMARY KEY,
  content TEXT NOT NULL,
  embedding vector(1024)
);
```

***
* build

```
go mod init example.com/go-mcp-server-18
go mod tidy

go build
```


***
* test-code

```js
import { spawn } from "child_process";

class RpcClient {
  constructor(command) {
    this.proc = spawn(command);
    this.idCounter = 1;
    this.pending = new Map();

    this.proc.stdout.setEncoding("utf8");
    this.proc.stdout.on("data", (data) => this._handleData(data));
    this.proc.stderr.on("data", (err) => console.error("Rust stderr:", err.toString()));
    this.proc.on("exit", (code) => console.log(`Rust server exited (${code})`));
  }

  _handleData(data) {
    // 複数行対応
    data.split("\n").forEach((line) => {
      console.log("line=", line)
      if (!line.trim()) return;
      try {
        const msg = JSON.parse(line);
        if (msg.id && this.pending.has(msg.id)) {
          const { resolve } = this.pending.get(msg.id);
          this.pending.delete(msg.id);
          resolve(msg.result);
        }
      } catch (e) {
        //console.error("JSON parse error:", e, line);
      }
    });
  }

  call(method, params = {}) {
    const id = this.idCounter++;
    const payload = {
      jsonrpc: "2.0",
      id,
      method,
      params,
    };

    return new Promise((resolve, reject) => {
      this.pending.set(id, { resolve, reject });
      this.proc.stdin.write(JSON.stringify(payload) + "\n");
    });
  }

  close() {
    this.proc.kill();
  }
}

// -----------------------------
// 実行例
// -----------------------------
async function main() {
  const client = new RpcClient("/work/go/mcp/mcp_18/go-mcp-server-18.exe");

  const result1 = await client.call(
    "tools/call", 
    { 
      name: "rag_search", 
      arguments: {
        query: "二十四節季", 
        pg_conect_str: "postgres://root:admin@localhost:5432/mydb"
      }, 
    },
  );  
  console.log("add結果:", result1);

  client.close();
}

main().catch(console.error);

```

***
