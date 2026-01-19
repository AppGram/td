package model

type Task struct {
	ID        int64    `json:"id"`
	ParentID  *int64   `json:"parent_id"`
	Workspace int64    `json:"workspace"`
	Title     string   `json:"title"`
	Completed bool     `json:"completed"`
	Tags      []string `json:"tags"`
	DueDate   string   `json:"due_date"`
	Priority  int      `json:"priority"`
	Order     int      `json:"order"`
	CreatedAt string   `json:"created_at"`
	Children  []*Task  `json:"-"`
}

type Workspace struct {
	ID             int64  `json:"id"`
	Name           string `json:"name"`
	Order          int    `json:"order"`
	TaskCount      int    `json:"task_count"`
	CompletedCount int    `json:"completed_count"`
}

type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeSearch
)

type Pane int

const (
	PaneTasks Pane = iota
	PaneWorkspaces
)

type UIState struct {
	Mode          Mode
	ActivePane    Pane
	SelectedWS    int
	SelectedTask  int
	ExpandedTasks map[int64]bool
	CommandBuf    string
	SearchBuf     string
	SearchQuery   string
	Msg           string
	MsgTimeout    int
}

type Layout struct {
	HeaderH  int
	SidebarW int
	StatusH  int
}
