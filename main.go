package chats

import (
	"github.com/FireGM/chats/goodgame"
	"github.com/FireGM/chats/interfaces"
	"github.com/FireGM/chats/peka2tv"
	"github.com/FireGM/chats/twitch"
)

func MakerHandlers(ch chan interfaces.Message) func(interfaces.Message, interfaces.Bot) {
	return func(message interfaces.Message, b interfaces.Bot) {
		ch <- message
	}
}

func GetPeka2tv(token string, handler func(interfaces.Message, interfaces.Bot)) *peka2tv.Bot {
	bot := peka2tv.New(handler)
	bot.Connect()
	bot.LoginByToken(token)
	return bot
}

func GetGoodGame(login, pass string, handler func(interfaces.Message, interfaces.Bot)) *goodgame.Bot {
	bot := goodgame.New(handler)
	bot.Connect()
	bot.LoginByPass(login, pass)
	return bot
}

func GetTwitchChat(nickname, token, clientID string, handler func(interfaces.Message, interfaces.Bot)) *twitch.Bot {
	bot := twitch.NewWithRender(nickname, token,
		clientID, handler)
	bot.Connect()
	return bot
}
