package goodgame

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

const authURL = "https://goodgame.ru/ajax/login/"
const infoStreamURL = "https://api2.goodgame.ru/streams/"

var client = http.Client{Timeout: time.Second * 20}

type AuthStructToken struct {
	UserID   int    `json:"user_id"`
	Token    string `json:"token"`
	Username string `json:"user_name"`
}

type ChatUser struct {
	UserID   int    `json:"user_id"`
	Username string `json:"user_name"`
}

type GGruct struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

type MessageReq struct {
	ChannelId string `json:"channel_id"`
	Text      string `json:"text"`
}

type UserGG struct {
	ID       int    `json:"id"`
	Token    string `json:"token"`
	Username string `json:"username"`
}

type AuthResp struct {
	Result bool   `json:"result"`
	Return UserGG `json:"return"`
}

type StreamInfo struct {
	Id int `json:"id"`
}

func getUserByLoginPass(login, pass string) UserGG {
	v := url.Values{}
	v.Set("login", login)
	v.Set("password", pass)
	v.Set("return", "user")
	res, err := client.PostForm(authURL, v)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	var resp AuthResp
	err = json.Unmarshal(b, &resp)
	if err != nil {
		log.Println(err)
	}
	if !resp.Result {
		log.Println("can't auth")
	}
	return resp.Return
}

func GetStreamInfo(slug string) (int, error) {
	req, _ := http.NewRequest("GET", infoStreamURL+slug, nil)
	req.Header.Set("Accept", "application/hal+json")
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var s StreamInfo
	err = json.Unmarshal(b, &s)
	if err != nil {
		return 0, err
	}
	if s.Id == 0 {
		return 0, errors.New(string(b))
	}
	return s.Id, nil
}
