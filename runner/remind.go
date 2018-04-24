package runner

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	"github.com/zpatrick/slackbot"
)

// The hour and minute to send hiring pipeline reminders
const (
	HiringPipelineReminderHour   = 9
	HiringPipelineReminderMinute = 0
)

// NewReminderRunner will return a runner that will send reminders to slack users.
// Each time the runner executes, it will read from the store and
func NewReminderRunner(store db.Store, client slackbot.SlackClient) *Runner {
	timers := []*time.Timer{}
	return &Runner{
		Name: "Remind",
		run: func() error {
			hiringPipelineTimers, err := getHiringPipelineTimers(store, client)
			if err != nil {
				return err
			}

			interviewTimers, err := getInterviewTimers(store, client)
			if err != nil {
				return err
			}

			// stop all of our timers before overwriting them
			for i := 0; i < len(timers); i++ {
				timers[i].Stop()
			}

			timers = append(hiringPipelineTimers, interviewTimers...)
			return nil
		},
	}
}

func getHiringPipelineTimers(store db.Store, client slackbot.SlackClient) ([]*time.Timer, error) {
	pipelines := models.Pipelines{}
	if err := store.Read(db.PipelinesKey, &pipelines); err != nil {
		return nil, err
	}

	candidates := models.Candidates{}
	if err := store.Read(db.CandidatesKey, &candidates); err != nil {
		return nil, err
	}

	pipelines.FilterByType(models.HiringPipelineType)
	for i := 0; i < len(pipelines); i++ {
		if pipelines[i].CurrentStep >= len(pipelines[i].Steps) {
			pipelines = append(pipelines[:i], pipelines[i+1:]...)
			i--
		}
	}

	// set a reminder for today or tomorrow if the remind time has already passed
	now := time.Now()
	remindTime := time.Date(now.Year(), now.Month(), now.Day(), HiringPipelineReminderHour, HiringPipelineReminderMinute, 0, 0, time.Local)
	if now.After(remindTime) {
		remindTime.AddDate(0, 0, 1)
	}

	d := time.Until(remindTime)

	timers := make([]*time.Timer, len(pipelines))
	for i := 0; i < len(pipelines); i++ {
		pipeline := pipelines[i]
		candidate, ok := candidates.Get(pipeline.Name)
		if !ok {
			return nil, fmt.Errorf("could not find candidate %s", pipeline.Name)
		}

		timer := time.AfterFunc(d, func() {
			text := "Hello! Just reminding you to "
			text += fmt.Sprintf("finish the hiring pipeline for *%s*. \n", strings.Title(candidate.Name))
			text += fmt.Sprintf("You can view the current step by running `!hire show %s`\n", candidate.Name)
			text += fmt.Sprintf("You can mark a step as complete by running `!hire next %s`\n", candidate.Name)

			_, _, channelID, err := client.OpenIMChannel(candidate.ManagerID)
			if err != nil {
				log.Printf("[ERROR] [Reminder] %v", err)
			}

			if _, _, _, err := client.SendMessage(channelID, slack.MsgOptionText(text, true)); err != nil {
				log.Printf("[ERROR] [Reminder] %v", err)
			}
		})

		timers[i] = timer
	}

	return timers, nil
}

func getInterviewTimers(store db.Store, client slackbot.SlackClient) ([]*time.Timer, error) {
	interviews := models.Interviews{}
	if err := store.Read(db.InterviewsKey, &interviews); err != nil {
		return nil, err
	}

	timers := []*time.Timer{}
	for i := 0; i < len(interviews); i++ {
		interview := interviews[i]
		d := time.Until(interview.Time) - interview.Reminder
		if d <= 0 {
			continue
		}

		fmt.Printf("Sending reminder for %s in %v\n", interview.Candidate, d)

		timer := time.AfterFunc(d, func() {
			for _, interviewerID := range interview.InterviewerIDs {
				text := "Hello! Just reminding you that "
				text += fmt.Sprintf("you have an interview with *%s* ", interview.Candidate)
				text += fmt.Sprintf(" in %d minutes", int(interview.Reminder.Minutes()))

				_, _, channelID, err := client.OpenIMChannel(interviewerID)
				if err != nil {
					log.Printf("[ERROR] [Reminder] %v", err)
				}

				if _, _, _, err := client.SendMessage(channelID, slack.MsgOptionText(text, true)); err != nil {
					log.Printf("[ERROR] [Reminder] %v", err)
				}
			}
		})

		timers = append(timers, timer)
	}

	return timers, nil
}
