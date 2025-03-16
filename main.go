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
				Name:  "snapshot",
				Usage: "Generate snapshots of all files",
				Action: func(c *cli.Context) error {
					rootDir := "./test_folder" // Change this if needed
					snapshotFile := "snapshots.json"

					fmt.Println("Generating snapshots for directory:", rootDir)
					err := utils.CreateSnapshots(rootDir, snapshotFile)
					if err != nil {
						log.Fatal("Error creating snapshots:", err)
					}
					fmt.Println("Snapshots saved to", snapshotFile)
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
