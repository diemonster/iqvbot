package behaviors

import (
	"fmt"

	"github.com/nlopes/slack"
)

// NewMesageRTMEvent is a helper function that creates a slack.RTMEvent with the formatted message
func NewMessageRTMEvent(format string, tokens ...interface{}) slack.RTMEvent {
	return slack.RTMEvent{
		Type: "message",
		Data: &slack.MessageEvent{
			Msg: slack.Msg{
				Text: fmt.Sprintf(format, tokens...),
			},
		},
	}
}
