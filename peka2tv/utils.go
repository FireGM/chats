package peka2tv

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const currentUserURL = "http://peka2.tv/api/user/current"
const userBySlug = "http://peka2.tv/api/stream"

var client = http.Client{Timeout: time.Second * 60}

type UserCurrent struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Block bool   `json:"block"`
	Guest bool   `json:"guest"`
}

type StreamInfo struct {
	Id    int         `json:"id"`
	Owner UserCurrent `json:"owner"`
}

func getCurrentUser(token string) (User, error) {
	var u UserCurrent
	req, err := http.NewRequest("POST", currentUserURL, nil)
	if err != nil {
		return User{}, err
	}
	req.Header.Add("Token", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return User{}, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	err = json.Unmarshal(b, &u)
	if err != nil {
		return User{}, err
	}
	if u.Block {
		return User{}, errors.New("User blocked")
	}
	if u.Guest {
		return User{}, errors.New("invalid token")
	}
	if u.ID == 0 || u.Name == "" {
		return User{}, errors.New("Unknown error")
	}
	return User{ID: u.ID, Name: u.Name}, nil
}

func GetUserIdBySlug(slug string) (int, error) {
	v := url.Values{}
	v.Set("slug", slug)
	res, err := client.PostForm(userBySlug, v)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	var s StreamInfo
	err = json.Unmarshal(b, &s)
	if err != nil {
		return 0, err
	}
	return s.Owner.ID, nil
}
