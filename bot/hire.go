package bot

import (
	"fmt"
	"io"
	"strings"

	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	"github.com/urfave/cli"
	"github.com/zpatrick/slackbot"
)

// NewHireCommand create a cli.Command that allows users to ...
func NewHireCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:  "hire",
		Usage: "manage hiring pipelines",
		Subcommands: []cli.Command{
			{
				Name:      "start",
				Usage:     "start a hiring pipeline for a candidate",
				ArgsUsage: "CANDIDATE",
				Action: func(c *cli.Context) error {
					candidateName := strings.Join(c.Args(), " ")
					if candidateName == "" {
						return slackbot.NewUserInputError("Argument CANDIDATE is required")
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					candidate, ok := candidates.Get(candidateName)
					if !ok {
						return candidateDoesNotExist(candidateName)
					}

					pipelines := models.Pipelines{}
					if err := store.Read(db.PipelinesKey, &pipelines); err != nil {
						return err
					}

					if _, ok := pipelines.Get(candidateName); ok {
						text := fmt.Sprintf("A hiring pipeline for *%s* already exists", candidateName)
						return slackbot.NewUserInputError(text)
					}

					pipeline := newHiringPipeline(candidate.Name)
					pipelines = append(pipelines, &pipeline)
					if err := store.Write(db.PipelinesKey, pipelines); err != nil {
						return err
					}

					name := strings.Title(candidate.Name)
					escapedManagerID := slackbot.EscapeUserID(candidate.ManagerID)
					text := fmt.Sprintf("Ok, I've started a new hiring pipeline for *%s*.\n", name)
					text += fmt.Sprintf("I will send daily reminders to *%s's* manager, ", name)
					text += fmt.Sprintf("%s, to complete the hiring pipeline.", escapedManagerID)
					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "stop",
				Usage:     "remove a candidate from a hiring pipeline",
				ArgsUsage: "CANDIDATE",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:  "ls",
				Usage: "list candidates currently in a hiring pipeling",
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "limit",
						Value: 50,
						Usage: "The maximum number of hires to display",
					},
					cli.BoolFlag{
						Name:  "ascending",
						Usage: "Show results in reverse-alphabetical order",
					},
				},
				Action: func(c *cli.Context) error {
					return nil
				},
			},
			{
				Name:      "show",
				Usage:     "show the hiring pipeline for a candidate",
				ArgsUsage: "CANDIDATE",
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}
}

func newHiringPipeline(candidateName string) models.Pipeline {
	return models.Pipeline{
		Name: candidateName,
		Steps: []string{
			"Do first thing",
			"Do second thing",
			"Do third thing",
		},
	}
}

/*
u: !hire start zack patrick
b: Ok, I've started a hiring pipeline for *Zack Patrick*.
I will send daily reminders to *Zack Patrick's* manager, @bivers, to complete the hiring pipeline.

u: !hire show zack patrick
b: This is the hiring pipeline for *Zack Patrick*:
```
1. Do the first thing
2. Do the second thing
3. Do the third thing
```

*Zack Patrick's* manager, @bivers, is currently on step *1* of the hiring pipeline: `Do the first thing`.

u: !hire next zack patrick
b: Ok, I'll make a note that you've completed step *1* of *Zack Patrick's* hiring pipeline.
The next step is to *Do the third thing*
( or )
There are no more steps for this pipeline.
Thank you for completing *Zack Patrick's* hiring pipeline.
You will no longer receive daily reminders to finish this process.

 - No '!hire back' for now: users can just delete/re-create
*/
