package main

import (
	"fmt"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/nvcnvn/sql-up/pkg"
)

func main() {
	app := &cli.App{
		Name:  "sql-up",
		Usage: "Apply SQL migrations",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "dbms",
				Aliases:  []string{"d"},
				Usage:    "Database management system (postgres, mysql, sqlite3)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "connection-string",
				Aliases:  []string{"c"},
				Usage:    "Connection string",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "sql-file",
				Aliases:  []string{"f"},
				Usage:    "SQL file",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			config := &pkg.Config{
				DBMS:             c.String("dbms"),
				ConnectionString: c.String("connection-string"),
				SQLFile:          c.String("sql-file"),
			}
			return pkg.Up(c.Context, config)
		},
	}

	err := app.Run([]string{"sql-up"})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
