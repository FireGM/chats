package interfaces

import "html/template"

type Bot interface {
	Join(string) error
	SendMessageToChan(string, string)
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
}
