package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/appgram/td/internal/db"
	"github.com/appgram/td/internal/tui"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	printVersion := flag.Bool("version", false, "Print version and exit")
	addTodo := flag.String("a", "", "Add a new todo item")
	flag.Parse()

	if *printVersion {
		fmt.Printf("td version %s, commit %s, date %s\n", version, commit, date)
		os.Exit(0)
	}

	database, err := db.NewDB()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	if *addTodo != "" {
		workspaces, err := database.GetWorkspaces()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		if len(workspaces) == 0 {
			wsID, err := database.CreateWorkspace("Default")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			workspaces = []db.Workspace{{ID: wsID, Name: "Default", Order: 0, TaskCount: 0, CompletedCount: 0}}
		}

		// Parse inline syntax: "task #tag @date !priority"
		parsed := tui.ParseTaskInput(*addTodo)
		if parsed.Title == "" {
			fmt.Fprintf(os.Stderr, "Error: task title is required\n")
			os.Exit(1)
		}

		_, err = database.AddTaskWithMeta(workspaces[0].ID, parsed.Title, nil, parsed.Tags, parsed.DueDate, parsed.Priority)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Added: %s", parsed.Title)
		if len(parsed.Tags) > 0 {
			fmt.Printf(" [tags: %v]", parsed.Tags)
		}
		if parsed.DueDate != "" {
			fmt.Printf(" [due: %s]", parsed.DueDate)
		}
		if parsed.Priority != 0 {
			p := "normal"
			switch parsed.Priority {
			case 2:
				p = "high"
			case 1:
				p = "low"
			case -1:
				p = "blocked"
			}
			fmt.Printf(" [priority: %s]", p)
		}
		fmt.Println()
		return
	}

	app := tui.New(database)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}
