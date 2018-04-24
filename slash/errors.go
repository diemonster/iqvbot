package slash

import (
	"fmt"

	"github.com/nlopes/slack"
)

type SlackMessageError struct {
	*slack.Msg
}

func NewSlackMessageError(text string) *SlackMessageError {
	return &SlackMessageError{
		&slack.Msg{
			ResponseType: "ephemeral",
			Text:         text,
		},
	}
}

func NewSlackMessageErrorf(format string, tokens ...interface{}) *SlackMessageError {
	return NewSlackMessageError(fmt.Sprintf(format, tokens...))
}

func (s *SlackMessageError) Error() string {
	return s.Text
}
