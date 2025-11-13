# mcp_15

 Version: 0.9.1

 date    : 2025/11/11

***

GoLang RAG , Vector search

* Postgres use
* model-embed: gemini-embedding-001
* model: gemini-2.0-flash
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
go mod init example.com/go-mcp-server-15
go mod tidy

go build
go run .
```

***
* test-code: test_search.js

```js
import { spawn } from "child_process";

class RpcClient {
  constructor(command) {
    this.proc = spawn(command);
    this.idCounter = 1;
    this.pending = new Map();

    this.proc.stdout.setEncoding("utf8");
    this.proc.stdout.on("data", (data) => this._handleData(data));
    this.proc.stderr.on("data", (err) => console.error("MCP stderr:", err.toString()));
    this.proc.on("exit", (code) => console.log(`MCP server exited (${code})`));
  }

  _handleData(data) {
    // 複数行対応
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
  const client = new RpcClient("/home/naka/work/go/mcp/mcp_15/go-mcp-server-15");

  const result1 = await client.call(
    "tools/call", 
    { 
      name: "rag_search", 
      arguments: {input_text: "縄文時代"}, 
    },
  );
  console.log("add結果:", result1);

  client.close();
}

main().catch(console.error);

```
***