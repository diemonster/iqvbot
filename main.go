package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/kballard/go-shellquote"
	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/db"
	"github.com/quintilesims/slackbot/utils"
	"github.com/zpatrick/slackbot"

	"github.com/urfave/cli"
)

// Version of the application
var Version string

func main() {
	if Version == "" {
		Version = "unset/develop"
	}

	iqvbot := cli.NewApp()
	iqvbot.Name = "iqvbot"
	iqvbot.Version = Version
	iqvbot.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:   "d, debug",
			Usage:  "enable debug logging",
			EnvVar: "IB_DEBUG",
		},
		cli.StringFlag{
			Name:   "slack-app-token",
			Usage:  "authentication token for the slack application",
			EnvVar: "IB_SLACK_APP_TOKEN",
		},
		cli.StringFlag{
			Name:   "slack-bot-token",
			Usage:  "authentication token for the slack bot",
			EnvVar: "IB_SLACK_BOT_TOKEN",
		},
	}

	iqvbot.Action = func(c *cli.Context) error {
		// todo: create dynamodb data store
		store := db.NewStoreAdapter(db.NewMemoryStore())
		if err := db.Init(store); err != nil {
			return err
		}

		// todo: start runners

		// create the slack client
		appToken := c.String("slack-app-token")
		if appToken == "" {
			return fmt.Errorf("App Token is not set!")
		}

		botToken := c.String("slack-bot-token")
		if botToken == "" {
			return fmt.Errorf("Bot Token is not set!")
		}

		client := slackbot.NewDualSlackClient(appToken, botToken)

		// todo: create behaviors
		behaviors := []slackbot.Behavior{
			slackbot.NewExpandPromptBehavior("!", "iqvbot "),
			slackbot.NewAliasBehavior(store),
		}

		// start the real-time-messaging
		rtm := client.NewRTM()
		go rtm.ManageConnection()
		defer rtm.Disconnect()

		for e := range rtm.IncomingEvents {
			ctx := context.Background()
			info := rtm.GetInfo()

			for _, behavior := range behaviors {
				if err := behavior(ctx, e); err != nil {
					log.Printf("[ERROR] %s", err.Error())
				}
			}

			switch data := e.Data.(type) {
			case *slack.ConnectedEvent:
				log.Printf("[INFO] Slack connection successful!")
			case *slack.InvalidAuthEvent:
				return fmt.Errorf("The bot's auth token is invalid")
			case *slack.MessageEvent:
				text := data.Msg.Text
				if !strings.HasPrefix(text, "iqvbot ") {
					continue
				}

				args, err := shellquote.Split(text)
				if err != nil {
					m := rtm.NewOutgoingMessage(err.Error(), data.Channel)
					rtm.SendMessage(m)
					continue
				}

				var isDisplayingHelp bool
				w := bytes.NewBuffer(nil)

				app := cli.NewApp()
				app.Name = "slackbot"
				app.Usage = "making email obsolete one step at a time"
				app.UsageText = "command [flags...] arguments..."
				app.Version = Version
				app.Writer = utils.WriterFunc(func(b []byte) (n int, err error) {
					isDisplayingHelp = true
					return w.Write(b)
				})
				app.CommandNotFound = func(c *cli.Context, command string) {
					text := fmt.Sprintf("Command '%s' does not exist", command)
					w.WriteString(text)
				}
				app.Commands = []cli.Command{
					slackbot.NewAliasCommand(store, w, store.InvalidateBefore(db.AliasesKey)),
					slackbot.NewEchoCommand(w),
					slackbot.NewDeleteCommand(client, info.User.ID, data.Channel),
				}

				// idea for alias behavior: we use a command option to do .Before(), we invalidate the alisa store/cache?

				if err := app.Run(args); err != nil {
					w.WriteString(err.Error())
				}

				response := w.String()
				if isDisplayingHelp {
					response = fmt.Sprintf("```%s```", response)
				}

				m := rtm.NewOutgoingMessage(response, data.Channel)
				rtm.SendMessage(m)
			}
		}

		return nil
	}

	if err := iqvbot.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
