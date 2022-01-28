package slacker

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
)

type Slacker struct {
	api     *slack.Client
	onError func(error)
}

func New(token string, onError func(error)) (s *Slacker, err error) {
	s = &Slacker{
		api:     slack.New(token, slack.OptionDebug(true)),
		onError: onError,
	}

	_, err = s.api.AuthTest()
	if err != nil {
		return
	}

	return
}

func (s *Slacker) MakeSendMessage(chanID string) func(string, ...interface{}) {
	return func(format string, v ...interface{}) {
		_, _, err := s.api.PostMessage(chanID, slack.MsgOptionText(fmt.Sprintf(format, v...), false))
		if err != nil {
			s.onError(err)
		}
	}
}

func (s *Slacker) SetReader(
	ctx context.Context, chanID string, lastTimestamp string, onSetLastTimestamp, onMessage func(string),
) {
	go func() {
		if e := ctx.Err(); e != nil {
			s.onError(e)
			return
		}

		last := lastTimestamp
		for {
			res, err := s.api.GetConversationHistory(
				&slack.GetConversationHistoryParameters{
					ChannelID: chanID,
					Cursor:    "",
					Inclusive: false,
					Latest:    "",
					Limit:     0,
					Oldest:    last,
				},
			)
			if err != nil {
				s.onError(err)
			}

			if len(res.Messages) > 0 {
				last = res.Messages[0].Timestamp
				onSetLastTimestamp(last)
			}

			for i := range res.Messages {
				onMessage(res.Messages[i].Text)
			}

			time.Sleep(time.Second * 5)
		}
	}()
}
