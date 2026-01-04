# qdrant_6

 Version: 0.9.1

 date    : 2026/01/03
 
***

GoLang MCP Server , RAG Search 

* Agent Skill use
* Qdrant Database
* embedding: gemini-embedding-001
* GEMINI API

***
* vector data add

https://github.com/kuc-arc-f/golang_3pri/tree/main/qdrant_3

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
go mod init example.com/go-qdrant-6
go mod tidy

go build
```

***
* test-code (JS)

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
    data.split("\n").forEach((line) => {
      console.log("line=", line);
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
  const client = new RpcClient("/work/go/mcp/qdrant_6/go-qdrant-6.exe");

  const result1 = await client.call(
    "tools/call", 
    { 
      name: "rag_search", 
      arguments: {query: "二十四節気"}, 
    },
  );
  console.log("add結果:", result1);

  client.close();
}

main().catch(console.error);

```
