package slash

import (
	"fmt"
	"time"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/models"
)

const (
	ActionSelectDate        = "select_date"
	ActionSelectTime        = "select_time"
	ActionAddInterviewer    = "add_interviewer"
	ActionSelectReminder    = "select_reminder"
	ActionSelectInterviewer = "select_interviewer"
	ActionSchedule          = "schedule"
	ActionCancel            = "cancel"
	ActionDelete            = "delete"
	DateDisplayFormat       = "Monday, January 2"
	TimeDisplayFormat       = "3:04 PM MST"
	TimeValueFormat         = "2006-01-02 15:04:05 -0700 MST"
)

func AddInterviewView(interview models.Interview) *slack.Message {
	view := slack.Msg{
		ResponseType: "in_channel",
		Text:         fmt.Sprintf("Schedule an interview for *%s*", interview.Candidate),
		Attachments: []slack.Attachment{
			selectDateAttachment(interview),
			selectTimeAttachment(interview),
			selectReminderAttachment(interview),
			selectInterviewersAttachment(interview),
			addInterviewerAttachment(interview),
			scheduleAttachment(interview),
		},
	}

	return &slack.Message{Msg: view}
}

func selectDateAttachment(interview models.Interview) slack.Attachment {
	options := make([]slack.AttachmentActionOption, 14)
	for i := 0; i < len(options); i++ {
		t := time.Now().In(PST).AddDate(0, 0, i)
		options[i] = slack.AttachmentActionOption{
			Text:  t.Format(DateDisplayFormat),
			Value: t.Format(TimeValueFormat),
		}
	}

	selectedOptions := []slack.AttachmentActionOption{
		{
			Text:  interview.Time.Format(DateDisplayFormat),
			Value: interview.Time.Format(TimeValueFormat),
		},
	}

	return slack.Attachment{
		Text:       "*What day is the interview?*",
		Fallback:   "You are currently unable to specify the date of the interview. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:            ActionSelectDate,
				Text:            "select a date",
				Type:            "select",
				Options:         options,
				SelectedOptions: selectedOptions,
			},
		},
	}
}

func selectTimeAttachment(interview models.Interview) slack.Attachment {
	minutes := []int{0, 15, 30, 45}
	options := make([]slack.AttachmentActionOption, 24*len(minutes))

	for hour := 0; hour < 24; hour++ {
		for j, minute := range minutes {
			t := time.Date(interview.Time.Year(), interview.Time.Month(),
				interview.Time.Day(), hour, minute, 0, 0, PST)

			options[(hour*len(minutes))+j] = slack.AttachmentActionOption{
				Text:  t.Format(TimeDisplayFormat),
				Value: t.Format(TimeValueFormat),
			}
		}
	}

	selectedOptions := []slack.AttachmentActionOption{
		{
			Text:  interview.Time.Format(TimeDisplayFormat),
			Value: interview.Time.Format(TimeValueFormat),
		},
	}

	return slack.Attachment{
		Text:       "*What time is the interview?*",
		Fallback:   "You are currently unable to specify the time of the interview. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:            ActionSelectTime,
				Text:            "select a time",
				Type:            "select",
				Options:         options,
				SelectedOptions: selectedOptions,
			},
		},
	}
}

func selectReminderAttachment(interview models.Interview) slack.Attachment {
	minutes := []int{5, 15, 30, 60}
	options := make([]slack.AttachmentActionOption, len(minutes))
	for i := 0; i < len(options); i++ {
		options[i] = slack.AttachmentActionOption{
			Text:  fmt.Sprintf("%d minutes before", minutes[i]),
			Value: (time.Minute * time.Duration(minutes[i])).String(),
		}
	}

	selectedOptions := []slack.AttachmentActionOption{
		{
			Text:  fmt.Sprintf("%d minutes before", int(interview.Reminder.Minutes())),
			Value: interview.Reminder.String(),
		},
	}

	return slack.Attachment{
		Text:       "*When should I send a reminder?*",
		Fallback:   "You are currently unable to specify when to send a reminder. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:            ActionSelectReminder,
				Text:            "select a time",
				Type:            "select",
				Options:         options,
				SelectedOptions: selectedOptions,
			},
		},
	}
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

func scheduleAttachment(interview models.Interview) slack.Attachment {
	return slack.Attachment{
		Text:       "*Schedule*",
		Fallback:   "You are currently unable to schedule the interview. Please try again later.",
		Color:      "good",
		CallbackID: interview.InterviewID,
		Actions: []slack.AttachmentAction{
			{
				Name:  ActionSchedule,
				Text:  "Schedule",
				Type:  "button",
				Style: "primary",
			},
			{
				Name:  ActionCancel,
				Text:  "Cancel",
				Type:  "button",
				Style: "danger",
				Confirm: &slack.ConfirmationField{
					Title: "Are you sure?",
					Text:  "This interview will be cancelled.",
				},
			},
		},
	}
}

func ListInterviewsView(interviews models.Interviews) *slack.Message {
	if len(interviews) == 0 {
		view := slack.Msg{
			ResponseType: "in_channel",
			Text:         "There are no interviews currently scheduled",
		}

		return &slack.Message{Msg: view}
	}

	attachments := make([]slack.Attachment, len(interviews))
	for i := 0; i < len(attachments); i++ {
		interview := interviews[i]

		var interviewersText string
		for _, userID := range interview.InterviewerIDs {
			interviewersText += fmt.Sprintf("<@%s> ", userID)
		}

		attachments[i] = slack.Attachment{
			Pretext:    fmt.Sprintf("*%s*", interview.Candidate),
			Fallback:   "You are currently unable to view this interview. Please try again later.",
			Color:      "good",
			CallbackID: interview.InterviewID,
			Fields: []slack.AttachmentField{
				{
					Title: "Date",
					Value: interview.Time.Format(DateDisplayFormat),
					Short: true,
				},
				{
					Title: "Time",
					Value: interview.Time.Format(TimeDisplayFormat),
					Short: true,
				},
				{
					Title: "Interviewers",
					Value: interviewersText,
				},
			},
			Footer: "Careful! This cannot be undone!",
			Actions: []slack.AttachmentAction{
				{
					Name:  ActionDelete,
					Text:  "Delete",
					Type:  "button",
					Style: "danger",
					Confirm: &slack.ConfirmationField{
						Title: "Are you sure?",
						Text:  "This interview will be deleted.",
					},
				},
			},
		}
	}

	view := slack.Msg{
		ResponseType: "in_channel",
		Text:         "Here are the interviews I currently have scheduled: ",
		Attachments:  attachments,
	}

	return &slack.Message{Msg: view}
}
