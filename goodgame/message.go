package goodgame

import (
	"fmt"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

const (
	privMsg  = "PRIVMSG"
	clearMsg = "CLEARCHAT"
)

type Message struct {
	Channel        int              `json:"channel_id"`
	UserID         int              `json:"user_id"`
	Username       string           `json:"user_name"`
	UserRights     int              `json:"user_rights"`
	Paidsmiles     interface{}      `json:"paidsmiles"`
	Color          string           `json:"color"`
	Text           string           `json:"text"`
	Icon           string           `json:"icon"`
	Donat          int              `json:"donat"`
	Premium        interface{}      `json:"premium"`
	Premiums       []int            `json:"premiums"`
	Emotes         map[string]Smile `json:"emotes,omitempty"`
	TextWithEmotes template.HTML    `json:"text_with_emotes"`
	NicknameRender template.HTML    `json:"nickname_render"`
	FullRender     template.HTML    `json:"full_render"`
	Type           string           `json:"type"`
}

type MessageBan struct {
	Message
	UserID string `json:"user_id"`
}

func (m *MessageBan) GetUID() string {
	return m.UserID
}

func (m *Message) Init() {
	m.sortPremiums()
	m.parseEmotes()
}

func (m *Message) GetChatName() string {
	return "goodgame"
}

func (m *Message) GetChannelName() string {
	return strconv.Itoa(m.Channel)
}

func (m *Message) GetTextMessage() string {
	return m.Text
}
func (m *Message) ToUser(user string) bool {
	return strings.Contains(strings.ToLower(m.Text), strings.ToLower(user+","))
}

func (m *Message) GetUserFrom() string {
	return m.Username
}

func (m *Message) GetUID() string {
	return strconv.Itoa(m.UserID)
}

func (m *Message) IsFromUser() bool {
	return m.Username != ""
}

func (m *Message) GetRenderMessHTML() template.HTML {
	m.Init()
	if m.TextWithEmotes != "" {
		return m.TextWithEmotes
	}
	premium, ok := m.Premium.(bool)
	if !ok {
		premium = false
	}
	//neednt escape, gg send escaped string
	escaped := m.Text
	if len(m.Emotes) < 1 {
		m.TextWithEmotes = `<div class="message goodgame-message">` + template.HTML(escaped) + `</div>`
		return m.TextWithEmotes
	}
	for _, emote := range m.Emotes {
		url := emote.ImgBig
		if emote.Animated && premium {
			url = emote.ImgGif
		}
		escaped = strings.Replace(escaped, ":"+emote.Name+":",
			`<img class="smile gg-smile" src="`+url+`" alt="`+emote.Name+`">`,
			-1)
	}
	m.TextWithEmotes = `<div class="message goodgame-message">` + template.HTML(escaped) + `</div>`
	return m.TextWithEmotes
}

func (m *Message) GetRenderNicknameHTML() template.HTML {
	if m.NicknameRender != "" {
		return m.NicknameRender
	}
	icon := m.getIcon()
	nickname := fmt.Sprintf(`<p class="nickname goodgame-nickname">%s</p>`,
		m.GetUserFrom())
	m.NicknameRender = template.HTML(`<div class="nickname-badge goodgame-nickname-badge">` +
		icon + nickname + `</div>`)
	return m.NicknameRender
}

func (m *Message) getIcon() string {
	icon := ""
	if prem, ok := m.Premium.(bool); prem && ok {
		icon += `<span class="subscribe goodgame-subscribe"></span>`
	}
	if m.UserRights > 0 {
		icon += `<span class="moderator goodgame-moderator"></span>`
	}
	return icon
}

func (m *Message) GetRenderFullHTML() template.HTML {
	if m.FullRender != "" {
		return m.FullRender
	}
	m.FullRender = template.HTML(`<div class="full-message goodgame-full-message">` +
		m.GetRenderNicknameHTML() + `<span class="separator goodgame-separator"></span>` + m.GetRenderMessHTML() + `</div>`)
	return m.FullRender
}

// delete message
func (m *Message) IsClearMessage() bool {
	return m.Type == clearMsg
}

func (m *Message) parseEmotes() {
	m.Emotes = map[string]Smile{}
	for _, word := range strings.Fields(m.Text) {
		if !strings.HasPrefix(word, ":") && !strings.HasSuffix(word, ":") {
			continue
		}
		if len(word) < 3 {
			continue
		}
		smileStr := word[1 : len(word)-1]
		if s, ok := smiles[smileStr]; ok {
			if s.ChannelID == 0 || m.checkPremiumChan(s.ChannelID) ||
				(m.Donat == s.Donat && m.Donat != 0) {
				m.Emotes[word] = s
			}
		}
	}
}

func (m *Message) sortPremiums() {
	sort.Ints(m.Premiums)
}

func (m *Message) checkPremiumChan(id int) bool {
	if len(m.Premiums) < 1 {
		return false
	}
	i := sort.SearchInts(m.Premiums, id)
	if i == len(m.Premiums) || m.Premiums[i] != id {
		return false
	}
	return true
}

func (m *Message) GetColorNickname() string {
	return m.Color
}

func (m *Message) IsSubscriber() (bool, string) {
	if prem, ok := m.Premium.(bool); prem && ok {
		return true, ""
	}
	return false, ""
}
func (m *Message) IsModerator() (bool, string) {
	if m.UserRights > 0 {
		return true, ""
	}
	return false, ""
}
