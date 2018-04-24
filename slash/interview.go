package slash

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
)

type InterviewCommand struct {
	store db.Store
}

func NewInterviewCommand(store db.Store) *InterviewCommand {
	return &InterviewCommand{
		store: store,
	}
}

func (cmd *InterviewCommand) Schema() *CommandSchema {
	return &CommandSchema{
		Name:     "/interview",
		Help:     "View/Manage interviews with `/interview`, or add an interview with `/interview add`",
		Run:      cmd.run,
		Callback: cmd.callback,
	}
}

func (cmd *InterviewCommand) run(req slack.SlashCommand) (*slack.Message, error) {
	args := strings.Split(req.Text, " ")
	switch {
	case len(args) == 0 || args[0] == "":
		return cmd.list()
	case args[0] == "add":
		candidate := strings.Join(args[1:], " ")
		if candidate == "" {
			return nil, NewSlackMessageError("Invalid usage: please specify the candidate's name using `/interview add NAME`")
		}

		return cmd.add(req, strings.Title(candidate))
	default:
		return nil, NewSlackMessageError("Invalid usage: please use `/interview help` for more information")
	}
}

func (cmd *InterviewCommand) add(req slack.SlashCommand, candidate string) (*slack.Message, error) {
	interviews := models.Interviews{}
	if err := cmd.store.Read(db.InterviewsKey, &interviews); err != nil {
		return nil, err
	}

	interview := &models.Interview{
		InterviewID:    randomString(10),
		Candidate:      candidate,
		InterviewerIDs: make([]string, 1),
	}

	interviews = append(interviews, interview)
	if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
		return nil, err
	}

	return AddInterviewView(*interview), nil
}

func (cmd *InterviewCommand) list() (*slack.Message, error) {
	return nil, NewSlackMessageError("list is not implemented")
}

func (cmd *InterviewCommand) callback(req slack.AttachmentActionCallback) (*slack.Message, error) {
	interviews := models.Interviews{}
	if err := cmd.store.Read(db.InterviewsKey, &interviews); err != nil {
		return nil, err
	}

	interviewID := req.CallbackID
	interview, ok := interviews.Get(interviewID)
	if !ok {
		return nil, NewSlackMessageError("This interview no longer exists!")
	}

	switch action := req.Actions[0].Name; {
	case action == ActionAddInterviewer:
		interview.InterviewerIDs = append(interview.InterviewerIDs, "")
	case strings.HasPrefix(action, ActionSelectInterviewer):
		index, err := strconv.Atoi(action[len(action)-1:])
		if err != nil {
			return nil, err
		}

		interview.InterviewerIDs[index] = req.Actions[0].SelectedOptions[0].Value
	default:
		return nil, fmt.Errorf("Unexpected callback action '%s'", action)
	}

	if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
		return nil, err
	}

	return AddInterviewView(*interview), nil
}
