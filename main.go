package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kballard/go-shellquote"
	"github.com/nlopes/slack"
	"github.com/quintilesims/iqvbot/bot"
	"github.com/quintilesims/iqvbot/db"
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
		cli.StringFlag{
			Name:   "tenor-key",
			Usage:  "authentication key for the Tenor API",
			EnvVar: "IB_TENOR_KEY",
		},
		cli.StringFlag{
			Name:   "aws-access-key",
			Usage:  "access key for aws api",
			EnvVar: "IB_AWS_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "aws-secret-key",
			Usage:  "secret key for aws api",
			EnvVar: "IB_AWS_SECRET_KEY",
		},
		cli.StringFlag{
			Name:   "aws-region",
			Usage:  "region for aws api",
			Value:  "us-west-2",
			EnvVar: "IB_AWS_REGION",
		},
		cli.StringFlag{
			Name:   "dynamodb-table",
			Usage:  "name of the dynamodb table",
			EnvVar: "IB_DYNAMODB_TABLE",
		},
	}

	iqvbot.Action = func(c *cli.Context) error {
		tenorKey := c.String("tenor-key")
		if tenorKey == "" {
			return fmt.Errorf("Tenor Key is not set! (envvar: IB_TENOR_KEY)")
		}

		accessKey := c.String("aws-access-key")
		if accessKey == "" {
			return fmt.Errorf("AWS Access Key is not set! (envvar: IB_AWS_ACCESS_KEY)")
		}

		secretKey := c.String("aws-secret-key")
		if secretKey == "" {
			return fmt.Errorf("AWS Secret Key is not set! (envvar: IB_AWS_SECRET_KEY)")
		}

		region := c.String("aws-region")
		if region == "" {
			return fmt.Errorf("AWS Region is not set! (envvar: IB_AWS_REGION)")
		}

		table := c.String("dynamodb-table")
		if table == "" {
			return fmt.Errorf("DynamoDB Table is not set! (envvar: IB_DYNAMODB_TABLE)")
		}

		config := &aws.Config{
			Credentials: credentials.NewStaticCredentials(accessKey, secretKey, ""),
			Region:      aws.String(region),
		}

		store := db.NewDynamoDBStore(session.New(config), table)
		if err := db.Init(store); err != nil {
			return err
		}

		aliasStore := db.NewKeyValueStoreAdapter(store, db.AliasesKey)
		kvsStore := db.NewKeyValueStoreAdapter(store, db.KVSKey)
		triviaStore := slackbot.InMemoryTriviaStore{}

		// todo: start cleanup runner
		// todo: start reminders runner

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

		behaviors := []slackbot.Behavior{
			slackbot.NewStandardizeTextBehavior(),
			slackbot.NewExpandPromptBehavior("!", "iqvbot "),
			slackbot.NewAliasBehavior(aliasStore),
			bot.NewKarmaBehavior(store),
		}

		// start the real-time-messaging api
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
				app.Name = "iqvbot"
				app.Usage = "making email obsolete one step at a time"
				app.UsageText = "command [flags...] arguments..."
				app.Version = Version
				app.Writer = slackbot.WriterFunc(func(b []byte) (n int, err error) {
					isDisplayingHelp = true
					return w.Write(b)
				})
				app.CommandNotFound = func(c *cli.Context, command string) {
					text := fmt.Sprintf("Command '%s' does not exist", command)
					w.WriteString(text)
				}
				// todo: CandidateCommand
				// todo: InterviewCommand
				app.Commands = []cli.Command{
					slackbot.NewAliasCommand(aliasStore, w, slackbot.WithBefore(func(c *cli.Context) error {
						aliasStore.Invalidate()
						return nil
					})),
					slackbot.NewDefineCommand(slackbot.DatamuseAPIEndpoint, w),
					slackbot.NewDeleteCommand(client, info.User.ID, data.Channel),
					slackbot.NewEchoCommand(w),
					slackbot.NewGIFCommand(slackbot.TenorAPIEndpoint, tenorKey, w),
					bot.NewKarmaCommand(store, w),
					slackbot.NewKVSCommand(kvsStore, w),
					slackbot.NewRepeatCommand(client, data.Channel, rtm.IncomingEvents, func(m slack.Message) bool {
						return strings.HasPrefix(m.Text, "iqvbot ") && !strings.HasPrefix(m.Text, "iqvbot repeat")
					}),
					slackbot.NewTriviaCommand(triviaStore, slackbot.OpenTDBAPIEndpoint, data.Channel, w),
				}

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
