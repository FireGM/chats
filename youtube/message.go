package youtube

import "html/template"
import "strings"
import "fmt"
import "time"

type Message struct {
	ChannelID      string        `json:"channel_id`
	Owner          string        `json:"owner"`
	ChatOwner      bool          `json:"chat_owner"`
	Moderator      bool          `json:"moderator"`
	Text           string        `json:"text"`
	SendTime       time.Time     `json:"-"`
	TextWithEmotes template.HTML `json:"text_with_emotes"`
	NicknameRender template.HTML `json:"nickname_render"`
	FullRender     template.HTML `json:"full_render"`
}

func (m *Message) GetChatName() string {
	return "youtube"
}

func (m *Message) GetTextMessage() string {
	return m.Text
}

func (m *Message) ToUser(user string) bool {
	return strings.Contains(m.Text, user)
}

func (m *Message) GetUserFrom() string {
	return m.Owner
}

func (m *Message) IsFromUser() bool {
	if m.Owner == "" {
		return false
	}
	return true
}

func (m *Message) GetChannelName() string {
	return m.ChannelID
}

func (m *Message) GetColorNickname() string {
	return "#000"
}

func (m *Message) GetRenderMessHTML() template.HTML {
	if m.TextWithEmotes != "" {
		return m.TextWithEmotes
	}
	m.TextWithEmotes = template.HTML(fmt.Sprintf(`<div class="message youtube-message">%s</div>`, m.Text)) //todo: render emoji?
	return m.TextWithEmotes
}

func (m *Message) GetRenderNicknameHTML() template.HTML {
	if m.NicknameRender != "" {
		return m.NicknameRender
	}
	r := ""
	if m.ChatOwner {
		r += `<div class="badge youtube-badge"><span class="chat-owner"></span></div>`
	} else if m.Moderator {
		r += `<div class="badge youtube-badge"><span class="chat-moderator"></span></div>`
	}
	r += fmt.Sprintf(`<p class="nickname youtube-nickname">%s</p>`, m.Owner)
	m.NicknameRender = template.HTML(fmt.Sprintf(`<div class="nickname-badge youtube-nickname-badge">%s</div>`, r))
	return m.NicknameRender
}

func (m *Message) GetRenderFullHTML() template.HTML {
	if m.FullRender != "" {
		return m.FullRender
	}
	m.FullRender = template.HTML(fmt.Sprintf(`<div class="full-message youtube-full-message">%s<span class="separator youtube-separator"></span>%s</div>`,
		m.GetRenderNicknameHTML(), m.GetRenderMessHTML()))
	return m.FullRender
}

func parseMessage(mesR MessageResp, channelID string) (Message, error) {
	var m Message
	m.ChannelID = channelID
	m.Owner = mesR.AuthorDetails.DisplayName
	m.ChatOwner = mesR.AuthorDetails.IsChatOwner
	m.Moderator = mesR.AuthorDetails.IsChatModerator
	m.Text = mesR.Snippet.DisplayMessage
	m.SendTime = mesR.Snippet.PublishedAt
	return m, nil //todo: errors?
}
