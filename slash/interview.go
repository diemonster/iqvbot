package slash

import (
	"fmt"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
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
	// todo: get requester's time zone
	now := time.Now().UTC()

	msg := slack.Msg{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("Schedule an interview for *%s*", candidate),
		Attachments: []slack.Attachment{
			{
				Text:     "*When is the interview?*",
				Fallback: "You are currently unable to specify when the interview occurs. Please try again later.",
				Color:    "good",
				Actions: []slack.AttachmentAction{
					{
						Name: "day",
						Text: "select a day",
						Type: "select",
						Options: []slack.AttachmentActionOption{
							{
								Text:  "Today",
								Value: now.String(),
							},
							{
								Text:  "Tomorrow",
								Value: now.AddDate(0, 0, 1).String(),
							},
							{
								Text:  "In 2 Days",
								Value: now.AddDate(0, 0, 2).String(),
							},
							{
								Text:  "In 3 Days",
								Value: now.AddDate(0, 0, 3).String(),
							},
							{
								Text:  "In 4 Days",
								Value: now.AddDate(0, 0, 4).String(),
							},
							{
								Text:  "In 5 Days",
								Value: now.AddDate(0, 0, 5).String(),
							},
							{
								Text:  "In 6 Days",
								Value: now.AddDate(0, 0, 6).String(),
							},
							{
								Text:  "In 7 Days",
								Value: now.AddDate(0, 0, 7).String(),
							},
						},
						SelectedOptions: []slack.AttachmentActionOption{
							{
								Text:  "Today",
								Value: now.String(),
							},
						},
					},
					{
						Name: "time",
						Text: "select a time",
						Type: "select",
						Options: []slack.AttachmentActionOption{
							{
								Text:  "9:00 am",
								Value: time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "10:00 am",
								Value: time.Date(0, 0, 0, 10, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "11:00 am",
								Value: time.Date(0, 0, 0, 11, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "12:00 pm",
								Value: time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "1:00 pm",
								Value: time.Date(0, 0, 0, 13, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "2:00 pm",
								Value: time.Date(0, 0, 0, 14, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "3:00 pm",
								Value: time.Date(0, 0, 0, 15, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "4:00 pm",
								Value: time.Date(0, 0, 0, 16, 0, 0, 0, time.UTC).String(),
							},
							{
								Text:  "5:00 pm",
								Value: time.Date(0, 0, 0, 17, 0, 0, 0, time.UTC).String(),
							},
						},
						SelectedOptions: []slack.AttachmentActionOption{
							{
								Text:  "9:00 am",
								Value: time.Date(0, 0, 0, 9, 0, 0, 0, time.UTC).String(),
							},
						},
					},
				},
			},
			{
				Text:     "*Who is performing the interview?*",
				Fallback: "You are currently unable to select who performs the interview. Please try again later.",
				Color:    "good",
				Actions: []slack.AttachmentAction{
					{
						Name:       "interviewer",
						Text:       "select an interviewer",
						Type:       "select",
						DataSource: "users",
					},
				},
			},
			{
				//Text:       "*todo*",
				Fallback:   "You are currently unable to add more interviewers. Please try again later.",
				Color:      "good",
				CallbackID: "add_interviewer",
				Actions: []slack.AttachmentAction{
					{
						Name:  "add_interviewer",
						Text:  "+ Add Interviewer",
						Type:  "button",
						Style: "primary",
					},
				},
			},
			{
				Text:     "*When should I send out a reminder?*",
				Fallback: "You are currently unable to specify when to send a reminder. Please try again later.",
				Color:    "good",
				Actions: []slack.AttachmentAction{
					{
						Name: "time",
						Text: "select a time",
						Type: "select",
						Options: []slack.AttachmentActionOption{
							{
								Text:  "5 minutes before",
								Value: (time.Minute * 5).String(),
							},
							{
								Text:  "15 minutes before",
								Value: (time.Minute * 15).String(),
							},
							{
								Text:  "30 minutes before",
								Value: (time.Minute * 30).String(),
							},
							{
								Text:  "1 hour before",
								Value: (time.Hour).String(),
							},
							{
								Text:  "1 day before",
								Value: (time.Hour * 24).String(),
							},
						},
						SelectedOptions: []slack.AttachmentActionOption{
							{
								Text:  "5 minutes before",
								Value: (time.Minute * 5).String(),
							},
						},
					},
				},
			},
			{
				Text:       "*Submit*",
				Fallback:   "You are currently unable to submit the interview. Please try again later.",
				Color:      "good",
				CallbackID: "submit",
				Actions: []slack.AttachmentAction{
					{
						Name:  "submit",
						Text:  "Submit",
						Type:  "button",
						Style: "primary",
					},
					{
						Name:  "cancel",
						Text:  "Cancel",
						Type:  "button",
						Style: "danger",
						Confirm: &slack.ConfirmationField{
							Title: "Are you sure?",
							Text:  "This interview will be cancelled.",
						},
					},
				},
			},
		},
	}

	return &slack.Message{Msg: msg}, nil
}

func (cmd *InterviewCommand) list() (*slack.Message, error) {
	return nil, NewSlackMessageError("list is not implemented")
}

func (cmd *InterviewCommand) callback(req slack.AttachmentActionCallback) (*slack.Message, error) {
	switch action := req.Actions[0].Name; action {
	case "add_interviewer":
		return cmd.addInterviewer(req)
	case "submit":
		return cmd.submit(req)
	case "cancel":
		return cmd.cancel(req)
	default:
		return nil, fmt.Errorf("Unexpected callback action '%s' (callbackID: '%s')", action, req.CallbackID)
	}
}

func (cmd *InterviewCommand) addInterviewer(req slack.AttachmentActionCallback) (*slack.Message, error) {
	return nil, NewSlackMessageError("Add Interviewer is currently not implemented")
}

func (cmd *InterviewCommand) submit(req slack.AttachmentActionCallback) (*slack.Message, error) {
	fmt.Printf("Req: %#v\n", req)
	return nil, NewSlackMessageError("Submit is currently not implemented")
}

func (cmd *InterviewCommand) cancel(req slack.AttachmentActionCallback) (*slack.Message, error) {
	msg := slack.Msg{
		Text: "Interview has been cancelled",
	}

	return &slack.Message{Msg: msg}, nil
}

/*
const dateFormat = "01/02"

func NewInterviewCommand(store db.Store) *CommandSchema {
	return &CommandSchema{
		Name:     "/interview",
		Help:     "View/Manage interviews with `/interview`, or add an interview with `/interview add @MANAGER mm/dd INTERVIEWEE`",
		Run:      newInterviewRun(store),
		Callback: newInterviewCallback(store),
	}
}

func newInterviewRun(store db.Store) func(slack.SlashCommand) (*slack.Message, error) {
	return func(req slack.SlashCommand) (*slack.Message, error) {
		args := strings.Split(req.Text, " ")
		switch {
		case len(args) == 0 || args[0] == "":
			return interviewsShow(store)
		case args[0] == "add":
			return interviewAdd(store, args[1:])
		default:
			return nil, NewSlackMessageError("Invalid usage: please use `/interview help` for more information")
		}
	}
}

func newAttachmentForInterview(interviewID string, i models.Interview) slack.Attachment {
	return slack.Attachment{
		Text:       fmt.Sprintf("*%s* on %s (manager: %s)", i.Interviewee, i.Date.Format(dateFormat), i.ManagerName),
		Color:      "#3AA3E3",
		CallbackID: interviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:  "delete",
				Text:  "delete",
				Type:  "button",
				Style: "danger",
			},
		},
	}
}

func interviewsShow(store db.Store) (*slack.Message, error) {
	interviews := models.Interviews{}
	if err := store.Read(models.StoreKeyInterviews, &interviews); err != nil {
		return nil, err
	}

	if len(interviews) == 0 {
		return nil, NewSlackMessageError("I currently don't have any interviews scheduled")
	}

	attachments := make([]slack.Attachment, 0, len(interviews))
	for interviewID, interview := range interviews {
		attachments = append(attachments, newAttachmentForInterview(interviewID, interview))
	}

	msg := &slack.Message{
		Msg: slack.Msg{
			Text:        "Here are the interviews I currently have scheduled:",
			Attachments: attachments,
		},
	}

	return msg, nil
}

func addInterviewChecklistItems(store db.Store, interviewID string, interview models.Interview) error {
	checklists := models.Checklists{}
	if err := store.Read(models.StoreKeyChecklists, &checklists); err != nil {
		return err
	}

	checklist, ok := checklists[interview.ManagerID]
	if !ok {
		checklist = models.Checklist{}
	}

	// todo: use guuiid generator
	// todo: have a time for reminders, or print reminders every day
	checklist = append(checklist,
		models.ChecklistItem{
			ID:     strconv.Itoa(int(now.UnixNano())),
			Text:   fmt.Sprintf("Pre-interview stuff for %s", interview.Interviewee),
			Source: interviewID,
		},
		models.ChecklistItem{
			ID:     strconv.Itoa(int(now.UnixNano())),
			Text:   fmt.Sprintf("Post-interview stuff for %s", interview.Interviewee),
			Source: interviewID,
		},
	)

	checklists[interview.ManagerID] = checklist
	return store.Write(models.StoreKeyChecklists, checklists)
}

func interviewAdd(store db.Store, args []string) (*slack.Message, error) {
	if len(args) < 3 {
		return nil, NewSlackMessageError("@MANAGER DATE and INTERVIEWEE are required")
	}

	managerID, managerName, err := parseEscapedUser(args[0])
	if err != nil {
		return nil, NewSlackMessageError("Invalid MANAGER: specify a manager by typing `@<username>`")
	}

	date, err := time.Parse(dateFormat, args[1])
	if err != nil {
		return nil, NewSlackMessageError("Invalid DATE: %v", err)
	}

	interviews := models.Interviews{}
	if err := store.Read(models.StoreKeyInterviews, &interviews); err != nil {
		return nil, err
	}

	// todo: use guid generator
	interviewID := strconv.Itoa(int(now.UnixNano()))
	interview := models.Interview{
		ManagerID:   managerID,
		ManagerName: managerName,
		Interviewee: strings.Join(args[2:], " "),
		Date:        date,
	}

	interviews[interviewID] = interview
	if err := store.Write(models.StoreKeyInterviews, interviews); err != nil {
		return nil, err
	}

	if err := addInterviewChecklistItems(store, interviewID, interview); err != nil {
		return nil, err
	}

	msg := &slack.Message{
		Msg: slack.Msg{
			Text: fmt.Sprintf("Ok, I've added an interview for %s on %s", interview.Interviewee, date.Format(dateFormat)),
		},
	}

	return msg, nil
}

// this function is idempotent
func deleteInterviewChecklistItems(store db.Store, managerID, interviewID string) error {
	checklists := models.Checklists{}
	if err := store.Read(models.StoreKeyChecklists, &checklists); err != nil {
		return err
	}

	checklist, ok := checklists[managerID]
	if !ok {
		return nil
	}

	for i := 0; i < len(checklist); i++ {
		if checklist[i].Source == interviewID {
			checklist = append(checklist[:i], checklist[i+1:]...)
			i--
		}
	}

	checklists[managerID] = checklist
	return store.Write(models.StoreKeyChecklists, checklists)
}

func newInterviewCallback(store db.Store) func(slack.AttachmentActionCallback) (*slack.Message, error) {
	return func(req slack.AttachmentActionCallback) (*slack.Message, error) {
		interviewID := req.CallbackID
		// todo: delete checklist items

		interviews := models.Interviews{}
		if err := store.Read(models.StoreKeyInterviews, &interviews); err != nil {
			return nil, err
		}

		interview, ok := interviews[interviewID]
		if !ok {
			return nil, NewSlackMessageError("That interview no longer exists!")
		}

		if err := deleteInterviewChecklistItems(store, interview.ManagerID, interviewID); err != nil {
			return nil, err
		}

		delete(interviews, interviewID)
		if err := store.Write(models.StoreKeyInterviews, interviews); err != nil {
			return nil, err
		}

		return interviewsShow(store)
	}
}
*/
