package main

import (
	"fmt"
	"log"
	"os"

	"axle/utils"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "Axle",
		Usage: "Test Axle functionalities",
		Commands: []*cli.Command{
			{
				Name:  "init",
				Usage: "Generate snapshots of all files and initialize Git repo",
				Action: func(c *cli.Context) error {
					rootDir := "./test_folder" // Change this if needed
					snapshotFile := "snapshots.json"

					fmt.Println("Initializing Git repository...")
					if err := utils.InitGitRepo(rootDir); err != nil {
						log.Fatal("Error initializing Git repo:", err)
					}

					fmt.Println("Generating snapshots for directory:", rootDir)
					if err := utils.CreateSnapshots(rootDir, snapshotFile); err != nil {
						log.Fatal("Error creating snapshots:", err)
					}
					fmt.Println("Snapshots saved to", snapshotFile)
					return nil
				},
			},
			{
				Name:  "test-git-diff",
				Usage: "Test Git diff function on a given file",
				Action: func(c *cli.Context) error {
					if c.NArg() == 0 {
						log.Fatal("Please provide a file path to test git diff")
					}
					filePath := c.Args().Get(0)

					fmt.Println("Generating Git diff for", filePath)
					diff, err := utils.GetGitDiff(filePath)
					if err != nil {
						log.Fatal("Error generating Git diff:", err)
					}
					fmt.Println("Git Diff Output:\n", diff)
					return nil
				},
			},
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
