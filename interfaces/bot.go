package interfaces

import "html/template"

type Bot interface {
	// (channel)
	Join(string) error
	// (channel)
	Leave(string) error
	// (channel, message)
	SendMessageToChan(string, string) error
	// ban(channel, nickname)
	Ban(string, string) error
	//with time in seconds
	Timeout(string, string, int) error
}

type Message interface {
	GetRenderMessHTML() template.HTML
	GetRenderNicknameHTML() template.HTML
	GetRenderFullHTML() template.HTML
	GetChatName() string
	GetTextMessage() string
	ToUser(string) bool
	GetUserFrom() string
	IsFromUser() bool
	GetChannelName() string
	GetColorNickname() string
	IsClearMessage() bool
}
