package youtube

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const channelUrl = `https://www.googleapis.com/youtube/v3/search`
const streamInfoUrl = `https://www.googleapis.com/youtube/v3/videos`
const messagesUrlFormat = `https://www.googleapis.com/youtube/v3/liveChat/messages?liveChatId=%s&part=id,snippet,authorDetails&key=%s`
const messagesInsertUrl = `https://www.googleapis.com/youtube/v3/liveChat/messages`

var client = &http.Client{Timeout: time.Second * 10}

func GetChatIDByChannel(channelId string, apiKey string) (string, error) {
	channelResp, err := getChannelResp(channelId, apiKey)
	if err != nil {
		return "", err
	}
	if channelResp.PageInfo.TotalResults < 1 || len(channelResp.Items) < 1 {
		return "", errors.New("No stream live. Please, use stream with live stream. It's need only one time")
	}
	streamResp, err := getStreamResp(channelResp.Items[0].ID.VideoID, apiKey)
	if err != nil {
		return "", err
	}
	if len(streamResp.Items) < 1 || streamResp.Items[0].LiveStreamingDetails.ActiveLiveChatID == "" {
		return "", errors.New("No chat for stream")
	}
	return streamResp.Items[0].LiveStreamingDetails.ActiveLiveChatID, nil
}

func getChannelResp(channelId string, apiKey string) (ChanResp, error) {
	var chanResp ChanResp
	values := url.Values{}
	values.Set("part", "snippet")
	values.Set("channelId", channelId)
	values.Set("type", "video")
	values.Set("eventType", "live")
	values.Set("key", apiKey)
	req, err := http.NewRequest("GET", channelUrl, nil)
	req.URL.RawQuery = values.Encode()
	if err != nil {
		return chanResp, err
	}
	res, err := client.Do(req)
	if err != nil {
		return chanResp, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return chanResp, err
	}
	err = json.Unmarshal(b, &chanResp)
	if err != nil {
		return chanResp, err
	}
	return chanResp, nil
}

func getStreamResp(streamId string, apiKey string) (StreamResp, error) {
	var streamResp StreamResp
	values := url.Values{}
	values.Set("id", streamId)
	values.Set("part", "liveStreamingDetails")
	values.Set("key", apiKey)
	req, err := http.NewRequest("GET", streamInfoUrl, nil)
	if err != nil {
		return streamResp, err
	}
	req.URL.RawQuery = values.Encode()
	res, err := client.Do(req)
	if err != nil {
		return streamResp, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return streamResp, err
	}
	err = json.Unmarshal(b, &streamResp)
	if err != nil {
		return streamResp, err
	}
	return streamResp, nil
}

func getMessages(chatID string, apiKey string) (MessagesResp, error) {
	var mr MessagesResp
	req, err := http.NewRequest("GET", fmt.Sprintf(messagesUrlFormat, chatID, apiKey), nil)
	if err != nil {
		return mr, err
	}
	res, err := client.Do(req)
	if err != nil {
		return mr, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return mr, err
	}
	err = json.Unmarshal(b, &mr)
	if err != nil {
		return mr, err
	}
	return mr, nil
}

func getSleepTime(t int) time.Duration {
	tt := time.Millisecond
	if t >= 10000 {
		return tt * time.Duration(10000)
	}
	return tt * time.Duration(3000)
}

func sendMessageToChat(chatId, message, token, apiKey string) error {
	body := fmt.Sprintf(`{"snippet": {"liveChatId" : "%s", "type": "textMessageEvent", "textMessageDetails": {"messageText": "%s"}}}`,
		chatId, message)
	req, err := http.NewRequest("POST", messagesInsertUrl, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	values := url.Values{}
	values.Set("part", "snippet")
	values.Set("access_token", token)
	values.Set("key", apiKey) //todo: apiKey required?
	req.URL.RawQuery = values.Encode()
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode != 200 {
		return errors.New(string(b))
	}
	return nil
}
