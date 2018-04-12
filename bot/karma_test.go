package bot

import (
	"bytes"
	"context"
	"strconv"
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

func TestKarmaCommand(t *testing.T) {
	store := newMemoryStore(t)
	karma := models.Karma{
		"alpha": models.KarmaEntry{Upvotes: 10, Downvotes: 0},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
		t.Fatal(err)
	}

	w := bytes.NewBuffer(nil)
	cmd := NewKarmaCommand(store, w)
	if err := slackbot.NewTestApp(cmd).Run(strings.Split("iqvbot karma alpha", " ")); err != nil {
		t.Fatal(err)
	}

	output := w.String()
	for name, values := range karma {
		assert.Contains(t, output, name)
		assert.Contains(t, output, strconv.Itoa(values.Downvotes))
		assert.Contains(t, output, strconv.Itoa(values.Upvotes))
	}
}

func TestKarmaCommandWithCountFlag(t *testing.T) {
	store := newMemoryStore(t)
	karma := models.Karma{
		"alpha":   models.KarmaEntry{Upvotes: 1, Downvotes: 0},
		"beta":    models.KarmaEntry{Upvotes: 2, Downvotes: 0},
		"charlie": models.KarmaEntry{Upvotes: 3, Downvotes: 0},
		"delta":   models.KarmaEntry{Upvotes: 4, Downvotes: 0},
		"echo":    models.KarmaEntry{Upvotes: 5, Downvotes: 0},
		"foxtrot": models.KarmaEntry{Upvotes: 6, Downvotes: 0},
		"golf":    models.KarmaEntry{Upvotes: 7, Downvotes: 0},
		"hotel":   models.KarmaEntry{Upvotes: 8, Downvotes: 0},
		"india":   models.KarmaEntry{Upvotes: 9, Downvotes: 0},
		"juliett": models.KarmaEntry{Upvotes: 10, Downvotes: 0},
		"kilo":    models.KarmaEntry{Upvotes: 11, Downvotes: 0},
	}

	if err := store.Write(db.KarmaKey, karma); err != nil {
		t.Fatal(err)
	}

	cases := map[string]struct {
		Input    []string
		Expected int
	}{
		"check default": {
			Input:    strings.Split("iqvbot karma *", " "),
			Expected: 10,
		},
		"count one": {
			Input:    strings.Split("iqvbot karma --count=1 *", " "),
			Expected: 1,
		},
		"count three": {
			Input:    strings.Split("iqvbot karma --count=3 *", " "),
			Expected: 3,
		},
	}

	for name := range cases {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			cmd := NewKarmaCommand(store, w)

			if err := slackbot.NewTestApp(cmd).Run(cases[name].Input); err != nil {
				t.Fatal(err)
			}

			assert.Len(t, strings.Split(w.String(), "\n"), cases[name].Expected)
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

	cases := map[string]struct {
		Input    []string
		Expected []string
	}{
		"enabled": {
			Input:    strings.Split("iqvbot karma --ascending *", " "),
			Expected: []string{"1", "2", "3"},
		},
		"disabled/default": {
			Input:    strings.Split("iqvbot karma *", " "),
			Expected: []string{"3", "2", "1"},
		},
	}

	for name := range cases {
		t.Run(name, func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			cmd := NewKarmaCommand(store, w)

			if err := slackbot.NewTestApp(cmd).Run(cases[name].Input); err != nil {
				t.Fatal(err)
			}

			output := strings.Split(w.String(), "\n")

			for i, entry := range output {
				assert.Contains(t, entry, cases[name].Expected[i])
			}
		})
	}
}

func TestKarmaCommandUserInputErrors(t *testing.T) {
	store := newMemoryStore(t)
	w := bytes.NewBuffer(nil)
	cmd := NewKarmaCommand(store, w)

	cases := map[string][]string{
		"missing GLOB":      strings.Split("iqvbot karma", " "),
		"no matching entry": strings.Split("iqvbot karma *", " "),
	}

	app := slackbot.NewTestApp(cmd)
	for name, args := range cases {
		t.Run(name, func(t *testing.T) {
			assert.IsType(t, &slackbot.UserInputError{}, app.Run(args))
		})
	}
}
