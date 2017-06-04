package twitch

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"net/textproto"
	"strings"
	"sync"
	"time"

	"github.com/FireGM/chats/interfaces"
)

const server = "irc.chat.twitch.tv"
const serverPort = "6667"

//for self u can get oauth token from
//http://twitchapps.com/tmi/
//recommend by twitch team
//https://github.com/justintv/Twitch-API/blob/master/IRC.md#connecting
func New(name, oauth string, handle func(interfaces.Message, interfaces.Bot)) *Bot {
	return &Bot{name: name, oauth: oauth, handleFunc: handle}
}

func NewWithRender(name, oauth string, clientId string, handle func(interfaces.Message, interfaces.Bot)) *Bot {
	clientID = clientId
	return &Bot{name: name, oauth: oauth, handleFunc: handle}
}

func defaultHandle(m interfaces.Message, b interfaces.Bot) {
	if !m.IsFromUser() {
		return
	}
	h := m.GetRenderMessHTML()
	log.Println(h)
}

type Bot struct {
	channels   map[string]*time.Time
	name       string
	oauth      string
	handleFunc func(interfaces.Message, interfaces.Bot)
	conn       net.Conn
	locker     sync.RWMutex
	disconnect bool
}

func (b *Bot) Connect() error {
	var err error
	b.conn, err = net.Dial("tcp", server+":"+serverPort)
	if err != nil {
		return err
	}
	go b.read()
	b.channels = map[string]*time.Time{}
	b.login()
	once.Do(goUpdater)
	return nil
}

func (b *Bot) Disconnect() error {

	err := b.conn.Close()
	if err != nil {
		return err
	}
	b.disconnect = true
	return nil
}

func (b *Bot) reconnect() error {
	var err error
	b.conn, err = net.Dial("tcp", server+":"+serverPort)
	if err != nil {
		return err
	}
	go b.read()
	b.login()
	for ch := range b.channels {
		b.Join(ch)
	}
	return nil
}

func (b *Bot) Close() error {
	return b.conn.Close()
}

func (b *Bot) Send(message string) error {
	_, err := fmt.Fprintf(b.conn, message+"\r\n")
	return err
}

func (b *Bot) Ban(channel, nickname string) error {
	return b.SendMessageToChan(channel, fmt.Sprintf(".ban %s", nickname))
}

func (b *Bot) Timeout(channel, nickname string, t int) error {
	return b.SendMessageToChan(channel, fmt.Sprintf(".timeout %s %d", nickname, t))
}

func (b *Bot) SendMessageToChan(ch, message string) error {
	return b.Send(fmt.Sprintf("PRIVMSG #%s :%s", ch, message))
}

func (b *Bot) Join(ch string) error {
	b.locker.Lock()
	defer b.locker.Unlock()
	if _, ok := b.channels[ch]; ok {
		return nil
	}
	err := b.Send("JOIN #" + ch)
	if err != nil {
		return err
	}
	n := time.Now()
	b.channels[ch] = &n
	return nil
}

func (b *Bot) Leave(ch string) error {
	b.locker.Lock()
	defer b.locker.Unlock()
	if _, ok := b.channels[ch]; ok {
		return nil
	}
	err := b.Send("PART #" + ch)
	if err != nil {
		return err
	}
	delete(b.channels, ch)
	return nil
}

func (b *Bot) login() {
	b.Send("PASS oauth:" + b.oauth)
	b.Send("NICK " + b.name)
	b.Send("CAP REQ twitch.tv/tags")
	b.Send("CAP REQ twitch.tv/commands")
}

func (b *Bot) read() {
	reader := textproto.NewReader(bufio.NewReader(b.conn))
	for {
		line, err := reader.ReadLine()
		if err != nil {
			log.Println(err)
			if !b.disconnect {
				b.reconnect()
			}
			return
		}
		// log.Println(line)
		if strings.HasPrefix(line, "PING") {
			b.Send(strings.Replace(line, "PING", "PONG", 1))
			continue
		}
		m, err := ParseMessage(line)
		if err != nil {
			continue
		}
		go b.handleFunc(&m, b)
	}
}
