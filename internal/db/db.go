package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"

	"github.com/appgram/td/internal/model"
)

type Workspace struct {
	ID             int64
	Name           string
	Order          int
	TaskCount      int
	CompletedCount int
}

type DB struct {
	*sql.DB
}

func NewDB() (*DB, error) {
	home := os.Getenv("HOME")
	if home == "" {
		home = os.Getenv("USERPROFILE")
	}
	dbPath := fmt.Sprintf("%s/.config/td/td.db", home)

	dir := fmt.Sprintf("%s/.config/td", home)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create config dir: %v", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	if err := initSchema(db); err != nil {
		return nil, fmt.Errorf("failed to init schema: %v", err)
	}

	return &DB{DB: db}, nil
}

func initSchema(db *sql.DB) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS workspaces (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			word_order INTEGER NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			parent_id INTEGER,
			workspace_id INTEGER NOT NULL,
			title TEXT NOT NULL,
			completed INTEGER NOT NULL DEFAULT 0,
			tags TEXT,
			due_date TEXT,
			priority INTEGER DEFAULT 0,
			task_order INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (parent_id) REFERENCES tasks(id) ON DELETE CASCADE,
			FOREIGN KEY (workspace_id) REFERENCES workspaces(id) ON DELETE CASCADE
		)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_workspace ON tasks(workspace_id)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_parent ON tasks(parent_id)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			return fmt.Errorf("failed to execute: %v", err)
		}
	}

	return nil
}

func (db *DB) GetSetting(key string) (string, error) {
	var value sql.NullString
	if err := db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&value); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	if value.Valid {
		return value.String, nil
	}
	return "", nil
}

func (db *DB) SetSetting(key, value string) error {
	_, err := db.Exec(`
		INSERT INTO settings (key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, key, value)
	return err
}

func (db *DB) GetWorkspaces() ([]Workspace, error) {
	rows, err := db.Query(`
		SELECT w.id, w.name, w.word_order,
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = w.id) as task_count,
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = w.id AND completed = 1) as completed_count
		FROM workspaces w ORDER BY w.word_order
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ws []Workspace
	for rows.Next() {
		var w Workspace
		rows.Scan(&w.ID, &w.Name, &w.Order, &w.TaskCount, &w.CompletedCount)
		ws = append(ws, w)
	}
	return ws, nil
}

func (db *DB) CreateWorkspace(name string) (int64, error) {
	var order int
	db.QueryRow("SELECT COALESCE(MAX(word_order), -1) + 1 FROM workspaces").Scan(&order)
	result, err := db.Exec("INSERT INTO workspaces (name, word_order) VALUES (?, ?)", name, order)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (db *DB) DeleteWorkspace(id int64) error {
	_, err := db.Exec("DELETE FROM workspaces WHERE id = ?", id)
	return err
}

func (db *DB) RenameWorkspace(id int64, name string) error {
	_, err := db.Exec("UPDATE workspaces SET name = ? WHERE id = ?", name, id)
	return err
}

func (db *DB) GetTasksForWorkspace(workspaceID int64) ([]*model.Task, error) {
	rows, err := db.Query(`
		SELECT id, parent_id, title, completed,
			   COALESCE(tags, ''), COALESCE(due_date, ''), priority, task_order, created_at
		FROM tasks WHERE workspace_id = ? ORDER BY parent_id, task_order
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := make(map[int64]*model.Task)
	var roots []*model.Task

	for rows.Next() {
		var t model.Task
		var parentID sql.NullInt64
		var tags, dueDate sql.NullString
		rows.Scan(&t.ID, &parentID, &t.Title, &t.Completed, &tags, &dueDate, &t.Priority, &t.Order, &t.CreatedAt)
		t.Workspace = workspaceID
		if parentID.Valid {
			t.ParentID = &parentID.Int64
		}
		if tags.Valid {
			t.Tags = splitTags(tags.String)
		}
		if dueDate.Valid {
			t.DueDate = dueDate.String
		}
		tasks[t.ID] = &t
	}

	for _, t := range tasks {
		if t.ParentID == nil {
			roots = append(roots, t)
		} else {
			if parent, ok := tasks[*t.ParentID]; ok {
				parent.Children = append(parent.Children, t)
			}
		}
	}

	return roots, nil
}

func splitTags(s string) []string {
	if s == "" {
		return nil
	}
	var tags []string
	for _, t := range splitString(s, ",") {
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func joinTags(tags []string) string {
	if len(tags) == 0 {
		return ""
	}
	result := tags[0]
	for i := 1; i < len(tags); i++ {
		result += "," + tags[i]
	}
	return result
}

func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] {
			result = append(result, s[start:i])
			start = i + 1
		}
	}
	result = append(result, s[start:])
	return result
}

func (db *DB) AddTask(workspaceID int64, title string, parentID *int64) (int64, error) {
	return db.AddTaskWithMeta(workspaceID, title, parentID, nil, "", 0)
}

func (db *DB) AddTaskWithMeta(workspaceID int64, title string, parentID *int64, tags []string, dueDate string, priority int) (int64, error) {
	var order int
	db.QueryRow("SELECT COALESCE(MAX(task_order), -1) + 1 FROM tasks WHERE workspace_id = ? AND (parent_id = ? OR (parent_id IS NULL AND ? IS NULL))",
		workspaceID, coalesceNull(parentID), coalesceNull(parentID)).Scan(&order)

	result, err := db.Exec("INSERT INTO tasks (workspace_id, parent_id, title, task_order, tags, due_date, priority) VALUES (?, ?, ?, ?, ?, ?, ?)",
		workspaceID, coalesceNull(parentID), title, order, joinTags(tags), nullIfEmpty(dueDate), priority)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func nullIfEmpty(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func (db *DB) UpdateTask(task *model.Task) error {
	_, err := db.Exec(`UPDATE tasks SET title = ?, completed = ?, tags = ?, due_date = ?, priority = ?
		WHERE id = ?`, task.Title, boolToInt(task.Completed), joinTags(task.Tags),
		task.DueDate, task.Priority, task.ID)
	return err
}

func (db *DB) DeleteTask(id int64) error {
	_, err := db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

func (db *DB) ToggleTask(id int64) error {
	_, err := db.Exec("UPDATE tasks SET completed = NOT completed WHERE id = ?", id)
	return err
}

func (db *DB) SetTaskCompleted(id int64, completed bool) error {
	_, err := db.Exec("UPDATE tasks SET completed = ? WHERE id = ?", boolToInt(completed), id)
	return err
}

func (db *DB) MoveTask(id int64, newParentID *int64) error {
	var workspaceID int64
	if err := db.QueryRow("SELECT workspace_id FROM tasks WHERE id = ?", id).Scan(&workspaceID); err != nil {
		return err
	}

	var order int
	if err := db.QueryRow(
		"SELECT COALESCE(MAX(task_order), -1) + 1 FROM tasks WHERE workspace_id = ? AND (parent_id = ? OR (parent_id IS NULL AND ? IS NULL))",
		workspaceID, coalesceNull(newParentID), coalesceNull(newParentID),
	).Scan(&order); err != nil {
		return err
	}

	_, err := db.Exec("UPDATE tasks SET parent_id = ?, task_order = ? WHERE id = ?", coalesceNull(newParentID), order, id)
	return err
}

func (db *DB) GetTaskStats(workspaceID int64) (total, completed, blocked int, err error) {
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE workspace_id = ?", workspaceID).Scan(&total)
	if err != nil {
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE workspace_id = ? AND completed = 1", workspaceID).Scan(&completed)
	if err != nil {
		return
	}
	err = db.QueryRow("SELECT COUNT(*) FROM tasks WHERE workspace_id = ? AND priority = -1", workspaceID).Scan(&blocked)
	return
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func coalesceNull(p *int64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
