package bot

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	"github.com/urfave/cli"
	"github.com/zpatrick/slackbot"
)

// date and time layouts
const (
	DateLayout       = "01/02/2006"
	TimeLayout       = "03:04PM"
	DateTimeLayout   = DateLayout + " " + TimeLayout
	DateAtTimeLayout = DateLayout + " at " + TimeLayout
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
					args := c.Args()
					candidateName := args.Get(0)
					if candidateName == "" {
						return fmt.Errorf("Argument CANDIDATE is required")
					}

					dateStr := args.Get(1)
					if dateStr == "" {
						return fmt.Errorf("Argument DATE is required")
					}

					timeStr := strings.ToUpper(args.Get(2))
					if timeStr == "" {
						return fmt.Errorf("Argument TIME is required")
					}

					interviewers := args[3:]
					if len(interviewers) == 0 {
						return fmt.Errorf("At least one interviewer is required")
					}

					t, err := time.ParseInLocation(DateTimeLayout, dateStr+" "+timeStr, time.Local)
					if err != nil {
						return err
					}

					interviewerIDs := make([]string, len(interviewers))
					for i := 0; i < len(interviewers); i++ {
						interviewerID, err := slackbot.ParseUserID(interviewers[i])
						if err != nil {
							return err
						}

						interviewerIDs[i] = interviewerID
					}

					candidates := models.Candidates{}
					if err := store.Read(db.CandidatesKey, &candidates); err != nil {
						return err
					}

					if _, ok := candidates.Get(candidateName); !ok {
						text := fmt.Sprintf("I don't have any candidates by the name *%s*\n", candidateName)
						text += "You can create a candidate by running `!candidate add NAME @MANAGER`"
						return slackbot.NewUserInputError(text)
					}

					interviews := models.Interviews{}
					if err := store.Read(db.InterviewsKey, &interviews); err != nil {
						return err
					}

					interview := models.Interview{
						Candidate:      candidateName,
						Time:           t.UTC(),
						InterviewerIDs: interviewerIDs,
					}

					interviews = append(interviews, interview)
					if err := store.Write(db.InterviewsKey, interviews); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've added an interview for *%s* on %s",
						strings.Title(interview.Candidate),
						interview.Time.In(time.Local).Format(DateAtTimeLayout))
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
					interviews := models.Interviews{}
					if err := store.Read(db.InterviewsKey, &interviews); err != nil {
						return err
					}

					if len(interviews) == 0 {
						return slackbot.WriteString(w, "I don't have any interviews scheduled")
					}

					interviews.Sort(!c.Bool("ascending"))

					var text string
					for i := 0; i < len(interviews) && i < c.Int("limit"); i++ {
						dateAtTime := interviews[i].Time.In(time.Local).Format(DateAtTimeLayout)
						text += fmt.Sprintf("*%s* on %s with ", strings.Title(interviews[i].Candidate), dateAtTime)
						for _, interviewerID := range interviews[i].InterviewerIDs {
							text += fmt.Sprintf("<@%s> ", interviewerID)
						}

						text += "\n"
					}

					return slackbot.WriteString(w, text)
				},
			},
			{
				Name:      "rm",
				Usage:     "remove an interview",
				ArgsUsage: "CANDIDATE DATE (mm/dd/yyyy) TIME (mm:hh{am|pm})",
				Action: func(c *cli.Context) error {
					args := c.Args()
					candidateName := args.Get(0)
					if candidateName == "" {
						return fmt.Errorf("Argument CANDIDATE is required")
					}

					dateStr := args.Get(1)
					if dateStr == "" {
						return fmt.Errorf("Argument DATE is required")
					}

					timeStr := strings.ToUpper(args.Get(2))
					if timeStr == "" {
						return fmt.Errorf("Argument TIME is required")
					}

					t, err := time.ParseInLocation(DateTimeLayout, dateStr+" "+timeStr, time.Local)
					if err != nil {
						return err
					}

					interviews := models.Interviews{}
					if err := store.Read(db.InterviewsKey, &interviews); err != nil {
						return err
					}

					interview := models.Interview{
						Candidate: candidateName,
						Time:      t.UTC(),
					}

					var exists bool
					for i := 0; i < len(interviews); i++ {
						if interviews[i].Equals(interview) {
							exists = true
							interviews = append(interviews[:i], interviews[i+1:]...)
							i--
						}
					}

					if !exists {
						text := "I couldn't find any interviews matching the specified candidate and time"
						return slackbot.NewUserInputError(text)
					}

					if err := store.Write(db.InterviewsKey, interviews); err != nil {
						return err
					}

					return slackbot.WriteStringf(w, "Ok, I've deleted *%s's* interview on %s",
						strings.Title(candidateName),
						t.In(time.Local).Format(DateLayout))
				},
			},
		},
	}
}
