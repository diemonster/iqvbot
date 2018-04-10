package bot

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/nlopes/slack"
	"github.com/quintilesims/slackbot/db"
	"github.com/quintilesims/slackbot/models"
	glob "github.com/ryanuber/go-glob"
	"github.com/urfave/cli"
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

// NewKarmaCommand returns a cli.Command that displays karma
func NewKarmaCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:      "karma",
		Usage:     "display karma for entries that match the given GLOB pattern",
		ArgsUsage: "GLOB",
		Flags: []cli.Flag{
			cli.IntFlag{
				Name:  "count",
				Value: 10,
				Usage: "The maximum number of entries to display",
			},
			cli.BoolFlag{
				Name:  "ascending",
				Usage: "Show results in ascending order",
			},
		},
		Action: func(c *cli.Context) error {
			g := c.Args().Get(0)
			if g == "" {
				return slackbot.NewUserInputError("Argument GLOB is required")
			}

			karmas := models.Karmas{}
			if err := store.Read(db.KarmasKey, &karmas); err != nil {
				return err
			}

			results := models.Karmas{}
			for k, v := range karmas {
				if glob.Glob(g, k) {
					results[k] = v
				}
			}

			keys := results.SortKeys(c.Bool("ascending"))
			if len(keys) == 0 {
				return slackbot.NewUserInputErrorf("Could not find any karma entries matching *%s*", g)
			}

			var text string
			for i := 0; i < len(keys) && i < c.Int("count"); i++ {
				karma := results[keys[i]]
				text += fmt.Sprintf("*%s*: %d (%d upvotes, %d downvotes)\n",
					keys[i],
					karma.Upvotes-karma.Downvotes,
					karma.Upvotes,
					karma.Downvotes)
			}

			return slackbot.WriteString(w, text)
		},
	}
}
