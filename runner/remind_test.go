package runner

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	"github.com/zpatrick/slackbot/mock_slack"
)

func TestGetHiringPipelineTimers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSlackClient := mock_slack.NewMockSlackClient(ctrl)

	candidates := models.Candidates{
		{
			Name:      "John Doe",
			ManagerID: "uid",
		},
	}

	pipelines := models.Pipelines{
		{
			Name:  "John Doe",
			Type:  models.HiringPipelineType,
			Steps: []string{"one", "two", "three"},
		},
	}

	store := newMemoryStore(t)
	if err := store.Write(db.CandidatesKey, candidates); err != nil {
		t.Fatal(err)
	}

	if err := store.Write(db.PipelinesKey, pipelines); err != nil {
		t.Fatal(err)
	}

	c := make(chan bool)
	record := func(channel string, options ...slack.MsgOption) {
		c <- true
	}

	mockSlackClient.EXPECT().
		OpenIMChannel("uid").
		Return(false, false, "cid", nil)

	mockSlackClient.EXPECT().
		SendMessage("cid", gomock.Any()).
		Do(record).
		Return("", "", "", nil)

	timers, err := getHiringPipelineTimers(store, mockSlackClient)
	if err != nil {
		t.Fatal(err)
	}

	// reset the timers to execute immediately
	for _, timer := range timers {
		timer.Reset(0)
	}

	select {
	case <-c:
	case <-time.After(time.Second):
		t.Fatal("timeout")
	}
}

func TestGetInterviewTimers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockSlackClient := mock_slack.NewMockSlackClient(ctrl)

	now := time.Now()
	interviews := models.Interviews{
		{Time: now.Add(time.Hour), Reminder: time.Minute, InterviewerIDs: []string{"uid1", "uid2"}},
		{Time: now.Add(time.Hour), Reminder: time.Minute * 2, InterviewerIDs: []string{"uid3"}},
		{Time: now.Add(-time.Hour), InterviewerIDs: []string{"bad"}},
	}

	store := newMemoryStore(t)
	if err := store.Write(db.InterviewsKey, interviews); err != nil {
		t.Fatal(err)
	}

	c := make(chan bool)
	record := func(channel string, options ...slack.MsgOption) {
		c <- true
	}

	for _, id := range []string{"uid1", "uid2", "uid3"} {
		channelID := "d" + id
		mockSlackClient.EXPECT().
			OpenIMChannel(id).
			Return(false, false, channelID, nil)

		mockSlackClient.EXPECT().
			SendMessage(channelID, gomock.Any()).
			Do(record).
			Return("", "", "", nil)
	}

	timers, err := getInterviewTimers(store, mockSlackClient)
	if err != nil {
		t.Fatal(err)
	}

	// reset the timers to execute immediately
	for _, timer := range timers {
		timer.Reset(0)
	}

	// expect 3 calls: one for "uid1", "uid2", and "uid3"
	for i := 0; i < 3; i++ {
		select {
		case <-c:
		case <-time.After(time.Second):
			t.Fatalf("Timeout on index %d", i)
		}
	}
}
