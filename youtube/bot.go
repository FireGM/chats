package youtube

import (
	"errors"
	"sync"
	"time"

	"log"

	"github.com/FireGM/chats/interfaces"
)

func New(handleFunc func(interfaces.Message, interfaces.Bot), apiKey string) *Bot {
	return &Bot{handleFunc: handleFunc, apiKey: apiKey, streams: map[string]YouChannel{}}
}

func NewWithAuth(handleFunc func(interfaces.Message, interfaces.Bot), apiKey, oAuth string) *Bot {
	return &Bot{handleFunc: handleFunc, apiKey: apiKey, oAuth: oAuth}
}

type YouChannel struct { //todo: oAuth in channel?
	ChannelID   string
	ChatID      string
	LastMessage time.Time
}

func (y *YouChannel) reader(handler func(interfaces.Message, interfaces.Bot), apiKey string, bot *Bot) {
	errorCounter := 0
	for {
		messages, err := getMessages(y.ChatID, apiKey)
		if err != nil {
			errorCounter++
			log.Println(err)
			if errorCounter < 10 {
				continue
			}
		}
		errorCounter = 0
		sleeper := getSleepTime(messages.PollingIntervalMillis)
		if messages.PageInfo.TotalResults < 1 || len(messages.Items) < 1 {
			time.Sleep(sleeper)
			continue
		}
		newLast := time.Time{}
		for _, message := range messages.Items {
			mes, err := parseMessage(message, y.ChannelID)
			if err != nil {
				log.Println(err)
				continue
			}
			if mes.SendTime.After(y.LastMessage) {
				handler(&mes, bot)
				if mes.SendTime.After(newLast) {
					newLast = mes.SendTime
				}
			}
		}
		y.LastMessage = newLast
		time.Sleep(sleeper)
	}
}

type Bot struct {
	streams    map[string]YouChannel
	handleFunc func(interfaces.Message, interfaces.Bot)
	apiKey     string
	oAuth      string
	sync.RWMutex
}

func (b *Bot) Join(channelID string) error {
	b.Lock()
	chatID, err := GetChatIDByChannel(channelID, b.apiKey)
	if err != nil {
		return err
	}
	ych := YouChannel{ChannelID: channelID, ChatID: chatID, LastMessage: time.Now()}
	go ych.reader(b.handleFunc, b.apiKey, b)
	b.streams[channelID] = ych
	return nil
}

func (b *Bot) SendMessageToChan(channel, message string) error {
	if b.oAuth == "" {
		return errors.New("Access token empty")
	}
	uChannel, ok := b.streams[channel]
	if !ok {
		return errors.New("Need join to channel")
	}
	err := sendMessageToChat(uChannel.ChatID, message, b.oAuth, b.apiKey)
	if err != nil {
		return err
	}
	return nil
}
