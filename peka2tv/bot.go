package peka2tv

import (
	"fmt"
	"sync"
	"time"

	"log"

	"github.com/FireGM/chats/interfaces"
	"github.com/graarh/golang-socketio"
	"github.com/graarh/golang-socketio/transport"
)

const (
	codeRequestURL  = "http://peka2.tv/api/oauth/request"
	tokenRequestURL = "http://peka2.tv/api/oauth/exchange"
)

const chatURL = "chat.peka2.tv"

func New(handleFunc func(interfaces.Message, interfaces.Bot)) *Bot {
	return &Bot{handleFunc: handleFunc}
}

func Default() *Bot {
	return &Bot{handleFunc: defaultHandleFunc}
}

func defaultHandleFunc(m interfaces.Message, b interfaces.Bot) {
	h := m.GetRenderMessHTML()
	fmt.Println(h)
}

type Bot struct {
	channels   map[string]time.Time
	handleFunc func(interfaces.Message, interfaces.Bot)
	locker     sync.RWMutex
	conn       *gosocketio.Client
	username   string
	userID     int
	token      string
}

func (b *Bot) Connect() error {
	conn, err := gosocketio.Dial(
		gosocketio.GetUrl(chatURL, 80, false),
		transport.GetDefaultWebsocketTransport(),
	)
	if err != nil {
		return err
	}
	once.Do(goUpdater)
	conn.On("/chat/message", func(h *gosocketio.Channel, m Message) {
		b.handleFunc(&m, b)
	})
	b.conn = conn
	return nil
}

//https://github.com/funstream-api/api/blob/master/oauth.md
func (b *Bot) LoginByToken(token string) string {
	user, err := getCurrentUser(token)
	if err != nil {
		panic(err)
	}
	b.username = user.Name
	b.userID = user.ID
	res, err := b.conn.Ack("/chat/login", struct {
		Token string `json:"token"`
	}{Token: token}, time.Second*10)
	if err != nil {
		panic(err)
	}
	return res
}

//need login before send messages
func (b *Bot) SendMessageToChan(channel, message string) {
	res, err := b.conn.Ack("/chat/publish", struct {
		Channel string `json:"channel"`
		Text    string `json:"text"`
		From    User   `json:"from"`
	}{Channel: channel, Text: message, From: User{b.userID, b.username}}, time.Second*10)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println(res)
}

func (b *Bot) Join(ch string) error {
	_, err := b.conn.Ack("/chat/join", struct {
		Channel string `json:"channel"`
	}{Channel: ch}, time.Second*10)
	if err != nil {
		panic(err)
	}
	return nil
}

func (b *Bot) JoinBySlug(slug string) error {
	id, err := GetUserIdBySlug(slug)
	if err != nil {
		return err
	}
	stream := fmt.Sprintf("stream/%d", id)
	return b.Join(stream)
}

func (b *Bot) Send(message string) error {
	return nil
}
