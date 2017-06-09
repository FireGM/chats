// thanks to nuuls
//geted some code from https://github.com/nuuls/chatlog
package twitch

import (
	"errors"
	"fmt"
	"html"
	"html/template"
	"regexp"
	"strconv"
	"strings"
)

const (
	privMsg       = "PRIVMSG"
	clearMsg      = "CLEARCHAT"
	usernoticeMsg = "USERNOTICE"
	userStateMsg  = "USERSTATE"
	joinMsg       = "JOIN"
	roomStateMsg  = "ROOMSTATE"
)

var messageRegexp = regexp.MustCompile(`^(?P<tags>.*):((?P<username>.*)!(.*)@(.*))?(\.*)tmi.twitch.tv (?P<typeOfMessage>PRIVMSG|USERNOTICE|CLEARCHAT|USERSTATE|JOIN|ROOMSTATE) #(?P<channel>\w*)( (:?)(?P<message>.*))?`)
var subNamesRegexp = messageRegexp.SubexpNames()

const emotUrl = "https://static-cdn.jtvnw.net/emoticons/v1/%s/1.0"

// type Emote struct {
// 	ID     int
// 	Name   string
// 	Ranges [2]int
// }

// type emotes []Emote

// func (e emotes) Len() int {
// 	return len(e)
// }
// func (e emotes) Swap(i, j int) {
// 	e[i], e[j] = e[j], e[i]
// }
// func (e emotes) Less(i, j int) bool {
// 	return e[i].Ranges[0] < e[j].Ranges[0]
// }

type Emote struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Type  string `json:"type"`
	Count int    `json:"count"`
}

type Message struct {
	Badges      map[string]string `json:"badges"`
	Tags        map[string]string `json:"tags"`
	Color       string            `json:"color"`
	DisplayName string            `json:"display_name"`
	Emotes      map[string]*Emote `json:"emotes"`
	Mod         int               `json:"mod"`
	RoomID      int               `json:"room_id"`
	Channel     string            `json:"channel"`
	Text        string            `json:"text"`
	// Time           time.Time         `json:"time"`
	User           string        `json:"user"`
	Type           string        `json:"type"`
	RawMessage     string        `json:"raw_message"`
	TextWithEmotes template.HTML `json:"text_with_emotes"`
	NicknameRender template.HTML `json:"nickname_render"`
	FullRender     template.HTML `json:"full_render"`
}

func (m *Message) IsFromUser() bool {
	if m.Type == privMsg {
		return true
	}
	return false
}

func (m Message) GetChatName() string {
	return "twitch"
}

func (m *Message) GetChannelName() string {
	return m.Channel
}

func (m *Message) GetUID() string {
	return m.User
}

func (m *Message) GetRenderSmiles() template.HTML {
	escaped := html.EscapeString(m.Text)
	if len(m.Emotes) < 1 {
		return template.HTML(template.HTML(escaped))
	}
	for _, emote := range m.Emotes {
		var url string
		switch emote.Type {
		case "twitch":
			url = fmt.Sprintf(emotUrl, emote.ID)

		}
		escaped = strings.Replace(escaped, emote.Name,
			`<img class="smile" src="`+url+`" alt="`+emote.Name+`">`,
			-1)
	}
	return template.HTML(escaped)
}

func (m *Message) GetRenderMessHTML() template.HTML {
	if m.TextWithEmotes != "" {
		return m.TextWithEmotes
	}
	m.TextWithEmotes = template.HTML(`<div class="message twitch-message">` + m.GetRenderSmiles() + `</div>`)
	return m.TextWithEmotes
}

func (m *Message) GetRenderNicknameHTML() template.HTML {
	if m.NicknameRender != "" {
		return m.NicknameRender
	}
	badge := ""
	for k, v := range m.Badges {
		url := ""
		alt := ""
		if k == "subscriber" {
			url, alt = getBadgeSubscriber(m.Channel, v)
		} else {
			url = badges[k][v].ImageURL1x
			alt = badges[k][v].Title
		}
		if url != "" {
			badge += fmt.Sprintf(`<img class="badge twitch-badge" src="%s" alt="%s">`, url, alt)
		}
	}
	nickname := fmt.Sprintf(`<p class="nickname twitch-nickname">%s</p>`,
		html.EscapeString(m.GetUserFrom()))
	m.NicknameRender = template.HTML(`<div class="nickname-badge twitch-nickname-badge">` +
		badge + nickname + `</div>`)
	return m.NicknameRender
}

func (m *Message) GetRenderFullHTML() template.HTML {
	if m.FullRender != "" {
		return m.FullRender
	}
	m.FullRender = `<div class="full-message twitch-full-message">` + m.GetRenderNicknameHTML() +
		`<span class="separator twitch-separator"></span>` + m.GetRenderMessHTML() + `</div>`
	return m.FullRender
}

func (m *Message) GetTextMessage() string {
	return m.Text
}

func (m *Message) ToUser(username string) bool {
	return strings.Contains(strings.ToLower(m.Text), "@"+strings.ToLower(username))
}

func (m *Message) GetUserFrom() string {
	return m.User
}

func (m *Message) GetColorNickname() string {
	return m.Color
}

func (m *Message) IsClearMessage() bool {
	return m.Type == clearMsg
}

func (m *Message) IsModerator() (bool, string) {
	v, ok := m.Badges["moderator"]
	return ok, badges["moderator"][v].ImageURL1x
}

func (m *Message) IsSubscriber() (bool, string) {
	if v, ok := m.Badges["subscriber"]; ok {
		url, _ := getBadgeSubscriber(m.Channel, v)
		return true, url
	}
	return false, ""
}

func ParseMessage(line string) (Message, error) {
	var m Message
	m, err := messageTypeParse(line)
	if err != nil {
		return m, err
	}
	if m.Type == clearMsg {
		m.User = m.Text
	}
	m.RawMessage = line
	return m, nil
}

func messageTypeParse(line string) (Message, error) {
	var message Message
	data, err := getDataFromLine(line)
	if err != nil {
		return message, err
	}
	message.Type = data["typeOfMessage"]
	message.Channel = data["channel"]
	message.Text = data["message"]
	message.User = data["username"]
	err = parseTags(data["tags"], &message)
	if err != nil {
		return message, err
	}
	return message, nil
}

func getDataFromLine(line string) (map[string]string, error) {
	md := map[string]string{}
	r := messageRegexp.FindAllStringSubmatch(line, 1)
	if len(r) < 1 {
		return nil, errors.New("unsupported line of chat")
	}
	for i, n := range r[0] {
		md[subNamesRegexp[i]] = n
	}
	if v, ok := md["typeOfMessage"]; !ok || v == "" {
		return nil, errors.New("Unsupported message type")
	}
	return md, nil
}

func parseTags(tagsString string, m *Message) error {
	tagsString = strings.TrimPrefix(tagsString, "@")
	tagsString = strings.TrimSuffix(tagsString, " ")
	tags := strings.Split(tagsString, ";")
	m.Tags = make(map[string]string)
	for _, tag := range tags {
		spl := strings.SplitN(tag, "=", 2)
		value := ""
		if len(spl) > 1 {
			value = spl[1]
		}
		switch spl[0] {
		case "":
			continue
		case "badges":
			m.Badges = getBadges(value)
		case "color":
			m.Color = value
		case "display-name":
			m.DisplayName = value
		case "emotes":
			m.Emotes = getEmotions(value, m.Text)
		case "mod":
			m.Mod, _ = strconv.Atoi(value)
		case "room-id":
			m.RoomID, _ = strconv.Atoi(value)
		default:
			m.Tags[spl[0]] = value
		}
	}
	return nil
}

func getEmotions(emotions string, text string) map[string]*Emote {
	emotes := map[string]*Emote{}

	if emotions == "" {
		return emotes
	}

	runes := []rune(text)

	emoteSlice := strings.Split(emotions, "/")
	for i := range emoteSlice {
		spl := strings.Split(emoteSlice[i], ":")
		pos := strings.Split(spl[1], ",")
		sp := strings.Split(pos[0], "-")
		start, _ := strconv.Atoi(sp[0])
		end, _ := strconv.Atoi(sp[1])
		id := spl[0]
		e := &Emote{
			Type:  "twitch",
			ID:    id,
			Count: strings.Count(emoteSlice[i], "-"),
			Name:  string(runes[start : end+1]),
		}

		emotes[e.Name] = e
	}
	return emotes
}

func getBadges(badges string) map[string]string {
	var badgeMap = make(map[string]string)
	for _, badge := range strings.Split(badges, ",") {
		b := strings.Split(badge, "/")
		if len(b) < 2 {
			continue
		}
		// badgeVersion, _ := strconv.Atoi(b[1])
		badgeMap[b[0]] = b[1]
	}
	return badgeMap
}

func getColor(line string, m *Message) (string, error) {
	colorName, color, next := getNextAndSplited(line)
	if colorName != "color" {
		return "", errors.New("no color")
	}
	m.Color = color
	return next, nil
}

func getDisplayName(line string, m *Message) (string, error) {
	name, displayName, next := getNextAndSplited(line)
	if name != "display-name" {
		return "", errors.New("No display name")
	}
	m.DisplayName = displayName
	return next, nil
}

func getNextAndSplited(line string) (string, string, string) {
	splitted := strings.SplitN(line[1:], ";", 1)
	next := splitted[1]
	splited := strings.Split(splitted[0], "=")
	return splited[0], splited[1], next
}
