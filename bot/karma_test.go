package bot

import (
	"context"
	"testing"

	"github.com/nlopes/slack"
	"github.com/quintilesims/slackbot/db"
	"github.com/quintilesims/slackbot/models"
	"github.com/stretchr/testify/assert"
	"github.com/zpatrick/slackbot"
)

func TestKarmaBehavior(t *testing.T) {
	store := db.NewMemoryStore()
	karmas := models.Karmas{
		"dogs": models.Karma{Upvotes: 10, Downvotes: 0},
		"cats": models.Karma{Upvotes: 0, Downvotes: 10},
	}

	if err := store.Write(db.KarmasKey, karmas); err != nil {
		t.Fatal(err)
	}

	events := []slack.RTMEvent{
		slackbot.NewMessageRTMEvent("dogs++"),
		slackbot.NewMessageRTMEvent("dogs++"),
		slackbot.NewMessageRTMEvent("cats--"),
		slackbot.NewMessageRTMEvent("cats--"),
		slackbot.NewMessageRTMEvent("new++"),
		slackbot.NewMessageRTMEvent("new--"),
		slackbot.NewMessageRTMEvent("new+-"),
		slackbot.NewMessageRTMEvent("new-+"),
		slackbot.NewMessageRTMEvent("blah blah"),
		{},
	}

	b := NewKarmaBehavior(store)
	for _, e := range events {
		if err := b(context.Background(), e); err != nil {
			t.Fatal(err)
		}
	}

	result := models.Karmas{}
	if err := store.Read(db.KarmasKey, &result); err != nil {
		t.Fatal(err)
	}

	expected := models.Karmas{
		"dogs": models.Karma{Upvotes: 12, Downvotes: 0},
		"cats": models.Karma{Upvotes: 0, Downvotes: 12},
		"new":  models.Karma{Upvotes: 3, Downvotes: 3},
	}

	assert.Equal(t, expected, result)
}

// TestKarmaCommandDefaults
// TestKarmaCommandWithCountFlag
// TestKarmaCommandWithAscendingFlag
// TestKarmaCommandUserInputErrors
func TestKarmaCommandWithDefaults(t *testing.T) {
}
