package youtube

import (
	"time"
)

type PageInfo struct {
	TotalResults int `json:"totalResults"`
}

type Snippet struct {
	ChannelTitle string `json:"channelTitle"`
}

type ChannelInfo struct {
	ID struct {
		VideoID string `json:"videoId"`
	} `json:"id"`
	Snippet Snippet `json:"snippet"`
}
type ChanResp struct {
	PageInfo PageInfo      `json:"pageInfo"`
	Items    []ChannelInfo `json:"items"`
}

type LiveStreamingDetails struct {
	ActiveLiveChatID string `json:"activeLiveChatId"`
}

type VideoInfo struct {
	LiveStreamingDetails LiveStreamingDetails `json:"liveStreamingDetails"`
}

type StreamResp struct {
	PageInfo PageInfo    `json:"pageInfo"`
	Items    []VideoInfo `json:"items"`
}

type MessageSnippet struct {
	Type              string    `json:"type"`
	HasDisplayContent bool      `json:"hasDisplayContent"`
	DisplayMessage    string    `json:"displayMessage"`
	PublishedAt       time.Time `json:"publishedAt"`
}

type MessageAuthorDetails struct {
	ChannelID       string `json:"channelId"`
	DisplayName     string `json:"displayName"`
	IsChatOwner     bool   `json:"isChatOwner"`
	IsChatSponsor   bool   `json:"isChatSponsor"`
	IsChatModerator bool   `json:"isChatModerator"`
}

type MessageResp struct {
	Snippet       MessageSnippet       `json:"snippet"`
	AuthorDetails MessageAuthorDetails `json:"authorDetails"`
}

type MessagesResp struct {
	PollingIntervalMillis int           `json:"pollingIntervalsMillis`
	PageInfo              PageInfo      `json:"pageInfo"`
	Items                 []MessageResp `json:"items"`
}
