# skill-sqlite-1

 Version: 0.9.1

 date    : 2026/03/10

***

GoLang  , todo app CLI

* Agent Skills
* GitHub Copilot SDK

***
### Setup

* DB-Path change
* main.go

```
const DB_PATH = "/home/user123/skill-sqlite-1/todos.db"
```

***
* build

```
go mod init example.com/skill-sqlite-1
go mod tidy

go build
```

***
* add
```
./skill-sqlite-1 todo_add todo-test1
```

* list
```
./skill-sqlite-1 todos
```

* delete
```
./skill-sqlite-1 todo_delete 1
```

***


***
