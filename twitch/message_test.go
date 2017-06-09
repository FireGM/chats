package twitch

import (
	"html/template"
	"reflect"
	"testing"
)

func TestParseMessage(t *testing.T) {
	type args struct {
		line string
	}
	tests := []struct {
		name    string
		args    args
		want    Message
		wantErr bool
	}{
		{ //check line of ban and delete message command
			name: "clearchat",
			args: args{line: `@ban-duration=10;ban-reason=;room-id=24991333;target-user-id=135206893 :tmi.twitch.tv CLEARCHAT #imaqtpie :edyzetg`},
			want: Message{
				Type: clearMsg,
				// Badges:     map[string]string{},
				Tags: map[string]string{"ban-duration": "10", "ban-reason": "", "target-user-id": "135206893"},
				// Emotes:     map[string]*Emote{},
				RoomID:     24991333,
				Channel:    "imaqtpie",
				Text:       "edyzetg",
				User:       "edyzetg",
				RawMessage: `@ban-duration=10;ban-reason=;room-id=24991333;target-user-id=135206893 :tmi.twitch.tv CLEARCHAT #imaqtpie :edyzetg`,
			},
		},
		{ //check line of standard message from user to channel
			name: "privmsg",
			args: args{line: `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :hashinshin PogChamp`},
			want: Message{
				Type:        privMsg,
				Badges:      map[string]string{"subscriber": "6"},
				Color:       "#E1630E",
				DisplayName: "Haunterxx",
				Tags:        map[string]string{"sent-ts": "1496852224609", "id": "04de4e6b-646d-4e02-94a9-d0ee80c93a3f", "subscriber": "1", "tmi-sent-ts": "1496852221750", "turbo": "0", "user-id": "39647543", "user-type": ""},
				Emotes:      map[string]*Emote{"PogChamp": &Emote{Type: "twitch", ID: "88", Count: 1, Name: "PogChamp"}},
				Mod:         0,
				RoomID:      24991333,
				Channel:     "imaqtpie",
				Text:        "hashinshin PogChamp",
				User:        "haunterxx",
				RawMessage:  `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :hashinshin PogChamp`,
			},
		},
		{
			name:    "not a message",
			args:    args{line: `ggwpnotmessage`},
			want:    Message{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		got, err := ParseMessage(tt.args.line)
		if (err != nil) != tt.wantErr {
			t.Errorf("%q. ParseMessage() error = %v, wantErr %v", tt.name, err, tt.wantErr)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%q. ParseMessage() = \n%v\n, want \n%v\n", tt.name, got, tt.want)
		}
	}
}

func TestMessage_IsFromUser(t *testing.T) {
	type fields struct {
		Type string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "privmsg type",
			fields: fields{Type: privMsg},
			want:   true,
		},
		{
			name:   "clearmessage type",
			fields: fields{Type: clearMsg},
			want:   false,
		},
	}
	for _, tt := range tests {
		m := &Message{
			Type: tt.fields.Type,
		}
		if got := m.IsFromUser(); got != tt.want {
			t.Errorf("%q. Message.IsFromUser() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestMessage_GetChatName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "twitch always",
			want: "twitch",
		},
	}
	for _, tt := range tests {
		m := Message{}
		if got := m.GetChatName(); got != tt.want {
			t.Errorf("%q. Message.GetChatName() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestMessagePublicFuns(t *testing.T) {
	chatLine := `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :hashinshin PogChamp`
	message, err := ParseMessage(chatLine)
	if err != nil {
		t.Errorf("message not parsed ParseMessage() error = %v", err)
	}
	if !message.IsFromUser() {
		t.Error("Message.IsFromUser() return false for line of message")
	}
	if message.GetChannelName() != "imaqtpie" {
		t.Errorf("Message.GetChannel() = %v, want %v", message.GetChannelName(), "imaqtpie")
	}
	if message.GetUID() != "haunterxx" {
		t.Errorf("Message.GetUID() = %v, want %v", message.GetUID(), "haunterxx")
	}

	smileRender := template.HTML(`hashinshin <img class="smile" src="https://static-cdn.jtvnw.net/emoticons/v1/88/1.0" alt="PogChamp">`)
	if message.GetRenderSmiles() != smileRender {
		t.Errorf("Message.GetRenderSmiles() = %v, want %v", message.GetRenderSmiles(), smileRender)
	}

	if message.GetTextMessage() != "hashinshin PogChamp" {
		t.Errorf("Message.GetTextMessage() = %v, want %v", message.GetTextMessage(), "hashinshin PogChamp")
	}

	if message.GetUserFrom() != "haunterxx" {
		t.Errorf("Message.GetUserFrom = %v, want %v", message.GetUserFrom(), "haunterxx")
	}

	if message.GetColorNickname() != "#E1630E" {
		t.Errorf("Message.GetColorNickname() = %v, want %v", message.GetColorNickname(), "#E1630E")
	}

	if message.IsClearMessage() != false {
		t.Errorf("Message.IsClearMessage() = %v, want %v", message.IsClearMessage(), false)
	}

	if mod, _ := message.IsModerator(); mod != false {
		t.Errorf("Message.IsModerator() = %v, want %v", mod, false)
	}

	if sub, _ := message.IsSubscriber(); sub != true {
		t.Errorf("Message.IsSubscriber() = %v, want %v", sub, true)
	}
}

func TestMessage_ToUser(t *testing.T) {
	tests := []struct {
		name        string
		messageLine string
		want        bool
		toUser      string
	}{
		{
			name:        "to user firence",
			messageLine: `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :@firence hashinshin PogChamp`,
			want:        true,
			toUser:      "firence",
		},
		{
			name:        "to user noFirence",
			messageLine: `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :@noFirencehashinshin PogChamp`,
			want:        false,
			toUser:      "firence",
		},
		{
			name:        "to user GriDer",
			messageLine: `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :@GriDer hashinshin PogChamp`,
			want:        true,
			toUser:      "GriDer",
		},
		{
			name:        "to user noGriDer",
			messageLine: `@badges=subscriber/6;color=#E1630E;display-name=Haunterxx;emotes=88:11-18;id=04de4e6b-646d-4e02-94a9-d0ee80c93a3f;mod=0;room-id=24991333;sent-ts=1496852224609;subscriber=1;tmi-sent-ts=1496852221750;turbo=0;user-id=39647543;user-type= :haunterxx!haunterxx@haunterxx.tmi.twitch.tv PRIVMSG #imaqtpie :@noGriDerhashinshin PogChamp`,
			want:        false,
			toUser:      "GriDer",
		},
	}
	for _, tt := range tests {
		m, err := ParseMessage(tt.messageLine)
		if err != nil {
			t.Errorf("Message.ToUser() got err = %v", err)
		}
		if got := m.ToUser(tt.toUser); got != tt.want {
			t.Errorf("%q. Message.ToUser() = %v, want %v", tt.name, got, tt.want)
		}
	}
}
