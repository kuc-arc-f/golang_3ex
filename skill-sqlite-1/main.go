package main

import (
	"bufio"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

const DB_PATH = "/home/naka/work/go/extra/skill-sqlite-1/todos.db"
/**
*
* @param
*
* @return
*/
func main() {
	db, err := sql.Open("sqlite3", DB_PATH)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	createTable(db)
	fmt.Println("全引数:", os.Args)
	fmt.Println("全引数.count:", len(os.Args))

    if len(os.Args) > 2 {
        fmt.Println("最初の引数:", os.Args[1])
        if os.Args[1] == "todo_add" && len(os.Args) <= 3 {
            //fmt.Println("error: argment,")
            var input = os.Args[2]; 
            fmt.Println("input:", input)
            addTodo(db, input)
        }        
        if os.Args[1] == "todo_delete" && len(os.Args) <= 3 {
            var input = os.Args[2];
            fmt.Println("input:", input)
            var id int
            _, err = fmt.Sscanf(input, "%d", &id)
            if err != nil {
                fmt.Println("無効なIDです")
                return
            }            
            //fmt.Println("id:", id)
            deleteTodo(db , id)
        }
        return
    }
    //todos
    if len(os.Args) > 1 {
        if os.Args[1] == "todos" {
            displayTodos(db)
        }
    }
    return
}

/**
*
* @param
*
* @return
*/
func createTable(db *sql.DB) {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS todos (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		created_at TEXT,
		updated_at TEXT
	);`

	_, err := db.Exec(createTableSQL)
	if err != nil {
		log.Fatal(err)
	}
}
/**
*
* @param
*
* @return
*/
func displayTodos(db *sql.DB) {
	todos, err := getAllTodos(db)
	if err != nil {
		log.Printf("TODOの取得に失敗しました: %v", err)
		return
	}

	if len(todos) == 0 {
		fmt.Println("TODOはありません")
		return
	}

	fmt.Println("\n=== TODO一覧 ===")
	for _, todo := range todos {
		fmt.Printf("ID: %d\n", todo.ID)
		fmt.Printf("タイトル: %s\n", todo.Title)
		fmt.Printf("作成日時: %s\n", todo.CreatedAt)
		//fmt.Printf("更新日時: %s\n", todo.UpdatedAt)
		fmt.Println("---")
	}
}

/**
*
* @param
*
* @return
*/
func addTodo(db *sql.DB, title string) {

	if title == "" {
		fmt.Println("タイトルは必須です")
		return
	}

    var content = ""
	now := time.Now().Format("2006-01-02 15:04:05")

	insertSQL := `INSERT INTO todos (title, content, created_at, updated_at) VALUES (?, ?, ?, ?)`
	_, err := db.Exec(insertSQL, title, content, now, now)
	if err != nil {
		log.Printf("TODOの追加に失敗しました: %v", err)
		return
	}

	fmt.Println("TODOを追加しました！")
}

/**
*
* @param
*
* @return
*/
func getAllTodos(db *sql.DB) ([]Todo, error) {
	query := "SELECT id, title, content, created_at, updated_at FROM todos"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var todos []Todo
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Content, &todo.CreatedAt, &todo.UpdatedAt)
		if err != nil {
			return nil, err
		}
		todos = append(todos, todo)
	}

	return todos, nil
}

/**
*
* @param
*
* @return
*/
func deleteTodo(db *sql.DB, id int) {
	todos, err := getAllTodos(db)
	if err != nil {
		log.Printf("TODOの取得に失敗しました: %v", err)
		return
	}

	if len(todos) == 0 {
		fmt.Println("削除するTODOはありません")
		return
	}

	var exists bool
	for _, todo := range todos {
		if todo.ID == id {
			exists = true
			break
		}
	}

	if !exists {
		fmt.Println("指定されたIDのTODOは存在しません")
		return
	}

	deleteSQL := "DELETE FROM todos WHERE id = ?"
	_, err = db.Exec(deleteSQL, id)
	if err != nil {
		log.Printf("TODOの削除に失敗しました: %v", err)
		return
	}

	fmt.Println("TODOを削除しました！")
}

/**
*
* @param
*
* @return
*/
func updateTodo(db *sql.DB) {
	todos, err := getAllTodos(db)
	if err != nil {
		log.Printf("TODOの取得に失敗しました: %v", err)
		return
	}

	if len(todos) == 0 {
		fmt.Println("更新するTODOはありません")
		return
	}

	fmt.Println("\n=== TODO更新 ===")
	for _, todo := range todos {
		fmt.Printf("ID: %d - %s\n", todo.ID, todo.Title)
	}

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Print("更新するTODOのIDを入力してください: ")
	scanner.Scan()
	idStr := strings.TrimSpace(scanner.Text())

	var id int
	_, err = fmt.Sscanf(idStr, "%d", &id)
	if err != nil {
		fmt.Println("無効なIDです")
		return
	}

	var selectedTodo *Todo
	for _, todo := range todos {
		if todo.ID == id {
			selectedTodo = &todo
			break
		}
	}

	if selectedTodo == nil {
		fmt.Println("指定されたIDのTODOは存在しません")
		return
	}

	fmt.Printf("現在のタイトル: %s\n", selectedTodo.Title)
	fmt.Print("新しいタイトルを入力してください（空白で変更なし）: ")
	scanner.Scan()
	newTitle := strings.TrimSpace(scanner.Text())
	if newTitle == "" {
		newTitle = selectedTodo.Title
	}

	fmt.Printf("現在の内容: %s\n", selectedTodo.Content)
	fmt.Print("新しい内容を入力してください（空白で変更なし）: ")
	scanner.Scan()
	newContent := strings.TrimSpace(scanner.Text())
	if newContent == "" {
		newContent = selectedTodo.Content
	}

	now := time.Now().Format("2006-01-02 15:04:05")
	updateSQL := "UPDATE todos SET title = ?, content = ?, updated_at = ? WHERE id = ?"
	_, err = db.Exec(updateSQL, newTitle, newContent, now, id)
	if err != nil {
		log.Printf("TODOの更新に失敗しました: %v", err)
		return
	}

	fmt.Println("TODOを更新しました！")
}
