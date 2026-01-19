package tui

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/appgram/td/internal/db"
	"github.com/appgram/td/internal/model"
)

var (
	bgColor     = lipgloss.Color("#0b1116")
	sidebarBg   = lipgloss.Color("#0d151c")
	selectionBg = lipgloss.Color("#1a2630")
	cursorBg    = lipgloss.Color("#22313b")
	accent      = lipgloss.Color("#6fb1a5")
	dim         = lipgloss.Color("#5a6670")
	textColor   = lipgloss.Color("#a5adb7")
	headerColor = lipgloss.Color("#7fb7b0")
	borderColor = lipgloss.Color("#172029")
	doneColor   = lipgloss.Color("#3e4852")
	infoBg      = lipgloss.Color("#2a2a2a")
)

const weatherUnknown = "--°"

// ParsedTask holds the result of parsing inline task syntax
type ParsedTask struct {
	Title    string
	Tags     []string
	DueDate  string
	Priority int
}

// ParseTaskInput parses inline task syntax: "task #tag @date !priority"
func ParseTaskInput(input string) ParsedTask {
	var result ParsedTask
	var titleParts []string

	words := strings.Fields(input)
	for _, word := range words {
		switch {
		case strings.HasPrefix(word, "#"):
			tag := strings.TrimPrefix(word, "#")
			if tag != "" {
				result.Tags = append(result.Tags, tag)
			}
		case strings.HasPrefix(word, "!"):
			p := strings.ToLower(strings.TrimPrefix(word, "!"))
			switch p {
			case "high", "h":
				result.Priority = 2
			case "low", "l":
				result.Priority = 1
			case "blocked", "b":
				result.Priority = -1
			case "normal", "n":
				result.Priority = 0
			}
		case strings.HasPrefix(word, "@"):
			date := strings.TrimPrefix(word, "@")
			result.DueDate = parseDueDate(date)
		default:
			titleParts = append(titleParts, word)
		}
	}

	result.Title = strings.Join(titleParts, " ")
	return result
}

func parseDueDate(input string) string {
	input = strings.ToLower(input)
	now := time.Now()

	switch input {
	case "today":
		return now.Format("2006-01-02")
	case "tomorrow", "tmr":
		return now.AddDate(0, 0, 1).Format("2006-01-02")
	case "week", "nextweek":
		return now.AddDate(0, 0, 7).Format("2006-01-02")
	case "mon", "monday":
		return nextWeekday(now, time.Monday)
	case "tue", "tuesday":
		return nextWeekday(now, time.Tuesday)
	case "wed", "wednesday":
		return nextWeekday(now, time.Wednesday)
	case "thu", "thursday":
		return nextWeekday(now, time.Thursday)
	case "fri", "friday":
		return nextWeekday(now, time.Friday)
	case "sat", "saturday":
		return nextWeekday(now, time.Saturday)
	case "sun", "sunday":
		return nextWeekday(now, time.Sunday)
	default:
		// Try parsing as date (YYYY-MM-DD or MM-DD)
		if t, err := time.Parse("2006-01-02", input); err == nil {
			return t.Format("2006-01-02")
		}
		if t, err := time.Parse("01-02", input); err == nil {
			return time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, now.Location()).Format("2006-01-02")
		}
		return input
	}
}

func nextWeekday(from time.Time, weekday time.Weekday) string {
	daysUntil := int(weekday) - int(from.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	return from.AddDate(0, 0, daysUntil).Format("2006-01-02")
}

type colorScheme struct {
	name        string
	bg          lipgloss.Color
	sidebar     lipgloss.Color
	selection   lipgloss.Color
	cursor      lipgloss.Color
	accent      lipgloss.Color
	dim         lipgloss.Color
	text        lipgloss.Color
	header      lipgloss.Color
	border      lipgloss.Color
	done        lipgloss.Color
	info        lipgloss.Color
}

var schemes = []colorScheme{
	{
		name:      "black",
		bg:        lipgloss.Color("#0a0a0a"),
		sidebar:   lipgloss.Color("#0d0d0d"),
		selection: lipgloss.Color("#1a1a1a"),
		cursor:    lipgloss.Color("#242424"),
		accent:    lipgloss.Color("#9aa3a8"),
		dim:       lipgloss.Color("#5b5f63"),
		text:      lipgloss.Color("#c0c5c8"),
		header:    lipgloss.Color("#b5babf"),
		border:    lipgloss.Color("#151515"),
		done:      lipgloss.Color("#3b3b3b"),
		info:      lipgloss.Color("#2a2a2a"),
	},
	{
		name:      "copper",
		bg:        lipgloss.Color("#11110f"),
		sidebar:   lipgloss.Color("#14120f"),
		selection: lipgloss.Color("#2a1f16"),
		cursor:    lipgloss.Color("#3a2c20"),
		accent:    lipgloss.Color("#c58b5a"),
		dim:       lipgloss.Color("#6f6256"),
		text:      lipgloss.Color("#b9ab9d"),
		header:    lipgloss.Color("#d3a57a"),
		border:    lipgloss.Color("#1f1a14"),
		done:      lipgloss.Color("#4d3f33"),
		info:      lipgloss.Color("#3a2b20"),
	},
	{
		name:      "seafoam",
		bg:        lipgloss.Color("#0a1214"),
		sidebar:   lipgloss.Color("#0b1518"),
		selection: lipgloss.Color("#16262b"),
		cursor:    lipgloss.Color("#20343a"),
		accent:    lipgloss.Color("#70c0b6"),
		dim:       lipgloss.Color("#5d7274"),
		text:      lipgloss.Color("#a7b6b6"),
		header:    lipgloss.Color("#88c9c0"),
		border:    lipgloss.Color("#162126"),
		done:      lipgloss.Color("#3f4d52"),
		info:      lipgloss.Color("#203237"),
	},
	{
		name:      "forest",
		bg:        lipgloss.Color("#0c120f"),
		sidebar:   lipgloss.Color("#0e1512"),
		selection: lipgloss.Color("#1a251f"),
		cursor:    lipgloss.Color("#243126"),
		accent:    lipgloss.Color("#7fa879"),
		dim:       lipgloss.Color("#5f6d60"),
		text:      lipgloss.Color("#aab3a7"),
		header:    lipgloss.Color("#93b98c"),
		border:    lipgloss.Color("#172019"),
		done:      lipgloss.Color("#3f4b41"),
		info:      lipgloss.Color("#27352b"),
	},
	{
		name:      "slate",
		bg:        lipgloss.Color("#0b0f14"),
		sidebar:   lipgloss.Color("#0e131a"),
		selection: lipgloss.Color("#1a2430"),
		cursor:    lipgloss.Color("#243242"),
		accent:    lipgloss.Color("#87a2c2"),
		dim:       lipgloss.Color("#5a6676"),
		text:      lipgloss.Color("#a4afbd"),
		header:    lipgloss.Color("#98b4d1"),
		border:    lipgloss.Color("#16202a"),
		done:      lipgloss.Color("#3c4654"),
		info:      lipgloss.Color("#243244"),
	},
}

type TaskLine struct {
	Task     *model.Task
	Depth    int
	Expanded bool
	Index    int
}

type App struct {
	db           *db.DB
	workspaces   []db.Workspace
	tasks        []*model.Task
	flatTasks    []TaskLine
	taskIndex    map[int64]*model.Task
	state        model.UIState
	width        int
	height       int
	inputBuf     string
	taskInputBuf string
	newTaskParent *int64
	editingTaskID *int64
	pendingKey    string
	pendingAt     time.Time
	headerCache   string
	headerDate    string
	headerWidth   int
	quitRequested bool
	commandLeader string
	asciiArt      string
	asciiNames    []string
	asciiArts     []string
	asciiLines    []string
	showAsciiList bool
	asciiScroll   int
	showTaskInfo  bool
	taskScroll    int
	showHelp      bool
	showDashboard bool
	weatherEnabled bool
	weatherCity    string
	weatherLat     float64
	weatherLon     float64
	weatherTemp    string
	weatherChecked time.Time
	weatherUnit    string
}

func New(database *db.DB) *App {
	app := &App{
		db: database,
		state: model.UIState{
			Mode:          model.ModeNormal,
			ActivePane:    model.PaneTasks,
			ExpandedTasks: make(map[int64]bool),
		},
		commandLeader: ":",
		weatherTemp:   weatherUnknown,
		weatherUnit:   "f",
		showDashboard: true,
	}
	app.applyScheme("black")
	return app
}

func (a *App) Init() tea.Cmd {
	return tickCmd()
}

func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		a.handleKey(msg)
		if a.quitRequested {
			a.quitRequested = false
			return a, tea.Quit
		}
	case tickMsg:
		a.clearPendingKey()
		a.tickMessage()
		if cmd := a.maybeFetchWeather(); cmd != nil {
			return a, cmd
		}
		return a, tickCmd()
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
	case weatherMsg:
		if msg.err != nil {
			a.setMessage("weather unavailable")
		} else {
			a.weatherTemp = msg.temp
			a.weatherChecked = time.Now()
			if msg.lat != 0 || msg.lon != 0 {
				a.weatherLat = msg.lat
				a.weatherLon = msg.lon
				_ = a.db.SetSetting("weather_lat", fmt.Sprintf("%0.4f", a.weatherLat))
				_ = a.db.SetSetting("weather_lon", fmt.Sprintf("%0.4f", a.weatherLon))
			}
		}
	}
	return a, nil
}

func (a *App) handleKey(msg tea.KeyMsg) {
	switch a.state.Mode {
	case model.ModeNormal:
		a.handleNormalMode(msg)
	case model.ModeInsert:
		a.handleInsertMode(msg)
	case model.ModeCommand:
		a.handleCommandMode(msg)
	case model.ModeSearch:
		a.handleSearchMode(msg)
	}
}

func (a *App) handleNormalMode(msg tea.KeyMsg) {
	if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == ' ' {
		if a.state.ActivePane == model.PaneTasks {
			a.toggleTask()
		}
		return
	}

	switch msg.String() {
	case "ctrl+c", "q":
		a.quitRequested = true
	case "tab":
		a.togglePane()
	case ":":
		a.openCommand(":")
	case "/":
		a.openCommand("/")
	case "?":
		a.state.Mode = model.ModeSearch
		a.state.SearchBuf = ""
	case "esc":
		if a.showAsciiList {
			a.showAsciiList = false
		}
		if a.showHelp {
			a.showHelp = false
		}
	case "i":
		if a.state.ActivePane == model.PaneTasks {
			a.editTask()
		}
	case "a":
		if a.state.ActivePane == model.PaneTasks {
			a.addTask()
		}
	case "m":
		if a.state.ActivePane == model.PaneTasks {
			a.toggleTaskInfo()
		}
	case "W":
		a.openCommandWithBuffer(":", "ws add ")
	case "R":
		a.openCommandWithBuffer(":", "ws rename ")
	case "X":
		a.openCommandWithBuffer(":", "ws delete")
	case "H":
		a.showHelp = !a.showHelp
	case "x", " ", "space":
		if a.state.ActivePane == model.PaneTasks {
			a.toggleTask()
		}
	case "h":
		if a.state.ActivePane == model.PaneTasks {
			a.collapseTask()
		}
	case "l":
		if a.state.ActivePane == model.PaneTasks {
			a.expandTask()
		}
	case "j", "down":
		if a.state.ActivePane == model.PaneWorkspaces {
			a.moveWorkspace(1)
		} else if a.showAsciiList {
			a.scrollAscii(1)
		} else {
			a.moveCursor(1)
		}
	case "k", "up":
		if a.state.ActivePane == model.PaneWorkspaces {
			a.moveWorkspace(-1)
		} else if a.showAsciiList {
			a.scrollAscii(-1)
		} else {
			a.moveCursor(-1)
		}
	case "G":
		if a.state.ActivePane == model.PaneTasks && a.showAsciiList {
			a.scrollAsciiToEnd()
		} else if a.state.ActivePane == model.PaneTasks {
			a.moveToBottom()
		}
	case "gg":
		if a.state.ActivePane == model.PaneTasks && a.showAsciiList {
			a.scrollAsciiToTop()
		} else if a.state.ActivePane == model.PaneTasks {
			a.moveToTop()
		}
	case "pgdown":
		if a.state.ActivePane == model.PaneTasks && a.showAsciiList {
			a.scrollAscii(a.asciiPageStep())
		}
	case "pgup":
		if a.state.ActivePane == model.PaneTasks && a.showAsciiList {
			a.scrollAscii(-a.asciiPageStep())
		}
	case "enter":
		if a.state.ActivePane == model.PaneWorkspaces {
			a.selectWorkspace(a.state.SelectedWS)
			a.state.ActivePane = model.PaneTasks
		} else {
			a.toggleTask()
		}
	case ">":
		if a.state.ActivePane == model.PaneTasks {
			a.indentTask()
		}
	case "<", "shift+tab":
		if a.state.ActivePane == model.PaneTasks {
			a.unindentTask()
		}
	default:
		if msg.Type == tea.KeySpace && a.state.ActivePane == model.PaneTasks {
			a.toggleTask()
			return
		}
		a.handlePendingKey(msg.String())
	}
}

func (a *App) handleInsertMode(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		a.state.Mode = model.ModeNormal
		a.taskInputBuf = ""
		a.newTaskParent = nil
		a.editingTaskID = nil
	case "enter":
		if a.taskInputBuf != "" {
			if a.editingTaskID != nil {
				a.saveTaskEdit()
			} else {
				a.saveNewTask()
			}
		}
	case "backspace":
		if len(a.taskInputBuf) > 0 {
			a.taskInputBuf = a.taskInputBuf[:len(a.taskInputBuf)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			a.taskInputBuf += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			a.taskInputBuf += " "
		}
	}
}

func (a *App) handleCommandMode(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		a.state.Mode = model.ModeNormal
		a.state.CommandBuf = ""
	case "enter":
		a.executeCommand()
	case "backspace":
		if len(a.state.CommandBuf) > 0 {
			a.state.CommandBuf = a.state.CommandBuf[:len(a.state.CommandBuf)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			a.state.CommandBuf += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			a.state.CommandBuf += " "
		}
	}
}

func (a *App) handleSearchMode(msg tea.KeyMsg) {
	switch msg.String() {
	case "esc":
		a.state.Mode = model.ModeNormal
		a.state.SearchBuf = ""
	case "enter":
		a.performSearch()
	case "backspace":
		if len(a.state.SearchBuf) > 0 {
			a.state.SearchBuf = a.state.SearchBuf[:len(a.state.SearchBuf)-1]
		}
	default:
		if msg.Type == tea.KeyRunes {
			a.state.SearchBuf += string(msg.Runes)
		} else if msg.Type == tea.KeySpace {
			a.state.SearchBuf += " "
		}
	}
}

func (a *App) View() string {
	if a.width == 0 {
		return "Loading..."
	}

	header := a.renderHeader()
	headerH := lipgloss.Height(header)
	statusH := a.statusHeight()
	contentH := a.height - headerH - statusH
	if contentH < 0 {
		contentH = 0
	}

	sidebarW := 24
	if sidebarW > a.width/2 {
		sidebarW = a.width / 2
	}

	sidebar := a.renderSidebar(sidebarW, contentH)
	tasks := a.renderTasks(a.width-sidebarW, contentH)
	status := a.renderStatus()

	return lipgloss.JoinVertical(lipgloss.Left,
		header,
		lipgloss.JoinHorizontal(lipgloss.Top, sidebar, tasks),
		status,
	)
}

type dashboardStats struct {
	totalTasks   int
	completed    int
	dueToday     int
	overdue      int
	highPriority int
	blocked      int
	todayTasks   []string
}

func (a *App) getDashboardStats() dashboardStats {
	var stats dashboardStats
	today := time.Now().Format("2006-01-02")

	var countTasks func([]*model.Task)
	countTasks = func(tasks []*model.Task) {
		for _, t := range tasks {
			stats.totalTasks++
			if t.Completed {
				stats.completed++
			} else {
				if t.DueDate == today {
					stats.dueToday++
					if len(stats.todayTasks) < 3 {
						stats.todayTasks = append(stats.todayTasks, t.Title)
					}
				} else if t.DueDate != "" && t.DueDate < today {
					stats.overdue++
				}
				if t.Priority >= 2 {
					stats.highPriority++
				}
				if t.Priority < 0 {
					stats.blocked++
				}
			}
			if len(t.Children) > 0 {
				countTasks(t.Children)
			}
		}
	}
	countTasks(a.tasks)
	return stats
}

func (a *App) renderDashboard() string {
	stats := a.getDashboardStats()

	// Minimal styles using theme colors
	dimStyle := lipgloss.NewStyle().Foreground(dim)
	accentStyle := lipgloss.NewStyle().Foreground(accent)
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e5c07b"))
	dangerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#e06c75"))

	var parts []string

	// Progress
	progress := 0
	if stats.totalTasks > 0 {
		progress = (stats.completed * 100) / stats.totalTasks
	}
	parts = append(parts, dimStyle.Render(fmt.Sprintf("%d/%d", stats.completed, stats.totalTasks))+" "+accentStyle.Render(fmt.Sprintf("%d%%", progress)))

	// Today
	if stats.dueToday > 0 {
		parts = append(parts, warnStyle.Render(fmt.Sprintf("today:%d", stats.dueToday)))
	}

	// Overdue
	if stats.overdue > 0 {
		parts = append(parts, dangerStyle.Render(fmt.Sprintf("overdue:%d", stats.overdue)))
	}

	// High priority
	if stats.highPriority > 0 {
		parts = append(parts, warnStyle.Render(fmt.Sprintf("!high:%d", stats.highPriority)))
	}

	// Blocked
	if stats.blocked > 0 {
		parts = append(parts, dangerStyle.Render(fmt.Sprintf("blocked:%d", stats.blocked)))
	}

	return dimStyle.Render("[ ") + strings.Join(parts, dimStyle.Render(" · ")) + dimStyle.Render(" ]")
}

func (a *App) renderHeader() string {
	today := time.Now().Format("2006-01-02")
	// Don't cache when dashboard is shown (dynamic content)
	if !a.showDashboard && a.headerCache != "" && a.headerDate == today && a.headerWidth == a.width {
		return a.headerCache
	}

	date := time.Now().Format("Monday, January 2, 2006")
	art := a.asciiArt
	if art == "" {
		art = "  .-.\n (o o)\n | O \\\n  \\   \\\n   `~~~'"
	}
	art = clipArt(art, maxWidth(a.width-2))
	quote := "keep the noise outside"

	var content string
	if a.showDashboard {
		dashboard := a.renderDashboard()
		content = lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(dim).Render(quote),
			lipgloss.NewStyle().Foreground(headerColor).Render(art),
			lipgloss.NewStyle().Foreground(textColor).Render(date),
			dashboard,
		)
	} else {
		content = lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(dim).Render(quote),
			lipgloss.NewStyle().Foreground(headerColor).Render(art),
			lipgloss.NewStyle().Foreground(textColor).Render(date),
		)
	}

	header := lipgloss.NewStyle().
		Width(a.width).
		Align(lipgloss.Center).
		Background(bgColor).
		PaddingTop(1).
		PaddingBottom(1).
		Render(content)

	if !a.showDashboard {
		a.headerCache = header
		a.headerDate = today
		a.headerWidth = a.width
	}

	return header
}

func (a *App) renderSidebar(w, h int) string {
	var b strings.Builder
	innerW := w - 2
	if innerW < 0 {
		innerW = 0
	}
	title := "Workspaces"
	if a.state.ActivePane == model.PaneWorkspaces {
		title = lipgloss.NewStyle().Foreground(accent).Render("▌ " + title)
	} else {
		title = lipgloss.NewStyle().Foreground(dim).Render("  " + title)
	}
	b.WriteString(title + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", innerW)) + "\n")

	for i, ws := range a.workspaces {
		prefix := "  "
		if i == a.state.SelectedWS {
			prefix = "»"
		}
		if i == a.state.SelectedWS {
			b.WriteString(lipgloss.NewStyle().
				Background(selectionBg).
				Render(strings.TrimSpace(fmt.Sprintf("%s %s", prefix, ws.Name))))
		} else {
			b.WriteString(strings.TrimSpace(fmt.Sprintf("%s %s", prefix, ws.Name)))
		}
		b.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Background(sidebarBg).
		PaddingLeft(1).
		PaddingRight(1).
		Render(b.String())
}

func (a *App) renderTasks(w, h int) string {
	var b strings.Builder
	innerW := w - 2
	if innerW < 0 {
		innerW = 0
	}
	title := "Todos"
	wsName := ""
	if a.state.SelectedWS < len(a.workspaces) {
		wsName = a.workspaces[a.state.SelectedWS].Name
	}
	if a.state.ActivePane == model.PaneTasks {
		title = lipgloss.NewStyle().Foreground(accent).Render("▌ " + title)
	} else {
		title = lipgloss.NewStyle().Foreground(dim).Render("  " + title)
	}
	if wsName != "" {
		title = fmt.Sprintf("%s  %s", title, lipgloss.NewStyle().Foreground(dim).Render(wsName))
	}
	if a.state.SearchQuery != "" {
		title = fmt.Sprintf("%s  %s", title, lipgloss.NewStyle().Foreground(dim).Render("[filter: "+a.state.SearchQuery+"]"))
	}

	b.WriteString(title + "\n")
	b.WriteString(lipgloss.NewStyle().Foreground(borderColor).Render(strings.Repeat("─", innerW)) + "\n")

	if a.showAsciiList {
		list := a.renderAsciiList(innerW, h-2)
		b.WriteString(list)
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Background(bgColor).
			PaddingLeft(1).
			PaddingRight(1).
			Render(b.String())
	}
	if a.showHelp {
		b.WriteString(a.renderHelpScreen(innerW, h-2))
		return lipgloss.NewStyle().
			Width(w).
			Height(h).
			Background(bgColor).
			PaddingLeft(1).
			PaddingRight(1).
			Render(b.String())
	}

	if len(a.flatTasks) == 0 {
		empty := "empty list"
		if a.state.SearchQuery != "" {
			empty = "no matches"
		}
		help := []string{
			"press a to add a task",
			"press W to add a workspace",
			"or use /ws add <name>",
		}
		content := lipgloss.JoinVertical(lipgloss.Center,
			lipgloss.NewStyle().Foreground(dim).Render(empty),
			lipgloss.NewStyle().Foreground(dim).Render(strings.Join(help, "\n")),
		)
		remaining := h - 2
		if remaining < 0 {
			remaining = 0
		}
		block := lipgloss.NewStyle().
			Width(innerW).
			Height(remaining).
			Align(lipgloss.Center).
			Background(bgColor).
			Render(content)
		b.WriteString(block)
	} else {
		infoHeight := 0
		if a.shouldShowTaskInfo() {
			infoHeight = a.taskInfoHeight()
		}
		maxLines := h
		if infoHeight > 0 && maxLines > infoHeight {
			maxLines = maxLines - infoHeight
		}
		visibleCount := maxLines - 2
		if visibleCount < 1 {
			visibleCount = 1
		}

		// Adjust scroll to keep selected task visible
		if a.state.SelectedTask < a.taskScroll {
			a.taskScroll = a.state.SelectedTask
		} else if a.state.SelectedTask >= a.taskScroll+visibleCount {
			a.taskScroll = a.state.SelectedTask - visibleCount + 1
		}
		// Clamp scroll to valid range
		maxScroll := len(a.flatTasks) - visibleCount
		if maxScroll < 0 {
			maxScroll = 0
		}
		if a.taskScroll > maxScroll {
			a.taskScroll = maxScroll
		}
		if a.taskScroll < 0 {
			a.taskScroll = 0
		}

		var taskLines []string
		endIdx := a.taskScroll + visibleCount
		if endIdx > len(a.flatTasks) {
			endIdx = len(a.flatTasks)
		}
		for i := a.taskScroll; i < endIdx; i++ {
			line := a.flatTasks[i]
			task := line.Task
			prefix := strings.Repeat("  ", line.Depth)
			marker := " "
			if len(task.Children) > 0 {
				if a.state.SearchQuery != "" || a.state.ExpandedTasks[task.ID] || task.ParentID == nil {
					marker = "v"
				} else {
					marker = ">"
				}
			}
			prefix = prefix + marker + " "

			rendered := a.renderTaskLine(task, prefix, innerW)
			if i == a.state.SelectedTask {
				rendered = lipgloss.NewStyle().
					Width(innerW).
					Background(cursorBg).
					Render(rendered)
			}
			taskLines = append(taskLines, rendered)
		}
		b.WriteString(strings.Join(taskLines, "\n"))
		if a.shouldShowTaskInfo() {
			if len(taskLines) > 0 {
				b.WriteString("\n")
			}
			b.WriteString(a.renderTaskInfo(innerW))
		}
	}

	if a.state.Mode == model.ModeInsert {
		b.WriteString("\n" + lipgloss.NewStyle().
			Foreground(accent).
			Render("+ "+a.taskInputBuf+"_"))
	}

	return lipgloss.NewStyle().
		Width(w).
		Height(h).
		Background(bgColor).
		PaddingLeft(1).
		PaddingRight(1).
		Render(b.String())
}

func (a *App) renderTaskLine(task *model.Task, prefix string, width int) string {
	checkbox, progress := a.checkboxAndProgress(task)
	meta := []string{}
	if progress != "" {
		meta = append(meta, "["+progress+"]")
	}
	if len(task.Tags) > 0 {
		meta = append(meta, formatTags(task.Tags))
	}
	metaStr := strings.TrimSpace(strings.Join(meta, " "))
	if metaStr != "" {
		metaStr = lipgloss.NewStyle().Foreground(dim).Render(metaStr)
	}

	left := strings.TrimSpace(fmt.Sprintf("%s%s %s", prefix, checkbox, task.Title))
	if metaStr != "" {
		left = left + " " + metaStr
	}

	rightParts := []string{}
	if icon := priorityIcon(task.Priority); icon != "" {
		rightParts = append(rightParts, icon)
	}
	if task.DueDate != "" {
		rightParts = append(rightParts, task.DueDate)
	}
	right := strings.Join(rightParts, " ")

	if right != "" {
		rightWidth := lipgloss.Width(right)
		availLeft := width - rightWidth - 1
		if availLeft < 0 {
			availLeft = 0
		}
		if lipgloss.Width(left) > availLeft {
			left = truncateText(left, availLeft)
		}
		padding := width - lipgloss.Width(left) - rightWidth
		if padding < 1 {
			padding = 1
		}
		left = left + strings.Repeat(" ", padding) + right
	} else if lipgloss.Width(left) > width {
		left = truncateText(left, width)
	}

	if task.Priority < 0 {
		left = lipgloss.NewStyle().Foreground(dim).Render(left)
	} else if a.taskIsComplete(task) {
		left = lipgloss.NewStyle().Strikethrough(true).Foreground(doneColor).Render(left)
	}

	return left
}

func (a *App) renderStatus() string {
	modeName := []string{"NORMAL", "INSERT", "COMMAND", "SEARCH"}[a.state.Mode]
	paneName := "TASKS"
	if a.state.ActivePane == model.PaneWorkspaces {
		paneName = "WORKSPACES"
	}

	left := fmt.Sprintf(" %s · %s ", modeName, paneName)

	right := ""
	if a.state.SelectedWS < len(a.workspaces) {
		completed, open, blocked := a.countStats(a.tasks)
		weather := ""
		if a.weatherEnabled {
			weather = a.weatherTemp + " "
		}
		right = fmt.Sprintf(" ✔ %d ☐ %d ✖ %d %s%s ", completed, open, blocked, weather, time.Now().Format("03:04 PM"))
	} else {
		weather := ""
		if a.weatherEnabled {
			weather = a.weatherTemp + " "
		}
		right = fmt.Sprintf(" %s%s ", weather, time.Now().Format("03:04 PM"))
	}

	statusWidth := a.width - lipgloss.Width(left) - lipgloss.Width(right)
	if statusWidth > 0 {
		left += strings.Repeat(" ", statusWidth)
	}

	return lipgloss.NewStyle().
		Width(a.width).
		Background(borderColor).
		Render(lipgloss.JoinVertical(lipgloss.Left, left+right, padToWidth(a.renderCommandLine(), a.width)))
}

func (a *App) renderCommandLine() string {
	switch a.state.Mode {
	case model.ModeCommand:
		leader := a.commandLeader
		if leader == "" {
			leader = ":"
		}
		return lipgloss.NewStyle().Foreground(accent).Render(leader + a.state.CommandBuf)
	case model.ModeSearch:
		return lipgloss.NewStyle().Foreground(accent).Render("?"+a.state.SearchBuf)
	default:
		if a.state.Msg != "" {
			return lipgloss.NewStyle().Foreground(accent).Render(a.state.Msg)
		}
		return lipgloss.NewStyle().Foreground(dim).Render(commandHelp())
	}
}

func (a *App) statusHeight() int {
	if a.state.Mode == model.ModeCommand || a.state.Mode == model.ModeSearch || a.state.Msg != "" {
		return 2
	}
	return 1
}

func (a *App) loadWorkspaces() {
	a.workspaces, _ = a.db.GetWorkspaces()
	if len(a.workspaces) == 0 {
		a.state.SelectedWS = 0
		a.tasks = nil
		a.flatTasks = nil
		return
	}
	if a.state.SelectedWS >= len(a.workspaces) {
		a.state.SelectedWS = len(a.workspaces) - 1
	}
	if len(a.workspaces) > 0 {
		a.loadTasks()
	}
}

func (a *App) loadTasks() {
	if a.state.SelectedWS >= len(a.workspaces) {
		a.tasks = nil
		a.flatTasks = nil
		return
	}
	ws := a.workspaces[a.state.SelectedWS]
	a.tasks, _ = a.db.GetTasksForWorkspace(ws.ID)
	a.taskIndex = make(map[int64]*model.Task)
	a.indexTasks(a.tasks)
	a.flattenTasks()
}

func (a *App) flattenTasks() {
	a.flatTasks = nil
	tasks := a.applyFilter(a.tasks, a.state.SearchQuery)
	a.walkTasks(tasks, 0, 0)
	if a.state.SelectedTask >= len(a.flatTasks) {
		a.state.SelectedTask = len(a.flatTasks) - 1
	}
	if a.state.SelectedTask < 0 {
		a.state.SelectedTask = 0
	}
}

func (a *App) walkTasks(tasks []*model.Task, depth, index int) int {
	for _, t := range tasks {
		a.flatTasks = append(a.flatTasks, TaskLine{
			Task:     t,
			Depth:    depth,
			Expanded: a.state.ExpandedTasks[t.ID],
			Index:    index,
		})
		index++

		if len(t.Children) > 0 && (a.state.SearchQuery != "" || a.state.ExpandedTasks[t.ID] || t.ParentID == nil) {
			index = a.walkTasks(t.Children, depth+1, index)
		}
	}
	return index
}

func (a *App) selectWorkspace(idx int) {
	a.state.SelectedWS = idx
	a.state.SelectedTask = 0
	a.taskScroll = 0
	a.loadTasks()
}

func (a *App) openCommand(leader string) {
	a.openCommandWithBuffer(leader, "")
}

func (a *App) openCommandWithBuffer(leader, buf string) {
	a.state.Mode = model.ModeCommand
	a.state.CommandBuf = buf
	a.commandLeader = leader
}

func (a *App) togglePane() {
	if a.state.ActivePane == model.PaneTasks {
		a.state.ActivePane = model.PaneWorkspaces
	} else {
		a.state.ActivePane = model.PaneTasks
	}
}

func (a *App) moveWorkspace(dir int) {
	if len(a.workspaces) == 0 {
		return
	}
	newPos := a.state.SelectedWS + dir
	if newPos < 0 {
		newPos = 0
	} else if newPos >= len(a.workspaces) {
		newPos = len(a.workspaces) - 1
	}
	a.selectWorkspace(newPos)
}

func (a *App) selectWorkspaceByToken(token string) {
	if len(a.workspaces) == 0 {
		return
	}

	if idx := parseIndexToken(token); idx >= 0 && idx < len(a.workspaces) {
		a.selectWorkspace(idx)
		return
	}

	lower := strings.ToLower(token)
	for i, ws := range a.workspaces {
		if strings.ToLower(ws.Name) == lower {
			a.selectWorkspace(i)
			return
		}
	}

	for i, ws := range a.workspaces {
		if strings.Contains(strings.ToLower(ws.Name), lower) {
			a.selectWorkspace(i)
			return
		}
	}
}

func (a *App) moveCursor(dir int) {
	if len(a.flatTasks) == 0 {
		a.state.SelectedTask = 0
		return
	}
	newPos := a.state.SelectedTask + dir
	if newPos < 0 {
		newPos = 0
	} else if newPos >= len(a.flatTasks) {
		newPos = len(a.flatTasks) - 1
	}
	a.state.SelectedTask = newPos
}

func (a *App) moveToTop() {
	if len(a.flatTasks) == 0 {
		a.state.SelectedTask = 0
		a.taskScroll = 0
		return
	}
	a.state.SelectedTask = 0
	a.taskScroll = 0
}

func (a *App) moveToBottom() {
	if len(a.flatTasks) == 0 {
		a.state.SelectedTask = 0
		return
	}
	a.state.SelectedTask = len(a.flatTasks) - 1
}

func (a *App) toggleTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	if len(task.Children) > 0 {
		target := !a.taskIsComplete(task)
		a.setTaskTreeCompleted(task, target)
	} else {
		a.db.SetTaskCompleted(task.ID, !task.Completed)
	}
	a.loadTasks()
}

func (a *App) toggleTaskInfo() {
	a.showTaskInfo = !a.showTaskInfo
}

func (a *App) shouldShowTaskInfo() bool {
	return a.showTaskInfo && !a.showAsciiList
}

func (a *App) selectedTaskHasInfo() bool {
	task := a.selectedTask()
	if task == nil {
		return false
	}
	return len(task.Tags) > 0 || task.DueDate != "" || task.Priority != 0
}

func (a *App) selectedTask() *model.Task {
	if a.state.SelectedTask >= len(a.flatTasks) || a.state.SelectedTask < 0 {
		return nil
	}
	return a.flatTasks[a.state.SelectedTask].Task
}

func (a *App) taskInfoHeight() int {
	return 6
}

func (a *App) renderTaskInfo(width int) string {
	task := a.selectedTask()
	if task == nil {
		return ""
	}

	// Use theme colors
	panelBg := infoBg
	labelStyle := lipgloss.NewStyle().Foreground(dim).Background(panelBg)
	valueStyle := lipgloss.NewStyle().Foreground(textColor).Background(panelBg)
	accentValue := lipgloss.NewStyle().Foreground(accent).Background(panelBg)

	// Header with accent background
	headerBg := accent
	headerFg := bgColor
	headerStyle := lipgloss.NewStyle().
		Foreground(headerFg).
		Background(headerBg).
		Bold(true).
		Padding(0, 1)

	// Status
	status := "pending"
	statusStyle := valueStyle
	if task.Completed {
		status = "done"
		statusStyle = lipgloss.NewStyle().Foreground(doneColor).Background(panelBg)
	}

	// Priority with color
	var priorityStr string
	priorityStyle := valueStyle
	switch {
	case task.Priority < 0:
		priorityStr = "blocked"
		priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e06c75")).Background(panelBg)
	case task.Priority >= 2:
		priorityStr = "high"
		priorityStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#e5c07b")).Background(panelBg)
	case task.Priority == 1:
		priorityStr = "low"
	default:
		priorityStr = "normal"
	}

	// Due date
	dueStr := "-"
	if task.DueDate != "" {
		dueStr = task.DueDate
	}

	// Tags - truncate if too long
	tagStr := "-"
	if len(task.Tags) > 0 {
		tagStr = formatTags(task.Tags)
	}
	maxValueWidth := width - 12 // account for label and padding
	if maxValueWidth < 10 {
		maxValueWidth = 10
	}
	if len(tagStr) > maxValueWidth {
		tagStr = tagStr[:maxValueWidth-3] + "..."
	}

	// Build header line with full-width accent background
	headerText := headerStyle.Render("▸ DETAILS")
	headerPadding := width - lipgloss.Width(headerText)
	if headerPadding < 0 {
		headerPadding = 0
	}
	headerLine := headerText + lipgloss.NewStyle().Background(headerBg).Render(strings.Repeat(" ", headerPadding))

	// Content lines
	line2 := labelStyle.Render(" Status   ") + statusStyle.Render(status)
	line3 := labelStyle.Render(" Priority ") + priorityStyle.Render(priorityStr)
	line4 := labelStyle.Render(" Due      ") + accentValue.Render(dueStr)
	line5 := labelStyle.Render(" Tags     ") + valueStyle.Render(tagStr)

	// Wrap content in box style
	boxStyle := lipgloss.NewStyle().
		Background(panelBg).
		Width(width)

	return headerLine + "\n" +
		boxStyle.Render(line2) + "\n" +
		boxStyle.Render(line3) + "\n" +
		boxStyle.Render(line4) + "\n" +
		boxStyle.Render(line5)
}

func (a *App) collapseTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	if len(task.Children) > 0 {
		delete(a.state.ExpandedTasks, task.ID)
		a.loadTasks()
	}
}

func (a *App) expandTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	if len(task.Children) > 0 {
		a.state.ExpandedTasks[task.ID] = true
		a.loadTasks()
	}
}

func (a *App) saveNewTask() {
	if a.state.SelectedWS >= len(a.workspaces) {
		return
	}
	ws := a.workspaces[a.state.SelectedWS]
	parsed := ParseTaskInput(a.taskInputBuf)
	if parsed.Title == "" {
		a.state.Mode = model.ModeNormal
		a.taskInputBuf = ""
		a.newTaskParent = nil
		return
	}
	a.db.AddTaskWithMeta(ws.ID, parsed.Title, a.newTaskParent, parsed.Tags, parsed.DueDate, parsed.Priority)
	a.state.Mode = model.ModeNormal
	a.taskInputBuf = ""
	a.newTaskParent = nil
	a.loadTasks()
}

func (a *App) deleteTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	a.db.DeleteTask(task.ID)
	a.loadTasks()
}

func (a *App) executeCommand() {
	command := strings.TrimSpace(a.state.CommandBuf)
	fields := strings.Fields(strings.ToLower(command))
	originalFields := strings.Fields(command) // preserve case for arguments
	if len(fields) == 0 {
		a.state.Mode = model.ModeNormal
		a.state.CommandBuf = ""
		return
	}

	switch fields[0] {
	case "q", "quit", "wq":
		a.quitRequested = true
	case "ws", "workspace", "workspaces":
		a.executeWorkspaceCommand(fields)
	case "ascii", "art":
		a.executeAsciiCommand(fields)
	case "scheme":
		a.executeSchemeCommand(fields)
	case "settings":
		a.executeSettingsCommand(fields)
	case "weather":
		a.executeWeatherCommand(fields)
	case "search":
		a.executeSearchCommand(fields)
	case "info":
		a.executeInfoCommand(fields)
	case "help":
		a.showHelp = true
	case "focus", "pane":
		if len(fields) > 1 {
			if fields[1] == "ws" || fields[1] == "workspaces" || fields[1] == "workspace" {
				a.state.ActivePane = model.PaneWorkspaces
			} else if fields[1] == "tasks" || fields[1] == "todos" {
				a.state.ActivePane = model.PaneTasks
			}
		}
	case "due", "date":
		a.executeDueCommand(originalFields)
	case "tag", "tags":
		a.executeTagCommand(originalFields)
	case "priority", "p":
		a.executePriorityCommand(fields)
	case "clear":
		a.executeClearCommand(fields)
	case "dashboard", "dash", "db":
		a.executeDashboardCommand(fields)
	}
	a.state.Mode = model.ModeNormal
	a.state.CommandBuf = ""
}

func (a *App) executeDueCommand(fields []string) {
	task := a.selectedTask()
	if task == nil {
		a.state.Msg = "no task selected"
		a.state.MsgTimeout = 3
		return
	}
	if len(fields) < 2 {
		a.state.Msg = "usage: :due <date|today|tomorrow|monday...>"
		a.state.MsgTimeout = 3
		return
	}
	task.DueDate = parseDueDate(fields[1])
	a.db.UpdateTask(task)
	a.loadTasks()
	a.state.Msg = "due date set to " + task.DueDate
	a.state.MsgTimeout = 2
}

func (a *App) executeTagCommand(fields []string) {
	task := a.selectedTask()
	if task == nil {
		a.state.Msg = "no task selected"
		a.state.MsgTimeout = 3
		return
	}
	if len(fields) < 2 {
		a.state.Msg = "usage: :tag <tag1> [tag2] ..."
		a.state.MsgTimeout = 3
		return
	}
	for _, t := range fields[1:] {
		tag := strings.TrimPrefix(t, "#")
		if tag != "" {
			task.Tags = append(task.Tags, tag)
		}
	}
	a.db.UpdateTask(task)
	a.loadTasks()
	a.state.Msg = "tags updated"
	a.state.MsgTimeout = 2
}

func (a *App) executePriorityCommand(fields []string) {
	task := a.selectedTask()
	if task == nil {
		a.state.Msg = "no task selected"
		a.state.MsgTimeout = 3
		return
	}
	if len(fields) < 2 {
		a.state.Msg = "usage: :priority <high|low|normal|blocked>"
		a.state.MsgTimeout = 3
		return
	}
	switch fields[1] {
	case "high", "h", "2":
		task.Priority = 2
	case "low", "l", "1":
		task.Priority = 1
	case "blocked", "b", "-1":
		task.Priority = -1
	case "normal", "n", "0":
		task.Priority = 0
	default:
		a.state.Msg = "unknown priority: " + fields[1]
		a.state.MsgTimeout = 3
		return
	}
	a.db.UpdateTask(task)
	a.loadTasks()
	a.state.Msg = "priority updated"
	a.state.MsgTimeout = 2
}

func (a *App) executeClearCommand(fields []string) {
	task := a.selectedTask()
	if task == nil {
		a.state.Msg = "no task selected"
		a.state.MsgTimeout = 3
		return
	}
	if len(fields) < 2 {
		a.state.Msg = "usage: :clear <due|tags|priority|all>"
		a.state.MsgTimeout = 3
		return
	}
	switch fields[1] {
	case "due", "date":
		task.DueDate = ""
	case "tags", "tag":
		task.Tags = nil
	case "priority", "p":
		task.Priority = 0
	case "all":
		task.DueDate = ""
		task.Tags = nil
		task.Priority = 0
	default:
		a.state.Msg = "unknown field: " + fields[1]
		a.state.MsgTimeout = 3
		return
	}
	a.db.UpdateTask(task)
	a.loadTasks()
	a.state.Msg = "cleared " + fields[1]
	a.state.MsgTimeout = 2
}

func (a *App) executeDashboardCommand(fields []string) {
	if len(fields) < 2 {
		// Toggle
		a.showDashboard = !a.showDashboard
		if a.showDashboard {
			a.state.Msg = "dashboard on"
		} else {
			a.state.Msg = "dashboard off"
		}
		a.state.MsgTimeout = 2
		return
	}
	switch fields[1] {
	case "on", "show", "1":
		a.showDashboard = true
		a.state.Msg = "dashboard on"
	case "off", "hide", "0":
		a.showDashboard = false
		a.state.Msg = "dashboard off"
	default:
		a.state.Msg = "usage: :dashboard [on|off]"
	}
	a.state.MsgTimeout = 2
}

func (a *App) executeWorkspaceCommand(fields []string) {
	if len(fields) == 1 {
		a.state.ActivePane = model.PaneWorkspaces
		return
	}

	switch fields[1] {
	case "add", "new", "create":
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		if name != "" {
			a.createWorkspace(name)
		}
	case "rename":
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		if name != "" {
			a.renameWorkspace(name)
		}
	case "delete", "del", "rm":
		a.deleteWorkspace()
	case "select", "open":
		if len(fields) > 2 {
			a.selectWorkspaceByToken(fields[2])
		}
	default:
		a.selectWorkspaceByToken(fields[1])
	}
}

func (a *App) createWorkspace(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	id, err := a.db.CreateWorkspace(name)
	if err != nil {
		return
	}
	a.loadWorkspaces()
	for i, ws := range a.workspaces {
		if ws.ID == id {
			a.selectWorkspace(i)
			break
		}
	}
}

func (a *App) renameWorkspace(name string) {
	if a.state.SelectedWS >= len(a.workspaces) {
		return
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	ws := a.workspaces[a.state.SelectedWS]
	if err := a.db.RenameWorkspace(ws.ID, name); err != nil {
		return
	}
	a.loadWorkspaces()
}

func (a *App) deleteWorkspace() {
	if a.state.SelectedWS >= len(a.workspaces) {
		return
	}
	ws := a.workspaces[a.state.SelectedWS]
	if err := a.db.DeleteWorkspace(ws.ID); err != nil {
		return
	}
	a.loadWorkspaces()
	if len(a.workspaces) == 0 {
		a.state.SelectedWS = 0
		a.tasks = nil
		a.flatTasks = nil
		return
	}
	if a.state.SelectedWS >= len(a.workspaces) {
		a.state.SelectedWS = len(a.workspaces) - 1
	}
	a.loadTasks()
}

func (a *App) executeAsciiCommand(fields []string) {
	if len(fields) < 2 || fields[1] == "list" {
		if len(a.asciiArts) == 0 {
			a.setMessage("no ascii art found")
			return
		}
		a.buildAsciiLines()
		a.asciiScroll = 0
		a.showAsciiList = true
		return
	}
	switch fields[1] {
	case "hide", "clear":
		a.showAsciiList = false
	case "random":
		a.pickRandomAscii()
		a.showAsciiList = false
	}
}

func (a *App) executeSchemeCommand(fields []string) {
	if len(fields) < 2 || fields[1] == "list" {
		a.setMessage("schemes: " + strings.Join(availableSchemes(), ", "))
		return
	}
	if a.applyScheme(fields[1]) {
		a.setMessage("scheme: " + fields[1])
	} else {
		a.setMessage("scheme not found")
	}
}

func (a *App) applyScheme(name string) bool {
	for _, scheme := range schemes {
		if scheme.name == name {
			bgColor = scheme.bg
			sidebarBg = scheme.sidebar
			selectionBg = scheme.selection
			cursorBg = scheme.cursor
			accent = scheme.accent
			dim = scheme.dim
			textColor = scheme.text
			headerColor = scheme.header
			borderColor = scheme.border
			doneColor = scheme.done
			infoBg = scheme.info
			a.headerCache = ""
			return true
		}
	}
	return false
}

func (a *App) performSearch() {
	a.state.SearchQuery = strings.TrimSpace(a.state.SearchBuf)
	a.state.SearchBuf = ""
	a.state.Mode = model.ModeNormal
	a.flattenTasks()
}

func (a *App) countCompletedChildren(task *model.Task) int {
	count := 0
	for _, c := range task.Children {
		if a.taskIsComplete(c) {
			count++
		}
	}
	return count
}

func (a *App) countStats(tasks []*model.Task) (completed, open, blocked int) {
	for _, task := range tasks {
		if task.Priority < 0 {
			blocked++
		} else if a.taskIsComplete(task) {
			completed++
		} else {
			open++
		}
		if len(task.Children) > 0 {
			c, o, b := a.countStats(task.Children)
			completed += c
			open += o
			blocked += b
		}
	}
	return completed, open, blocked
}

func (a *App) checkboxAndProgress(task *model.Task) (string, string) {
	if task.Priority < 0 {
		return "✖", ""
	}
	if len(task.Children) > 0 {
		completed := a.countCompletedChildren(task)
		total := len(task.Children)
		progress := fmt.Sprintf("%d/%d", completed, total)
		if completed == total {
			return "☑", progress
		}
		return "☐", progress
	}
	if task.Completed {
		return "☑", ""
	}
	return "☐", ""
}

func (a *App) taskIsComplete(task *model.Task) bool {
	if task.Priority < 0 {
		return false
	}
	if len(task.Children) == 0 {
		return task.Completed
	}
	return a.countCompletedChildren(task) == len(task.Children)
}

func (a *App) setTaskTreeCompleted(task *model.Task, completed bool) {
	a.db.SetTaskCompleted(task.ID, completed)
	for _, child := range task.Children {
		a.setTaskTreeCompleted(child, completed)
	}
}

func (a *App) addTask() {
	a.state.Mode = model.ModeInsert
	a.taskInputBuf = ""
	a.editingTaskID = nil
	a.newTaskParent = nil
	if a.state.SelectedTask < len(a.flatTasks) {
		a.newTaskParent = a.flatTasks[a.state.SelectedTask].Task.ParentID
	}
}

func (a *App) editTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	a.state.Mode = model.ModeInsert
	a.taskInputBuf = task.Title
	a.editingTaskID = &task.ID
	a.newTaskParent = nil
}

func (a *App) saveTaskEdit() {
	if a.editingTaskID == nil {
		return
	}
	task := a.taskIndex[*a.editingTaskID]
	if task == nil {
		return
	}
	parsed := ParseTaskInput(a.taskInputBuf)
	task.Title = parsed.Title
	if len(parsed.Tags) > 0 {
		task.Tags = append(task.Tags, parsed.Tags...)
	}
	if parsed.DueDate != "" {
		task.DueDate = parsed.DueDate
	}
	if parsed.Priority != 0 {
		task.Priority = parsed.Priority
	}
	a.db.UpdateTask(task)
	a.state.Mode = model.ModeNormal
	a.taskInputBuf = ""
	a.editingTaskID = nil
	a.loadTasks()
}

func (a *App) indentTask() {
	if a.state.SelectedTask <= 0 || a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	current := a.flatTasks[a.state.SelectedTask]
	prev := a.flatTasks[a.state.SelectedTask-1]
	if prev.Depth != current.Depth {
		return
	}
	if a.isDescendant(prev.Task, current.Task.ID) {
		return
	}
	a.db.MoveTask(current.Task.ID, &prev.Task.ID)
	a.state.ExpandedTasks[prev.Task.ID] = true
	a.loadTasks()
}

func (a *App) unindentTask() {
	if a.state.SelectedTask >= len(a.flatTasks) {
		return
	}
	task := a.flatTasks[a.state.SelectedTask].Task
	if task.ParentID == nil {
		return
	}
	parent := a.taskIndex[*task.ParentID]
	if parent == nil {
		return
	}
	a.db.MoveTask(task.ID, parent.ParentID)
	a.loadTasks()
}

func (a *App) isDescendant(task *model.Task, targetID int64) bool {
	if task.ID == targetID {
		return true
	}
	for _, child := range task.Children {
		if a.isDescendant(child, targetID) {
			return true
		}
	}
	return false
}

func (a *App) applyFilter(tasks []*model.Task, query string) []*model.Task {
	query = strings.TrimSpace(strings.ToLower(query))
	if query == "" {
		return tasks
	}

	var filtered []*model.Task
	for _, task := range tasks {
		matches := a.taskMatches(task, query)
		children := a.applyFilter(task.Children, query)
		if matches || len(children) > 0 {
			clone := *task
			clone.Children = children
			filtered = append(filtered, &clone)
		}
	}
	return filtered
}

func (a *App) taskMatches(task *model.Task, query string) bool {
	if strings.Contains(strings.ToLower(task.Title), query) {
		return true
	}
	for _, tag := range task.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	return false
}

func (a *App) indexTasks(tasks []*model.Task) {
	for _, task := range tasks {
		a.taskIndex[task.ID] = task
		if len(task.Children) > 0 {
			a.indexTasks(task.Children)
		}
	}
}

func (a *App) handlePendingKey(key string) {
	switch key {
	case "d", "g":
		if a.state.ActivePane != model.PaneTasks {
			a.pendingKey = ""
			return
		}
		if a.pendingKey == key && time.Since(a.pendingAt) < 800*time.Millisecond {
			a.pendingKey = ""
			if key == "d" {
				a.deleteTask()
			} else {
				a.moveToTop()
			}
			return
		}
		a.pendingKey = key
		a.pendingAt = time.Now()
	default:
		a.pendingKey = ""
	}
}

func (a *App) clearPendingKey() {
	if a.pendingKey == "" {
		return
	}
	if time.Since(a.pendingAt) > time.Second {
		a.pendingKey = ""
	}
}

func formatTags(tags []string) string {
	var cleaned []string
	for _, tag := range tags {
		tag = strings.TrimSpace(tag)
		if tag == "" {
			continue
		}
		if !strings.HasPrefix(tag, "#") {
			tag = "#" + tag
		}
		cleaned = append(cleaned, tag)
	}
	return strings.Join(cleaned, " ")
}

func priorityIcon(priority int) string {
	switch {
	case priority >= 2:
		return "^"
	case priority == 1:
		return "."
	default:
		return ""
	}
}

func truncateText(s string, width int) string {
	if width <= 0 || s == "" {
		return ""
	}
	if lipgloss.Width(s) <= width {
		return s
	}

	count := 0
	for i, r := range s {
		if count == width {
			return s[:i]
		}
		if r == '\n' || r == '\r' {
			return s[:i]
		}
		count++
	}
	return s
}

func padToWidth(s string, width int) string {
	if width <= 0 {
		return ""
	}
	padding := width - lipgloss.Width(s)
	if padding <= 0 {
		return s
	}
	return s + strings.Repeat(" ", padding)
}

func clipArt(art string, width int) string {
	if width <= 0 || art == "" {
		return ""
	}
	lines := strings.Split(art, "\n")
	for i := range lines {
		lines[i] = truncateText(strings.TrimRight(lines[i], " "), width)
	}
	return strings.Join(lines, "\n")
}

func maxWidth(width int) int {
	if width < 0 {
		return 0
	}
	return width
}

func commandHelp() string {
	return ":q  /help  /ws add <name>  /ws rename <name>  /ws delete  /ws <name|#>  /search <query>  /info on|off"
}

func (a *App) setMessage(msg string) {
	a.state.Msg = msg
	a.state.MsgTimeout = 6
}

func (a *App) tickMessage() {
	if a.state.MsgTimeout <= 0 {
		return
	}
	a.state.MsgTimeout--
	if a.state.MsgTimeout == 0 {
		a.state.Msg = ""
	}
}

func availableSchemes() []string {
	names := make([]string, 0, len(schemes))
	for _, scheme := range schemes {
		names = append(names, scheme.name)
	}
	return names
}

func (a *App) executeSettingsCommand(fields []string) {
	if len(fields) < 2 {
		a.setMessage("settings: weather on|off, city <name>, unit c|f")
		return
	}
	switch fields[1] {
	case "weather":
		if len(fields) > 2 {
			switch fields[2] {
			case "on":
				a.weatherEnabled = true
				_ = a.db.SetSetting("weather_enabled", "1")
				a.weatherChecked = time.Time{}
				a.setMessage("weather on")
			case "off":
				a.weatherEnabled = false
				_ = a.db.SetSetting("weather_enabled", "0")
				a.setMessage("weather off")
			}
		}
	case "city":
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		if name != "" {
			a.weatherCity = name
			_ = a.db.SetSetting("weather_city", name)
			a.weatherChecked = time.Time{}
			a.setMessage("city set: " + name)
		}
	case "unit":
		if len(fields) > 2 {
			unit := strings.ToLower(fields[2])
			if unit == "c" || unit == "celsius" {
				a.weatherUnit = "c"
				_ = a.db.SetSetting("weather_unit", "c")
				a.weatherChecked = time.Time{}
				a.setMessage("unit: celsius")
			} else if unit == "f" || unit == "fahrenheit" {
				a.weatherUnit = "f"
				_ = a.db.SetSetting("weather_unit", "f")
				a.weatherChecked = time.Time{}
				a.setMessage("unit: fahrenheit")
			}
		}
	}
}

func (a *App) executeWeatherCommand(fields []string) {
	if len(fields) < 2 {
		a.setMessage("weather: refresh | city <name>")
		return
	}
	switch fields[1] {
	case "refresh":
		a.weatherChecked = time.Time{}
	case "city":
		name := strings.TrimSpace(strings.Join(fields[2:], " "))
		if name != "" {
			a.weatherCity = name
			_ = a.db.SetSetting("weather_city", name)
			a.weatherChecked = time.Time{}
		}
	}
}

func (a *App) executeSearchCommand(fields []string) {
	query := strings.TrimSpace(strings.Join(fields[1:], " "))
	if query == "" {
		a.state.SearchQuery = ""
		a.setMessage("search cleared")
	} else {
		a.state.SearchQuery = query
		a.setMessage("search: " + query)
	}
	a.flattenTasks()
}

func (a *App) executeInfoCommand(fields []string) {
	if len(fields) < 2 {
		a.toggleTaskInfo()
		return
	}
	switch fields[1] {
	case "on":
		a.showTaskInfo = true
	case "off":
		a.showTaskInfo = false
	default:
		a.toggleTaskInfo()
	}
}

func (a *App) maybeFetchWeather() tea.Cmd {
	if !a.weatherEnabled {
		return nil
	}
	if time.Since(a.weatherChecked) < 30*time.Minute {
		return nil
	}
	if a.weatherCity == "" && (a.weatherLat == 0 || a.weatherLon == 0) {
		return nil
	}
	return fetchWeatherCmd(a.weatherCity, a.weatherLat, a.weatherLon, a.weatherUnit)
}

func fetchWeatherCmd(city string, lat, lon float64, unit string) tea.Cmd {
	return func() tea.Msg {
		temp, newLat, newLon, err := fetchWeather(city, lat, lon, unit)
		if err != nil {
			return weatherMsg{err: err}
		}
		return weatherMsg{temp: temp, lat: newLat, lon: newLon}
	}
}

func fetchWeather(city string, lat, lon float64, unit string) (string, float64, float64, error) {
	if city != "" && (lat == 0 || lon == 0) {
		geoURL := fmt.Sprintf("https://geocoding-api.open-meteo.com/v1/search?name=%s&count=1&language=en&format=json", urlQuery(city))
		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Get(geoURL)
		if err != nil {
			return "", 0, 0, err
		}
		defer res.Body.Close()
		var geo struct {
			Results []struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"results"`
		}
		if err := json.NewDecoder(res.Body).Decode(&geo); err != nil {
			return "", 0, 0, err
		}
		if len(geo.Results) == 0 {
			return "", 0, 0, fmt.Errorf("no results")
		}
		lat = geo.Results[0].Latitude
		lon = geo.Results[0].Longitude
	}

	unitParam := "fahrenheit"
	suffix := "°F"
	if unit == "c" || unit == "celsius" {
		unitParam = "celsius"
		suffix = "°C"
	}
	url := fmt.Sprintf("https://api.open-meteo.com/v1/forecast?latitude=%0.4f&longitude=%0.4f&current=temperature_2m&temperature_unit=%s", lat, lon, unitParam)
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(url)
	if err != nil {
		return "", lat, lon, err
	}
	defer res.Body.Close()
	var forecast struct {
		Current struct {
			Temperature float64 `json:"temperature_2m"`
		} `json:"current"`
	}
	if err := json.NewDecoder(res.Body).Decode(&forecast); err != nil {
		return "", lat, lon, err
	}
	temp := fmt.Sprintf("%.0f%s", forecast.Current.Temperature, suffix)
	return temp, lat, lon, nil
}

func urlQuery(s string) string {
	replacer := strings.NewReplacer(" ", "+")
	return replacer.Replace(strings.TrimSpace(s))
}

func parseIndexToken(token string) int {
	if token == "" {
		return -1
	}
	value := 0
	for i := 0; i < len(token); i++ {
		ch := token[i]
		if ch < '0' || ch > '9' {
			return -1
		}
		value = value*10 + int(ch-'0')
	}
	return value - 1
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type weatherMsg struct {
	temp string
	err  error
	lat  float64
	lon  float64
}

func (a *App) Run() error {
	a.loadAsciiArt()
	a.loadSettings()
	a.loadWorkspaces()
	p := tea.NewProgram(a, tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (a *App) loadSettings() {
	enabled, _ := a.db.GetSetting("weather_enabled")
	city, _ := a.db.GetSetting("weather_city")
	lat, _ := a.db.GetSetting("weather_lat")
	lon, _ := a.db.GetSetting("weather_lon")
	unit, _ := a.db.GetSetting("weather_unit")

	a.weatherEnabled = enabled == "1"
	a.weatherCity = city
	if lat != "" {
		if v, err := strconv.ParseFloat(lat, 64); err == nil {
			a.weatherLat = v
		}
	}
	if lon != "" {
		if v, err := strconv.ParseFloat(lon, 64); err == nil {
			a.weatherLon = v
		}
	}
	if unit != "" {
		a.weatherUnit = unit
	}
}

func (a *App) loadAsciiArt() {
	dir := asciiDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	var arts []string
	var names []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".txt") {
			continue
		}
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		art := strings.TrimRight(string(data), "\n")
		if art == "" {
			continue
		}
		arts = append(arts, art)
		names = append(names, strings.TrimSuffix(name, filepath.Ext(name)))
	}

	if len(arts) == 0 {
		return
	}
	a.asciiArts = arts
	a.asciiNames = names
	a.pickRandomAscii()
}

func (a *App) pickRandomAscii() {
	if len(a.asciiArts) == 0 {
		return
	}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	a.asciiArt = a.asciiArts[r.Intn(len(a.asciiArts))]
	a.headerCache = ""
}

func asciiDir() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		path := filepath.Join(dir, "ascii")
		if dirExists(path) {
			return path
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		path := filepath.Join(cwd, "ascii")
		if dirExists(path) {
			return path
		}
	}
	return "ascii"
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func (a *App) buildAsciiLines() {
	a.asciiLines = nil
	for i, art := range a.asciiArts {
		for _, line := range strings.Split(art, "\n") {
			a.asciiLines = append(a.asciiLines, strings.TrimRight(line, " "))
		}
		if i != len(a.asciiArts)-1 {
			a.asciiLines = append(a.asciiLines, "")
		}
	}
}

func (a *App) scrollAscii(delta int) {
	if len(a.asciiLines) == 0 {
		return
	}
	height := a.asciiListHeight()
	maxScroll := max(0, len(a.asciiLines)-height)
	a.asciiScroll = clamp(a.asciiScroll+delta, 0, maxScroll)
}

func (a *App) scrollAsciiToTop() {
	a.asciiScroll = 0
}

func (a *App) scrollAsciiToEnd() {
	if len(a.asciiLines) == 0 {
		return
	}
	height := a.asciiListHeight()
	a.asciiScroll = max(0, len(a.asciiLines)-height)
}

func (a *App) asciiListHeight() int {
	header := lipgloss.Height(a.renderHeader())
	return a.height - header - a.statusHeight() - 2
}

func (a *App) asciiPageStep() int {
	height := a.asciiListHeight()
	if height <= 2 {
		return 1
	}
	return height - 2
}

func clamp(value, minVal, maxVal int) int {
	if value < minVal {
		return minVal
	}
	if value > maxVal {
		return maxVal
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (a *App) renderAsciiList(width, height int) string {
	if len(a.asciiLines) == 0 {
		return lipgloss.NewStyle().Foreground(dim).Render("no ascii art found\n")
	}
	if height < 0 {
		height = 0
	}
	start := clamp(a.asciiScroll, 0, max(0, len(a.asciiLines)-height))
	end := start + height
	if end > len(a.asciiLines) {
		end = len(a.asciiLines)
	}
	lines := a.asciiLines[start:end]
	for i := range lines {
		lines[i] = truncateText(lines[i], width)
	}
	return strings.Join(lines, "\n")
}

func (a *App) renderHelpScreen(width, height int) string {
	title := lipgloss.NewStyle().Foreground(accent).Render("Help")
	lines := []string{
		title,
		"",
		"Navigation",
		"  j/k or arrows  move cursor",
		"  tab             switch pane",
		"  enter           open workspace / toggle task",
		"",
		"Tasks",
		"  a               add task",
		"  i               edit task",
		"  space / x       toggle task",
		"  dd              delete task",
		"  h/l             collapse / expand",
		"  m               toggle details panel",
		"",
		"Workspaces",
		"  W               add workspace",
		"  R               rename workspace",
		"  X               delete workspace",
		"",
		"Commands",
		"  /help           show this screen",
		"  /search <q>     filter tasks",
		"  /ws add <name>  create workspace",
		"  /scheme list    list themes",
		"  /settings city <name>",
		"  /settings weather on|off",
		"  /settings unit c|f",
		"",
		"Press H or Esc to close.",
	}

	if height > 0 && len(lines) > height {
		lines = lines[:height]
	}

	content := strings.Join(lines, "\n")
	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Left).
		Background(bgColor).
		Render(content)
}
