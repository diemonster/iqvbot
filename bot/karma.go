package bot

import (
	"fmt"
	"io"
	"strings"

	"github.com/nlopes/slack"

	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/iqvbot/models"
	glob "github.com/ryanuber/go-glob"
	"github.com/urfave/cli"
	"github.com/zpatrick/slackbot"
)

// NewKarmaBehavior returns a behavior that updates karma in the provided store.
// Karma updates are triggered by the presence of '++', '--', '+-', or '-+' at the end of a message.
func NewKarmaBehavior(store db.Store) slackbot.Behavior {
	return func(e slack.RTMEvent) error {
		d, ok := e.Data.(*slack.MessageEvent)
		if !ok {
			return nil
		}

		var update func(models.KarmaEntry) models.KarmaEntry
		switch {
		case strings.HasSuffix(d.Msg.Text, "++"):
			update = func(e models.KarmaEntry) models.KarmaEntry { e.Upvotes += 1; return e }
		case strings.HasSuffix(d.Msg.Text, "--"):
			update = func(e models.KarmaEntry) models.KarmaEntry { e.Downvotes += 1; return e }
		case strings.HasSuffix(d.Msg.Text, "+-"), strings.HasSuffix(d.Msg.Text, "-+"):
			update = func(e models.KarmaEntry) models.KarmaEntry { e.Upvotes += 1; e.Downvotes += 1; return e }
		default:
			return nil
		}

		karma := models.Karma{}
		if err := store.Read(db.KarmaKey, &karma); err != nil {
			return err
		}

		// strip '++', '--', etc. from key
		key := d.Msg.Text[:len(d.Msg.Text)-2]
		karma[key] = update(karma[key])
		return store.Write(db.KarmaKey, karma)
	}
}

// NewKarmaCommand returns a cli.Command that displays karma
func NewKarmaCommand(store db.Store, w io.Writer) cli.Command {
	return cli.Command{
		Name:      "karma",
		Usage:     "display karma entries that match the given GLOB pattern",
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

			karma := models.Karma{}
			if err := store.Read(db.KarmaKey, &karma); err != nil {
				return err
			}

			matches := models.Karma{}
			for k, v := range karma {
				if glob.Glob(g, k) {
					matches[k] = v
				}
			}

			keys := matches.SortKeys(c.Bool("ascending"))
			if len(keys) == 0 {
				return slackbot.NewUserInputErrorf("Could not find any karma entries matching *%s*", g)
			}

			var text string
			for i := 0; i < len(keys) && i < c.Int("count"); i++ {
				entry := matches[keys[i]]
				text += fmt.Sprintf("*%s*: %d (%d upvotes, %d downvotes)\n",
					keys[i],
					entry.Upvotes-entry.Downvotes,
					entry.Upvotes,
					entry.Downvotes)
			}

			return slackbot.WriteString(w, text)
		},
	}
}
