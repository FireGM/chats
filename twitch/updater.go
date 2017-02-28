package twitch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

const badgesURL = "https://badges.twitch.tv/v1/badges/global/display"
const chanInfoURL = "https://api.twitch.tv/kraken/channels/"
const badgesSubURLFormat = "https://badges.twitch.tv/v1/badges/channels/%d/display"

//need for getting subscribe badges
var clientID = ""

var client = http.Client{Timeout: time.Second * 20}
var once = sync.Once{}

var badges = map[string]map[string]Badge{}
var badgesSub = map[string]map[string]Badge{}

type Badge struct {
	ImageURL1x string `json:"image_url_1x"`
	ImageURL2x string `json:"image_url_2x"`
	ImageURL4x string `json:"image_url_4x"`
	Title      string `json:"title"`
}

type BadgeSets struct {
	Versions map[string]Badge `json:"versions"`
}

type BadgeResp struct {
	BadgeSets map[string]BadgeSets `json:"badge_sets"`
}

func goUpdater() {
	go updater()
}
func updater() {
	err := requestAndParse()
	if err != nil {
		panic(err)
	}
	for _ = range time.Tick(time.Minute * 60) {
		requestAndParse()
	}
}

func getBadge(name string, v string) (string, string) {
	if badge, ok := badges[name][v]; ok {
		return badge.ImageURL1x, badge.Title
	}
	return "", ""
}

func requestAndParse() error {
	res, err := client.Get(badgesURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	var bresp BadgeResp
	err = json.Unmarshal(b, &bresp)
	if err != nil {
		return err
	}
	for k, v := range bresp.BadgeSets {
		badges[k] = v.Versions
	}
	return nil
}

func getBadgeSubscriber(channel, version string) (string, string) {
	if badge, ok := badgesSub[channel][version]; ok {
		return badge.ImageURL1x, badge.Title
	}
	id, err := getChannelID(channel)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	bs, err := requestAndParseSubBadges(id)
	if err != nil {
		log.Println(err)
		return "", ""
	}
	badgesSub[channel] = map[string]Badge{}
	for k, v := range bs.BadgeSets["subscriber"].Versions {
		badgesSub[channel][k] = v
	}
	return badgesSub[channel][version].ImageURL1x, badgesSub[channel][version].Title
}

func requestAndParseSubBadges(id int) (BadgeResp, error) {
	var bs BadgeResp
	res, err := client.Get(fmt.Sprintf(badgesSubURLFormat, id))
	if err != nil {
		return bs, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return bs, err
	}

	err = json.Unmarshal(b, &bs)
	if err != nil {
		return bs, err
	}
	if _, ok := bs.BadgeSets["subscriber"]; !ok {
		return bs, errors.New("No subscribers")
	}

	return bs, nil
}

func getChannelID(channel string) (int, error) {
	req, _ := http.NewRequest("GET", chanInfoURL+channel, nil)
	req.Header.Add("Client-ID", clientID)
	res, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}
	var m struct {
		ID int `json:"_id"`
	}
	err = json.Unmarshal(b, &m)
	if err != nil {
		return 0, err
	}
	if m.ID == 0 {
		log.Println(channel, string(b))
		return 0, errors.New("No id")
	}
	return m.ID, nil
}
