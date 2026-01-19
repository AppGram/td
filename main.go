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

		_, err = database.AddTask(workspaces[0].ID, *addTodo, nil)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Todo added")
		return
	}

	app := tui.New(database)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "UI error: %v\n", err)
		os.Exit(1)
	}
}
