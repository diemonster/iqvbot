package slash

import (
	"fmt"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/models"
)

const (
	ActionAddInterviewer    = "add_interviewer"
	ActionSelectInterviewer = "select_interviewer"
)

func AddInterviewView(interview models.Interview) *slack.Message {
	view := slack.Msg{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("Schedule an interview for *%s*", interview.Candidate),
		Attachments: []slack.Attachment{
			selectInterviewersAttachment(interview),
			addInterviewerAttachment(interview),
		},
	}

	return &slack.Message{Msg: view}
}

func selectInterviewersAttachment(interview models.Interview) slack.Attachment {
	actions := make([]slack.AttachmentAction, len(interview.InterviewerIDs))
	for i := 0; i < len(actions); i++ {
		actions[i] = slack.AttachmentAction{
			Name:       fmt.Sprintf("%s_%d", ActionSelectInterviewer, i),
			Text:       "select an interviewer",
			Type:       "select",
			DataSource: "users",
			SelectedOptions: []slack.AttachmentActionOption{
				{
					Text:  interview.InterviewerIDs[i],
					Value: interview.InterviewerIDs[i],
				},
			},
		}
	}

	// todo: set values on selected optoins

	return slack.Attachment{
		Text:       "*Who is performing the interview?*",
		Fallback:   "You are currently unable to select who performs the interview. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions:    actions,
	}
}

func addInterviewerAttachment(interview models.Interview) slack.Attachment {
	return slack.Attachment{
		Fallback:   "You are currently unable to add more interviewers. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:  ActionAddInterviewer,
				Text:  "+ Add Interviewer",
				Type:  "button",
				Style: "primary",
			},
		},
	}
}
