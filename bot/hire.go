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

// todo: ensure only the candidate's manager can do the step command?
// todo: make step a subcommand? !hire step next, !hire step prev

// NewHireCommand create a cli.Command that allows users to ...
func NewHireCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:  "hire",
		Usage: "manage hiring pipelines",
		Subcommands: []cli.Command{
			{
				Name:      "add",
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
				Name:      "rm",
				Usage:     "remove a candidate from a hiring pipeline",
				ArgsUsage: "CANDIDATE",
				Action: func(c *cli.Context) error {
					candidateName := strings.Join(c.Args(), " ")
					if candidateName == "" {
						return slackbot.NewUserInputError("Argument CANDIDATE is required")
					}

					pipelines := models.Pipelines{}
					if err := store.Read(db.PipelinesKey, &pipelines); err != nil {
						return err
					}

					if !pipelines.Delete(candidateName) {
						return hiringPipelineDoesNotExist(candidateName)
					}

					if err := store.Write(db.PipelinesKey, pipelines); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've deleted *%s's* hiring pipeline", candidateName)
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
					pipelines := models.Pipelines{}
					if err := store.Read(db.PipelinesKey, &pipelines); err != nil {
						return err
					}

					pipelines.FilterByType(models.HiringPipelineType)
					if len(pipelines) == 0 {
						return slackbot.WriteString(w, "There aren't any candidates in hiring pipelines at the moment")
					}

					pipelines.Sort(!c.Bool("ascending"))

					text := "Here are the candidates currently in hiring pipelines: \n"
					for i := 0; i < c.Int("limit") && i < len(pipelines); i++ {
						text += fmt.Sprintf("*%s*\n", strings.Title(pipelines[i].Name))
					}

					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "next",
				Usage:     "move on to the next step in a hiring pipeline",
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

					pipeline, ok := pipelines.Get(candidateName)
					if !ok {
						return hiringPipelineDoesNotExist(candidateName)
					}

					if pipeline.CurrentStep == len(pipeline.Steps) {
						return slackbot.NewUserInputError("This pipeline has already been completed")
					}

					pipeline.CurrentStep += 1
					if err := store.Write(db.PipelinesKey, pipelines); err != nil {
						return err
					}

					name := strings.Title(candidateName)
					escapedManagerID := slackbot.EscapeUserID(candidate.ManagerID)
					if pipeline.CurrentStep >= len(pipeline.Steps) {
						text := "There are no more steps in this pipeline.\n"
						text += fmt.Sprintf("Thank you for completing *%s's* hiring pipeline!\n", name)
						text += fmt.Sprintf("%s will no longer receive reminders to finish this process.", escapedManagerID)
						return slackbot.WriteString(w, text)
					}

					text := fmt.Sprintf("Ok, I'll make a note that you've completed step *%d* ", pipeline.CurrentStep)
					text += fmt.Sprintf("of *%s's* hiring pipeline.\n", name)
					text += fmt.Sprintf("The next step is to: `%s`", pipeline.Steps[pipeline.CurrentStep])
					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "prev",
				Usage:     "revert to the previous step in a hiring pipeline",
				ArgsUsage: "CANDIDATE",
				Action: func(c *cli.Context) error {
					candidateName := strings.Join(c.Args(), " ")
					if candidateName == "" {
						return slackbot.NewUserInputError("Argument CANDIDATE is required")
					}

					pipelines := models.Pipelines{}
					if err := store.Read(db.PipelinesKey, &pipelines); err != nil {
						return err
					}

					pipeline, ok := pipelines.Get(candidateName)
					if !ok {
						return hiringPipelineDoesNotExist(candidateName)
					}

					if pipeline.CurrentStep == 0 {
						return slackbot.NewUserInputError("This pipeline is already on the first step")
					}

					pipeline.CurrentStep -= 1
					if err := store.Write(db.PipelinesKey, pipelines); err != nil {
						return err
					}

					name := strings.Title(candidateName)
					text := fmt.Sprintf("Ok, I've reverted *%s's* hiring pipeline back one step.\n", name)
					text += fmt.Sprintf("The current step is to: `%s`\n", pipeline.Steps[pipeline.CurrentStep])
					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "show",
				Usage:     "show the hiring pipeline for a candidate",
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

					pipeline, ok := pipelines.Get(candidateName)
					if !ok {
						return hiringPipelineDoesNotExist(candidateName)
					}

					var steps string
					for i, step := range pipeline.Steps {
						steps += fmt.Sprintf("%d. %s\n", i+1, step)
					}

					name := strings.Title(candidateName)
					escapedManagerID := slackbot.EscapeUserID(candidate.ManagerID)
					text := fmt.Sprintf("This is the hiring pipeline for *%s*: \n", name)
					text += fmt.Sprintf("```%s```\n", steps)
					text += fmt.Sprintf("*%s's* manager, %s, ", name, escapedManagerID)
					text += fmt.Sprintf("is currently on step *%d* of the hiring pipeline: ", pipeline.CurrentStep+1)
					text += fmt.Sprintf("`%s`", pipeline.Steps[pipeline.CurrentStep])
					return slackbot.WriteString(w, text)
				},
			},
		},
	}
}

func newHiringPipeline(candidateName string) models.Pipeline {
	return models.Pipeline{
		Name: candidateName,
		Type: models.HiringPipelineType,
		Steps: []string{
			"Order hardware (latop, keyboard, mouse, dock, etc.)",
			"Order software (MSDN, etc.)",
			"Grant Access to PSA",
			"Grant Access to TinyPulse",
			"Create personalized success plan",
		},
	}
}

func hiringPipelineDoesNotExist(name string) *slackbot.UserInputError {
	return slackbot.NewUserInputErrorf("There aren't any candidates in a hiring pipeline with the name *%s*", name)
}
