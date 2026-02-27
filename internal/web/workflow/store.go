// Package workflow 提供基于 SQLite 的设备管理工作流存储。
package workflow

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// StepType 定义工作流步骤类型。
type StepType string

const (
	StepTypeTTS     StepType = "tts"
	StepTypePlayURL StepType = "play_url"
	StepTypeMiIO    StepType = "miio"
	StepTypeDelay   StepType = "delay"
)

// Step 描述工作流中的一个动作。
type Step struct {
	Type       StepType `json:"type"`
	Label      string   `json:"label,omitempty"`
	Device     string   `json:"device,omitempty"`
	Text       string   `json:"text,omitempty"`
	URL        string   `json:"url,omitempty"`
	MiIOText   string   `json:"miio_text,omitempty"`
	DurationMS int      `json:"duration_ms,omitempty"`
}

// Workflow 是设备管理工作流。
type Workflow struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	Steps       []Step    `json:"steps"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Store 将工作流持久化到 SQLite。
type Store struct {
	mu sync.RWMutex
	db *sql.DB
}

// NewStore 创建工作流存储。dataDir 为 miflow.db 所在目录。
func NewStore(dataDir string) (*Store, error) {
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}
	dbPath := filepath.Join(dataDir, "miflow.db")
	db, err := sql.Open("sqlite", dbPath+"?_journal_mode=WAL")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// migrate 执行数据库迁移，创建 workflows 表。
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			steps_json TEXT,
			created_at TEXT NOT NULL,
			updated_at TEXT NOT NULL
		)
	`)
	return err
}

// List 返回所有工作流。
func (s *Store) List() ([]Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rows, err := s.db.Query(`SELECT id, name, description, steps_json, created_at, updated_at FROM workflows ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Workflow
	for rows.Next() {
		var w Workflow
		var stepsJSON, createdAt, updatedAt string
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &stepsJSON, &createdAt, &updatedAt); err != nil {
			return nil, err
		}
		if stepsJSON != "" {
			_ = json.Unmarshal([]byte(stepsJSON), &w.Steps)
		}
		w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
		out = append(out, w)
	}
	return out, rows.Err()
}

// Get 按 ID 返回工作流。
func (s *Store) Get(id string) (*Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var w Workflow
	var stepsJSON, createdAt, updatedAt string
	err := s.db.QueryRow(`SELECT id, name, description, steps_json, created_at, updated_at FROM workflows WHERE id = ?`, id).
		Scan(&w.ID, &w.Name, &w.Description, &stepsJSON, &createdAt, &updatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if stepsJSON != "" {
		_ = json.Unmarshal([]byte(stepsJSON), &w.Steps)
	}
	w.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	w.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)
	return &w, nil
}

// Upsert 创建或更新工作流。
func (s *Store) Upsert(w *Workflow) error {
	if w == nil {
		return fmt.Errorf("workflow is nil")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	if w.ID == "" {
		w.ID = fmt.Sprintf("%d-%s", now.UnixNano(), sanitizeID(w.Name))
	}
	w.UpdatedAt = now
	if w.CreatedAt.IsZero() {
		w.CreatedAt = now
	}

	stepsJSON, _ := json.Marshal(w.Steps)
	_, err := s.db.Exec(`
		INSERT INTO workflows (id, name, description, steps_json, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			name = excluded.name,
			description = excluded.description,
			steps_json = excluded.steps_json,
			updated_at = excluded.updated_at
	`, w.ID, w.Name, w.Description, string(stepsJSON), w.CreatedAt.Format(time.RFC3339), w.UpdatedAt.Format(time.RFC3339))
	return err
}

// Delete 删除工作流。
func (s *Store) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, err := s.db.Exec(`DELETE FROM workflows WHERE id = ?`, id)
	return err
}

// sanitizeID 将名称规范化为 ID 格式（小写、连字符，最长 32 字符）。
func sanitizeID(s string) string {
	var out []rune
	for _, r := range s {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			out = append(out, r)
		} else if r >= 'A' && r <= 'Z' {
			out = append(out, r+32)
		} else if r == ' ' || r == '-' {
			if len(out) > 0 && out[len(out)-1] != '-' {
				out = append(out, '-')
			}
		}
	}
	for len(out) > 0 && out[len(out)-1] == '-' {
		out = out[:len(out)-1]
	}
	if len(out) == 0 {
		return "workflow"
	}
	result := string(out)
	if len(result) > 32 {
		return result[:32]
	}
	return result
}
