package youtube

import (
	"errors"
	"sync"
	"time"

	"log"

	"github.com/FireGM/chats/interfaces"
)

func New(handleFunc func(interfaces.Message, interfaces.Bot), apiKey string) *Bot {
	return &Bot{handleFunc: handleFunc, apiKey: apiKey, streams: map[string]*YouChannel{}}
}

func NewWithAuth(handleFunc func(interfaces.Message, interfaces.Bot), apiKey, oAuth string) *Bot {
	return &Bot{handleFunc: handleFunc, apiKey: apiKey, oAuth: oAuth}
}

type YouChannel struct { //todo: oAuth in channel?
	ChannelID   string
	ChatID      string
	LastMessage time.Time
	stop        bool
	sync.RWMutex
}

func (y *YouChannel) reader(handler func(interfaces.Message, interfaces.Bot), apiKey string, bot *Bot) {
	errorCounter := 0
	for {
		y.RLock()
		if y.stop {
			y.RUnlock()
			break
		}
		messages, err := getMessages(y.ChatID, apiKey)
		if err != nil {
			errorCounter++
			log.Println(err)
			if errorCounter < 10 {
				y.RUnlock()
				continue
			}
		}
		errorCounter = 0
		sleeper := getSleepTime(messages.PollingIntervalMillis)
		if messages.PageInfo.TotalResults < 1 || len(messages.Items) < 1 {
			y.RUnlock()
			time.Sleep(sleeper)
			continue
		}
		newLast := y.LastMessage
		for _, message := range messages.Items {
			mes, err := parseMessage(message, y.ChannelID)
			if err != nil {
				log.Println(err)
				y.RUnlock()
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
		y.RUnlock()
		time.Sleep(sleeper)
	}
}

func (y *YouChannel) Stop() {
	y.stop = true
}

type Bot struct {
	streams    map[string]*YouChannel
	handleFunc func(interfaces.Message, interfaces.Bot)
	apiKey     string
	oAuth      string
	sync.RWMutex
}

func (b *Bot) Join(channelID string) error {
	b.Lock()
	defer b.Unlock()
	chatID, err := GetChatIDByChannel(channelID, b.apiKey)
	if err != nil {
		return err
	}
	ych := YouChannel{ChannelID: channelID, ChatID: chatID, LastMessage: time.Now()}
	if _, ok := b.streams[channelID]; ok {
		b.Leave(channelID)
	}
	go ych.reader(b.handleFunc, b.apiKey, b)
	b.streams[channelID] = &ych
	return nil
}

func (b *Bot) Leave(ch string) error {
	b.RLock()
	defer b.RUnlock()
	uChannel, ok := b.streams[ch]
	if !ok {
		return errors.New("Channel non found")
	}
	uChannel.Stop()
	delete(b.streams, ch)
	return nil
}

func (b *Bot) SendMessageToChan(channel, message string) error {
	if b.oAuth == "" {
		return errors.New("Access token empty")
	}
	b.RLock()
	defer b.RUnlock()
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
