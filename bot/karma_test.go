package bot

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	"github.com/stretchr/testify/assert"
	"github.com/zpatrick/slackbot"
)

func TestKarmaBehavior(t *testing.T) {
	store := db.NewMemoryStore()
	karma := models.Karma{
		"dogs": models.KarmaEntry{Upvotes: 10, Downvotes: 0},
		"cats": models.KarmaEntry{Upvotes: 0, Downvotes: 10},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
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

	result := models.Karma{}
	if err := store.Read(db.KarmaKey, &result); err != nil {
		t.Fatal(err)
	}

	expected := models.Karma{
		"dogs": {Upvotes: 12, Downvotes: 0},
		"cats": {Upvotes: 0, Downvotes: 12},
		"new":  {Upvotes: 3, Downvotes: 3},
	}

	assert.Equal(t, expected, result)
}

func TestKarmaCommandDefaults(t *testing.T) {
	store := newMemoryStore(t)

	karma := models.Karma{
		"alpha":   models.KarmaEntry{Upvotes: 11, Downvotes: 0},
		"beta":    models.KarmaEntry{Upvotes: 10, Downvotes: 0},
		"charlie": models.KarmaEntry{Upvotes: 9, Downvotes: 0},
		"delta":   models.KarmaEntry{Upvotes: 8, Downvotes: 0},
		"echo":    models.KarmaEntry{Upvotes: 7, Downvotes: 0},
		"foxtrot": models.KarmaEntry{Upvotes: 6, Downvotes: 0},
		"golf":    models.KarmaEntry{Upvotes: 5, Downvotes: 0},
		"hotel":   models.KarmaEntry{Upvotes: 4, Downvotes: 0},
		"india":   models.KarmaEntry{Upvotes: 3, Downvotes: 0},
		"juliet":  models.KarmaEntry{Upvotes: 2, Downvotes: 0},
		"kilo":    models.KarmaEntry{Upvotes: 1, Downvotes: 0},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		Input    []string
		Expected []string
	}{
		"exact match": {
			Input:    strings.Split("iqvbot karma alpha", " "),
			Expected: []string{"alpha"},
		},
		"wildcard preceding": {
			Input:    strings.Split("iqvbot karma *lpha", " "),
			Expected: []string{"alpha"},
		},
		"wildcard tailing": {
			Input:    strings.Split("iqvbot karma alph*", " "),
			Expected: []string{"alpha"},
		},
		"wildcard multi-match": {
			Input:    strings.Split("iqvbot karma *a", " "),
			Expected: []string{"alpha", "beta", "delta", "india"},
		},
		"two wildcards": {
			Input:    strings.Split("iqvbot karma *e*", " "),
			Expected: []string{"beta", "charlie", "delta", "echo", "hotel", "juliet"},
		},
		"count default": {
			Input:    strings.Split("iqvbot karma *", " "),
			Expected: []string{"alpha", "beta", "charlie", "delta", "echo", "foxtrot", "golf", "hotel", "india", "juliet"},
		},
		"descending default": {
			Input:    strings.Split("iqvbot karma *o", " "),
			Expected: []string{"echo", "kilo"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			cmd := NewKarmaCommand(store, w)

			if err := slackbot.NewTestApp(cmd).Run(c.Input); err != nil {
				t.Fatal(err)
			}

			output := w.String()
			for _, e := range c.Expected {
				assert.Contains(t, output, e)
			}
		})
	}
}

func TestKarmaCommandWithCountFlag(t *testing.T) {
	store := newMemoryStore(t)
	karma := models.Karma{
		"alpha":   models.KarmaEntry{Upvotes: 4, Downvotes: 0},
		"beta":    models.KarmaEntry{Upvotes: 3, Downvotes: 0},
		"charlie": models.KarmaEntry{Upvotes: 2, Downvotes: 0},
		"delta":   models.KarmaEntry{Upvotes: 1, Downvotes: 0},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		Input    []string
		Expected []string
	}{
		"count one": {
			Input:    strings.Split("iqvbot karma --count=1 *", " "),
			Expected: []string{"alpha"},
		},
		"count three": {
			Input:    strings.Split("iqvbot karma --count=3 *", " "),
			Expected: []string{"alpha", "beta", "charlie"},
		},
	}

	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			cmd := NewKarmaCommand(store, w)

			if err := slackbot.NewTestApp(cmd).Run(c.Input); err != nil {
				t.Fatal(err)
			}

			output := strings.Split(w.String(), "\n")
			for i, e := range c.Expected {
				assert.Contains(t, output[i], e)
			}
		})
	}
}

func TestKarmaCommandWithAscendingFlag(t *testing.T) {
	store := newMemoryStore(t)
	karma := models.Karma{
		"alpha":   models.KarmaEntry{Upvotes: 1, Downvotes: 0},
		"beta":    models.KarmaEntry{Upvotes: 2, Downvotes: 0},
		"charlie": models.KarmaEntry{Upvotes: 3, Downvotes: 0},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
		t.Fatal(err)
	}

	w := bytes.NewBuffer(nil)
	cmd := NewKarmaCommand(store, w)

	if err := slackbot.NewTestApp(cmd).Run(strings.Split("iqvbot karma --ascending *", " ")); err != nil {
		t.Fatal(err)
	}

	output := w.String()
	var expected string
	for name, entry := range karma {
		expected += fmt.Sprintf("*%s*: %d (%d upvotes, %d downvotes)\n",
			name,
			entry.Upvotes-entry.Downvotes,
			entry.Upvotes,
			entry.Downvotes)
	}

	assert.Equal(t, expected, output)
}

func TestKarmaCommandUserInputErrors(t *testing.T) {
	cases := map[string][]string{
		"missing GLOB":      strings.Split("iqvbot karma", " "),
		"no matching entry": strings.Split("iqvbot karma *", " "),
	}

	app := slackbot.NewTestApp(NewKarmaCommand(newMemoryStore(t), ioutil.Discard))
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			assert.IsType(t, &slackbot.UserInputError{}, app.Run(args))
		})
	}
}
