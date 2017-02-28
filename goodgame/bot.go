package goodgame

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"log"

	"github.com/FireGM/chats/interfaces"
	"github.com/gorilla/websocket"
)

func New(handleFunc func(interfaces.Message, interfaces.Bot)) *Bot {
	return &Bot{handleFunc: handleFunc}
}

func DefaultBot() *Bot {
	return &Bot{handleFunc: handleFunc}
}

func handleFunc(m interfaces.Message, b interfaces.Bot) {
	t := m.GetTextMessage()
	fmt.Println(t)
}

type Bot struct {
	channels   map[string]time.Time
	handleFunc func(interfaces.Message, interfaces.Bot)
	locker     sync.RWMutex
	conn       *websocket.Conn
	userID     int
	username   string
}

func (b *Bot) Connect() error {
	wsClient, _, err := websocket.DefaultDialer.Dial("ws://chat.goodgame.ru:8081/chat/websocket", nil)
	if err != nil {
		return err
	}
	b.conn = wsClient
	once.Do(goUpdater)
	go b.reader()
	return nil
}

func (b *Bot) LoginByPass(login, password string) error {
	user := getUserByLoginPass(login, password)
	err := b.conn.WriteJSON(GGruct{Type: "auth", Data: AuthStructToken{UserID: user.ID, Token: user.Token}})
	return err
}

func (b *Bot) Join(ch string) error {
	err := b.conn.WriteJSON(GGruct{Type: "join", Data: map[string]string{"channel_id": ch}})
	return err
}

func (b *Bot) SendMessageToChan(ch string, message string) {
	err := b.conn.WriteJSON(GGruct{Type: "send_message", Data: MessageReq{ch, message}})
	if err != nil {
		panic(err)
	}
}

func (b *Bot) JoinBySlug(slug string) error {
	id, err := GetStreamInfo(slug)
	if err != nil {
		return err
	}
	return b.Join(strconv.Itoa(id))
}

func (b *Bot) reader() {
	for {
		var data json.RawMessage
		gg := GGruct{Data: &data}
		err := b.conn.ReadJSON(&gg)
		if err != nil {
			log.Println(err)
		}
		// fmt.Println(string(data))
		switch gg.Type {
		case "welcome":
			fmt.Println("connect to gg")
		case "success_auth":
			var g AuthStructToken
			err := json.Unmarshal(data, &g)
			if err != nil {
				log.Println(err)
			}
			fmt.Println(g)
		case "success_join":
			fmt.Println("Join")
		case "message":
			var message Message
			err := json.Unmarshal(data, &message)
			if err != nil {
				log.Println(err)
			}
			b.handleFunc(&message, b)
		}
	}
}
