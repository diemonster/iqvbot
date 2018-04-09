package behaviors

import (
	"context"
	"strings"

	"github.com/nlopes/slack"
	"github.com/quintilesims/slackbot/db"
	"github.com/quintilesims/slackbot/models"
	"github.com/zpatrick/slackbot"
)

// NewKarmaBehavior returns a behavior that updates karma in the provided store.
// Karma updates are triggered by the presence of '++', '--', '+-', or '-+' at the end of a message.
func NewKarmaBehavior(store db.Store) slackbot.Behavior {
	return func(ctx context.Context, e slack.RTMEvent) error {
		d, ok := e.Data.(*slack.MessageEvent)
		if !ok {
			return nil
		}

		var update func(k models.Karma) models.Karma
		switch {
		case strings.HasSuffix(d.Msg.Text, "++"):
			update = func(k models.Karma) models.Karma { k.Upvotes += 1; return k }
		case strings.HasSuffix(d.Msg.Text, "--"):
			update = func(k models.Karma) models.Karma { k.Downvotes += 1; return k }
		case strings.HasSuffix(d.Msg.Text, "+-"), strings.HasSuffix(d.Msg.Text, "-+"):
			update = func(k models.Karma) models.Karma { k.Upvotes += 1; k.Downvotes += 1; return k }
		default:
			return nil
		}

		karmas := models.Karmas{}
		if err := store.Read(db.KarmasKey, &karmas); err != nil {
			return err
		}

		// strip '++', '--', etc. from key
		key := d.Msg.Text[:len(d.Msg.Text)-2]
		karmas[key] = update(karmas[key])
		return store.Write(db.KarmasKey, karmas)
	}
}
