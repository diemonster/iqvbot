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

// NewCandidateCommand create a cli.Command that allows users to add, update, list, and remove candidates
func NewCandidateCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:  "candidate",
		Usage: "manage candidates",
		Subcommands: []cli.Command{
			{
				Name:      "add",
				Usage:     "add a new candidate",
				ArgsUsage: "NAME @MANAGER",
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "meta",
						Usage: "metadata about the candidate in key=val format",
					},
				},
				Action: func(c *cli.Context) error {
					args := c.Args()
					name := args.Get(0)
					if name == "" {
						return slackbot.NewUserInputError("Argument NAME is required")
					}

					manager := args.Get(1)
					if manager == "" {
						return slackbot.NewUserInputError("Argument MANAGER is required")
					}

					managerID, err := slackbot.ParseUserID(manager)
					if err != nil {
						return slackbot.NewUserInputErrorf("'%s' is not in valid @username format", manager)
					}

					meta, err := parseMetaFlag(c.StringSlice("meta"))
					if err != nil {
						return err
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return slackbot.NewUserInputError(err.Error())
					}

					if _, ok := candidates.Get(name); ok {
						return slackbot.NewUserInputErrorf("Candidate with name '%s' already exists", name)
					}

					candidate := &models.Candidate{
						Name:      name,
						ManagerID: managerID,
						Meta:      meta,
					}

					candidates = append(candidates, candidate)
					if err := store.Write(db.CandidatesKey, candidates); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've added a new candidate named *%s*", name)
				},
			},
			{
				Name:  "ls",
				Usage: "list candidates",
				Flags: []cli.Flag{
					cli.IntFlag{
						Name:  "limit",
						Value: 50,
						Usage: "The maximum number of candidates to display",
					},
					cli.BoolFlag{
						Name:  "ascending",
						Usage: "Show results in reverse-alphabetical order",
					},
				},
				Action: func(c *cli.Context) error {
					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					if len(candidates) == 0 {
						return slackbot.WriteString(w, "I don't have any candidates at the moment")
					}

					candidates.Sort(!c.Bool("ascending"))

					text := "Here are the candidates I have: \n"
					for i := 0; i < c.Int("count") && i < len(candidates); i++ {
						text += fmt.Sprintf("*%s* \n", candidates[i].Name)
					}

					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "rm",
				Usage:     "remove a candidate",
				ArgsUsage: "NAME",
				Action: func(c *cli.Context) error {
					name := strings.Join(c.Args(), " ")
					if name == "" {
						return slackbot.NewUserInputErrorf("Argument NAME is required")
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					if ok := candidates.Delete(name); !ok {
						return slackbot.NewUserInputErrorf("I don't have any candidates by the name *%s*", name)
					}

					if err := store.Write(db.CandidatesKey, candidates); err != nil {
						return err
					}

					// delete the candidate's interviews
					interviews := models.Interviews{}
					if err := store.Read(db.InterviewsKey, &interviews); err != nil {
						return err
					}

					for i := 0; i < len(interviews); i++ {
						if strings.ToLower(interviews[i].Candidate) == strings.ToLower(name) {
							interviews = append(interviews[:i], interviews[i+1:]...)
							i--
						}
					}

					if err := store.Write(db.InterviewsKey, interviews); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've deleted candidate *%s*", name)
				},
			},
			{
				Name:      "show",
				Usage:     "show information about a candidate",
				ArgsUsage: "NAME",
				Action: func(c *cli.Context) error {
					name := strings.Join(c.Args(), " ")
					if name == "" {
						return slackbot.NewUserInputError("Argument NAME is required")
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					candidate, ok := candidates.Get(name)
					if !ok {
						return slackbot.NewUserInputErrorf("I don't have any candidates by the name *%s*", name)
					}

					text := fmt.Sprintf("*%s* (manager: %s)\n", candidate.Name, slackbot.EscapeUserID(candidate.ManagerID))
					for key, val := range candidate.Meta {
						text += fmt.Sprintf("*%s*: %s\n", key, val)
					}

					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "update",
				Usage:     "add/update a candidate's metadata",
				ArgsUsage: "NAME KEY VAL",
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "manager",
						Usage: "Update the candidate's manager",
					},
				},
				Action: func(c *cli.Context) error {
					args := c.Args()
					name := args.Get(0)
					if name == "" {
						return slackbot.NewUserInputError("Argument NAME is required")
					}

					key := args.Get(1)
					if key == "" {
						return slackbot.NewUserInputError("Argument KEY is required")
					}

					val := args.Get(2)
					if val == "" {
						return slackbot.NewUserInputError("Argument VAL is required")
					}

					update := func(candidate *models.Candidate) {
						candidate.Meta[key] = val
					}

					if manager := c.String("manager"); manager != "" {
						managerID, err := slackbot.ParseUserID(manager)
						if err != nil {
							return slackbot.NewUserInputErrorf("'%s' is not in valid @username format", manager)
						}

						update = func(candidate *models.Candidate) {
							candidate.Meta[key] = val
							candidate.ManagerID = managerID
						}
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					candidate, ok := candidates.Get(name)
					if !ok {
						return slackbot.NewUserInputErrorf("I don't have any candidates by the name *%s*", name)
					}

					update(candidate)
					if err := store.Write(db.CandidatesKey, candidates); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've updated information for *%s*", name)
				},
			},
		},
	}
}

func parseMetaFlag(inputs []string) (map[string]string, error) {
	meta := map[string]string{}
	for _, input := range inputs {
		split := strings.Split(input, "=")
		if len(split) != 2 {
			return nil, fmt.Errorf("'%s' is not in proper key=val format", input)
		}

		meta[split[0]] = split[1]
	}

	return meta, nil
}
