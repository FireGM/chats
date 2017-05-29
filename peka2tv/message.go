package peka2tv

import (
	"fmt"
	"html"
	"html/template"
	"sort"
	"strconv"
	"strings"
)

const (
	privMsg  = "PRIVMSG"
	clearMsg = "CLEARCHAT"
)

type ExtraData struct {
	Id int `json:"id"`
}

type Store struct {
	Icon    int
	Bonuses []int
	Subs    []int
}

//todo:
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Emote struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"`
	Url  string `json:"url"`
}

type Ban struct {
	ID      int    `json:"id"`
	Channel int    `json:"channel"`
	Reason  string `json:"reason"`
	Data    map[string]int
}

type Message struct {
	ID             int              `json:"id"`
	Channel        string           `json:"channel"`
	From           User             `json:"from"`
	Text           string           `json:"text"`
	Emotes         map[string]Emote `json:"emotes,omitempty"`
	To             User             `json:"to,omitempty"`
	Type           string           `json:"type"`
	Store          Store            `json:"store"`
	TextWithEmotes template.HTML    `json:"text_with_emotes"`
	NicknameRender template.HTML    `json:"nickname_render"`
	FullRender     template.HTML    `json:"full_render"`
	//id message for bans
}

func (m *Message) Init() {
	m.sortPerms()
	m.parseEmotes()
}

func (m *Message) GetChatName() string {
	return "peka2tv"
}

func (m *Message) GetChannelName() string {
	return m.Channel
}

func (m *Message) GetUID() string {
	return strconv.Itoa(m.ID)
}

func (m *Message) GetRenderMessHTML() template.HTML {
	if m.TextWithEmotes != "" {
		return m.TextWithEmotes
	}
	m.Init()
	escaped := html.EscapeString(m.Text)
	if len(m.Emotes) < 1 {
		m.TextWithEmotes = `<div class="message peka-message">` + template.HTML(escaped) + `</div>`
		return m.TextWithEmotes
	}
	maxSmiles := m.checkMaxSmiles()
	for _, emote := range m.Emotes {
		if maxSmiles == 0 {
			break
		}
		escaped = strings.Replace(escaped, emote.Code,
			`<img class="smile peka-smile" src="`+emote.Url+`" alt="`+emote.Name+`">`,
			1)
		maxSmiles--
	}
	m.TextWithEmotes = `<div class="message peka-message">` + template.HTML(escaped) + `</div>`
	return m.TextWithEmotes
}

func (m *Message) GetRenderNicknameHTML() template.HTML {
	if m.NicknameRender != "" {
		return m.NicknameRender
	}
	nickname := fmt.Sprintf(`<p class="nickname peka-nickname">%s</p>`,
		html.EscapeString(m.GetUserFrom()))
	badge := ""
	if m.Store.Icon != 0 {
		badge = `<img class="badge peka-badge" src="` + icons[m.Store.Icon].URL + `">`
	}
	m.NicknameRender = template.HTML(`<div class="nickname-badge peka-nickname-badge">` +
		badge + nickname + `</div>`)
	return m.NicknameRender
}

func (m *Message) GetRenderFullHTML() template.HTML {
	if m.FullRender != "" {
		return m.FullRender
	}
	m.FullRender = `<div class="full-message peka-full-message">` + m.GetRenderNicknameHTML() +
		`<span class="separator peka-separator"></span>` + m.GetRenderMessHTML() + `</div>`
	return m.FullRender
}

func (m *Message) GetColorNickname() string {
	c := "#000"
	for _, id := range m.Store.Bonuses {
		if color, ok := nickColors[id]; ok {
			c = color
		}
	}
	return c
}

func (m *Message) ToUser(user string) bool {
	return strings.ToLower(m.To.Name) == strings.ToLower(user)
}

func (m *Message) GetUserFrom() string {
	return m.From.Name
}

func (m *Message) IsFromUser() bool {
	return m.From.ID != 0
}

func (m *Message) GetTextMessage() string {
	return m.Text
}

// delete message
func (m *Message) IsClearMessage() bool {
	return m.Type == clearMsg
}

func (m *Message) checkMaxSmiles() int {
	max := 2
	for perm, mm := range smilesPerMessage {
		if m.checkPerm(perm) && mm > max {
			max = mm
		}
	}
	return max
}

func (m *Message) sortPerms() {
	sort.Ints(m.Store.Bonuses)
}

func (m *Message) parseEmotes() {
	m.Emotes = map[string]Emote{}
	for _, word := range strings.Fields(m.Text) {
		if !strings.HasPrefix(word, ":") && !strings.HasSuffix(word, ":") {
			continue
		}
		if len(word) < 3 {
			continue
		}
		smileStr := word[1 : len(word)-1]
		if s, ok := smiles[smileStr]; ok {
			if s.BonusId == 0 || m.checkPerm(s.BonusId) {
				m.Emotes[word] = Emote{Name: smileStr, Url: s.Url, Code: word}
			}
		}
	}
}

func (m *Message) checkPerm(id int) bool {
	i := sort.SearchInts(m.Store.Bonuses, id)
	if i == len(m.Store.Bonuses) || m.Store.Bonuses[i] != id {
		return false
	}
	return true
}
