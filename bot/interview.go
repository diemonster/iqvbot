package bot

import (
	"io"

	"github.com/quintilesims/iqvbot/db"
	"github.com/urfave/cli"
)

// NewInterviewCommand creates a cli.Command that allows users to add, list, and delete interviews
func NewInterviewCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:  "interview",
		Usage: "manage interviews",
		Subcommands: []cli.Command{
			{
				Name:      "add",
				Usage:     "add a new interview",
				ArgsUsage: "CANDIDATE DATE (mm/dd/yyyy) TIME (mm:hh{am|pm}) @INTERVIEWERS..",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:      "ls",
				Usage:     "list interviews",
				ArgsUsage: " ",
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "limit",
						Value: 50,
						Usage: "The maximum number of interviews to display",
					},
					cli.BoolFlag{
						Name:  "ascending",
						Usage: "Show results in reverse-chronological order",
					},
				},
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:      "rm",
				Usage:     "remove an interview",
				ArgsUsage: "CANDIDATE DATE (mm/dd/yyyy) TIME (mm:hh{am|pm})",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}
}
