package peka2tv

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sync"
	"time"
)

const bonusStoreURL = "http://peka2.tv/api/store/bonus/list"
const smileURL = "http://peka2.tv/api/smile"
const iconsURL = "http://peka2.tv/api/icon/list"

var once = sync.Once{}
var smiles = map[string]*Smile{}
var nickColors = map[int]string{}
var smilesPerMessage = map[int]int{}
var icons = map[int]Icon{}

type Icon struct {
	ID  int    `json:"id"`
	URL string `json:"url"`
}

type Smile struct {
	Url     string
	BonusId int
}

func (s *Smile) SetBonusId(b int) {
	s.BonusId = b
}

func (s *Smile) SetUrl(url string) {
	s.Url = url
}

type BonusStoreRequest struct {
	Id     int `json:"id"`
	Config struct {
		Smiles []string `json:"smiles"`
		Amount int      `json:"amount"`
		Color  string   `json:"color"`
	} `json:"config"`
	Type string `json:"type"`
}

type SmileRequest struct {
	Code string `json:"code"`
	URL  string `json:"url"`
}

func goUpdater() {
	go updateBonusStore()
}

func updateBonusStore() {
	var bs []BonusStoreRequest
	var sm []SmileRequest
	var ic []Icon
	if err := requestSmiles(&sm); err != nil {
		log.Println(err)
	}
	parseSmiles(sm)
	if err := requestStore(&bs); err != nil {
		log.Println(err)
	}
	if err := requestIcons(&ic); err != nil {
		log.Println(err)
	}
	parseIcons(ic)
	switcher(bs)
	for _ = range time.Tick(time.Minute * 60) {
		requestSmiles(&sm)
		parseSmiles(sm)
		requestStore(&bs)
		switcher(bs)
	}
}

func switcher(bs []BonusStoreRequest) {
	for _, b := range bs {
		switch b.Type {
		case "smiles":
			parseSmilesPerm(b.Id, b.Config.Smiles)
		case "nickColor":
			nickColors[b.Id] = b.Config.Color
		case "smilesPerMessage":
			smilesPerMessage[b.Id] = b.Config.Amount
		}
	}
}

func requestStore(bs *[]BonusStoreRequest) error {
	res, err := client.Get(bonusStoreURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	json.Unmarshal(b, bs)
	return nil
}

func requestSmiles(smiles *[]SmileRequest) error {
	res, err := client.Get(smileURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &smiles) //todo: need pointer?
	if err != nil {
		log.Println(err)
	}
	return nil
}

func parseSmilesPerm(id int, smilesP []string) {
	for _, s := range smilesP {
		if _, ok := smiles[s]; !ok {
			smiles[s] = &Smile{BonusId: id}
			continue
		}
		smiles[s].SetBonusId(id)
	}
}

func parseSmiles(sm []SmileRequest) {
	for _, s := range sm {
		if _, ok := smiles[s.Code]; !ok {
			smiles[s.Code] = &Smile{Url: s.URL}
			continue
		}
		smiles[s.Code].SetUrl(s.URL)
	}
}

func parseNickColor(id int, color string) {
	nickColors[id] = color
}

func requestIcons(ic *[]Icon) error {
	res, err := client.Get(iconsURL)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &ic)
	if err != nil {
		return err
	}
	return nil
}

func parseIcons(ic []Icon) {
	for _, i := range ic {
		icons[i.ID] = i
	}
}
