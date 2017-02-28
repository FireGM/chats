package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/FireGM/chats"
	"github.com/FireGM/chats/goodgame"
	"github.com/FireGM/chats/interfaces"
	"github.com/FireGM/chats/peka2tv"
	"github.com/FireGM/chats/twitch"
	"golang.org/x/net/websocket"
)

// func makeHandle(ch chan interfaces.Message) func(interfaces.Message, interfaces.Bot) {
// 	return func(message interfaces.Message, b interfaces.Bot) {
// 		ch <- message
// 	}
// }

func makeWSHandler(ch <-chan interfaces.Message) func(*websocket.Conn) {
	return func(ws *websocket.Conn) {
		for m := range ch {
			ws.Write([]byte(m.GetRenderFullHTML()))
		}
	}
}

func echoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func main() {
	ch := make(chan interfaces.Message)
	connectToChats(ch)
	http.Handle("/echo", websocket.Handler(makeWSHandler(ch)))
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func connectToChats(ch chan interfaces.Message) {
	botP := peka2tv.New(chats.MakerHandlers(ch))
	err := botP.Connect()
	if err != nil {
		panic(err)
	}
	err = botP.Join("all")
	if err != nil {
		panic(err)
	}
	botGG := goodgame.New(chats.MakerHandlers(ch))
	err = botGG.Connect()
	if err != nil {
		panic(err)
	}
	err = botGG.JoinBySlug("Miker")
	if err != nil {
		panic(err)
	}
	connectTwitch(ch)
}

type Conf struct {
	Token    string `json:"token"`
	ClientID string `json:"client_id"`
	Nickname string `json:"nickname"`
}

func connectTwitch(ch chan interfaces.Message) *twitch.Bot {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	conf := Conf{}
	err := decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
	botTwitch := twitch.NewWithRender(conf.Nickname, conf.Token,
		conf.ClientID, chats.MakerHandlers(ch))
	err = botTwitch.Connect()
	if err != nil {
		panic(err)
	}
	botTwitch.Join("lirik")
	return botTwitch
}
