package slash

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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

	n := time.Now().In(PDT)
	interview := &models.Interview{
		InterviewID:    randomString(10),
		Candidate:      candidate,
		InterviewerIDs: []string{req.UserID},
		Time:           time.Date(n.Year(), n.Month(), n.Day(), 9, 0, 0, 0, n.Location()),
		Reminder:       time.Minute * 5,
	}

	interviews = append(interviews, interview)
	if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
		return nil, err
	}

	return AddInterviewView(*interview), nil
}

func (cmd *InterviewCommand) list() (*slack.Message, error) {
	interviews := models.Interviews{}
	if err := cmd.store.Read(db.InterviewsKey, &interviews); err != nil {
		return nil, err
	}

	return ListInterviewsView(interviews), nil
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

	switch actionName := req.Actions[0].Name; {
	case actionName == ActionAddInterviewer:
		interview.InterviewerIDs = append(interview.InterviewerIDs, "")
	case actionName == ActionSelectDate:
		updated, err := time.Parse(TimeValueFormat, req.Actions[0].SelectedOptions[0].Value)
		if err != nil {
			return nil, err
		}

		// only update the date, not the time
		interview.Time = time.Date(
			updated.Year(),
			updated.Month(),
			updated.Day(),
			interview.Time.Hour(),
			interview.Time.Minute(),
			interview.Time.Second(),
			interview.Time.Nanosecond(),
			interview.Time.Location())
	case actionName == ActionSelectTime:
		updated, err := time.Parse(TimeValueFormat, req.Actions[0].SelectedOptions[0].Value)
		if err != nil {
			return nil, err
		}

		// only update the time, not the date
		interview.Time = time.Date(
			interview.Time.Year(),
			interview.Time.Month(),
			interview.Time.Day(),
			updated.Hour(),
			updated.Minute(),
			updated.Second(),
			updated.Nanosecond(),
			updated.Location())
	case actionName == ActionSelectReminder:
		d, err := time.ParseDuration(req.Actions[0].SelectedOptions[0].Value)
		if err != nil {
			return nil, err
		}

		interview.Reminder = d
	case strings.HasPrefix(actionName, ActionSelectInterviewer):
		index, err := strconv.Atoi(actionName[len(actionName)-1:])
		if err != nil {
			return nil, err
		}

		interview.InterviewerIDs[index] = req.Actions[0].SelectedOptions[0].Value
	case actionName == ActionSchedule:
		if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
			return nil, err
		}

		msg := slack.Msg{
			Text: fmt.Sprintf("Interview for *%s* on *%s* at *%s* has been scheduled!",
				interview.Candidate,
				interview.Time.Format(DateDisplayFormat),
				interview.Time.Format(TimeDisplayFormat)),
		}

		return &slack.Message{Msg: msg}, nil
	case actionName == ActionCancel || actionName == ActionDelete:
		for i := 0; i < len(interviews); i++ {
			if interviews[i].InterviewID == interviewID {
				interviews = append(interviews[:i], interviews[i+1:]...)
				i--
			}
		}

		if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
			return nil, err
		}

		if actionName == ActionDelete {
			return ListInterviewsView(interviews), nil
		}

		msg := slack.Msg{
			Text: fmt.Sprintf("Interview for *%s* has been cancelled", interview.Candidate),
		}

		return &slack.Message{Msg: msg}, nil
	default:
		return nil, fmt.Errorf("Unexpected callback action name '%s'", actionName)
	}

	if err := cmd.store.Write(db.InterviewsKey, interviews); err != nil {
		return nil, err
	}

	return AddInterviewView(*interview), nil
}
