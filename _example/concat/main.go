package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/FireGM/chats"
	"github.com/FireGM/chats/goodgame"
	"github.com/FireGM/chats/interfaces"
	"github.com/FireGM/chats/peka2tv"
	"github.com/FireGM/chats/twitch"
	"github.com/FireGM/chats/youtube"

	"github.com/gorilla/websocket"
)

// func makeHandle(ch chan interfaces.Message) func(interfaces.Message, interfaces.Bot) {
// 	return func(message interfaces.Message, b interfaces.Bot) {
// 		ch <- message
// 	}
// }

type Conf struct {
	TwitchToken    string `json:"twitch_token"`
	TwitchClientID string `json:"twitch_client_id"`
	TwitchNickname string `json:"twitch_nickname"`
	YoutubeApiKey  string `json:"youtube_api_key"`
}

var upgrader = websocket.Upgrader{}

func makeWSHandler(ch <-chan interfaces.Message) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		for m := range ch {
			c.WriteMessage(websocket.TextMessage, []byte(m.GetRenderFullHTML()))
		}
	}
}

func main() {
	ch := make(chan interfaces.Message)
	conf := getConf()
	connectToChats(ch, conf)
	http.HandleFunc("/echo", makeWSHandler(ch))
	http.Handle("/", http.FileServer(http.Dir(".")))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func getConf() Conf {
	file, _ := os.Open("conf.json")
	decoder := json.NewDecoder(file)
	conf := Conf{}
	err := decoder.Decode(&conf)
	if err != nil {
		panic(err)
	}
	return conf
}

func connectToChats(ch chan interfaces.Message, conf Conf) {
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
	botTwitch := connectTwitch(ch, conf)
	botTwitch.Join("lirik")
	botYoutube := connectYoutube(ch, conf)
	botYoutube.Join("UC3wf1pcRkwV-NvUpBfHovPw") //24x7 news stream
}

func connectTwitch(ch chan interfaces.Message, conf Conf) *twitch.Bot {
	botTwitch := twitch.NewWithRender(conf.TwitchNickname, conf.TwitchToken,
		conf.TwitchClientID, chats.MakerHandlers(ch))
	err := botTwitch.Connect()
	if err != nil {
		panic(err)
	}
	return botTwitch
}

func connectYoutube(ch chan interfaces.Message, conf Conf) *youtube.Bot {
	bot := youtube.New(chats.MakerHandlers(ch), conf.YoutubeApiKey)
	return bot
}
