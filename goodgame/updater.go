package goodgame

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const smilesURL = "https://goodgame.ru/js/minified/global.js"

var once = sync.Once{}
var smiles = map[string]Smile{}

type Smile struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Donat     int    `json:"donat"`
	Animated  bool   `json:"animated"`
	ImgBig    string `json:"img_big"`
	ImgGif    string `json:"img_gif"`
	ChannelID int    `json:"channel_id"`
}

type SmileJs struct {
	ID        string      `json:"id"`
	Name      string      `json:"name"`
	Donat     int         `json:"donat"`
	Animated  bool        `json:"animated"`
	ImgBig    string      `json:"img_big"`
	ImgGif    string      `json:"img_gif"`
	ChannelID interface{} `json:"channel_id"`
}

type GlobalJs struct {
	Smiles        []SmileJs            `json:"Smiles"`
	ChannelSmiles map[string][]SmileJs `json:"Channel_Smiles"`
}

func goUpdater() {
	go updaterSmiles()
}

func updaterSmiles() {
	smilesByte := getSmilesJs()
	globalJs := parseJs(smilesByte)
	parseToSmiles(globalJs)
	for _ = range time.Tick(time.Minute * 60) {
		smilesByte := getSmilesJs()
		globalJs := parseJs(smilesByte)
		parseToSmiles(globalJs)
	}
}

func getSmilesJs() []byte {
	client := http.Client{Timeout: time.Second * 20}
	res, err := client.Get(smilesURL)
	if err != nil {
		log.Println(err)
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println(err)
	}
	return b
}

func parseJs(js []byte) GlobalJs {
	trimJs := js[bytes.Index(js, []byte("{")):]
	trimJs = bytes.Replace(trimJs, []byte("Smiles"), []byte(`"smiles"`), 1)
	trimJs = bytes.Replace(trimJs, []byte("Channel_Smiles"), []byte(`"Channel_Smiles"`), 1)
	trimJs = bytes.Replace(trimJs, []byte("timezone_offset"), []byte(`"timezone_offset"`), 1)
	trimJs = bytes.Replace(trimJs, []byte("icons"), []byte(`"icons"`), 1)
	trimJs = bytes.Replace(trimJs, []byte("Content_Width"), []byte(`"Content_Width"`), 1)
	trimJs = bytes.Replace(trimJs, []byte("};"), []byte(`}`), 1)
	var g GlobalJs
	err := json.Unmarshal(trimJs, &g)
	if err != nil {
		log.Println(err)
	}
	return g
}

func parseToSmiles(g GlobalJs) {
	var id int
	var err error
	var chanID int
	for _, v := range g.ChannelSmiles {
		for _, sm := range v {
			id, err = strconv.Atoi(sm.ID)
			if err != nil {
				log.Println(err)
			}
			switch t := sm.ChannelID.(type) {
			case string:
				chanID, err = strconv.Atoi(t)
			case int:
				chanID = t
			}
			smiles[sm.Name] = Smile{id, sm.Name, sm.Donat, sm.Animated, sm.ImgBig, sm.ImgGif, chanID}
		}
	}
	for _, sm := range g.Smiles {
		id, err = strconv.Atoi(sm.ID)
		if err != nil {
			log.Println(err)
		}
		smiles[sm.Name] = Smile{id, sm.Name, sm.Donat, sm.Animated, sm.ImgBig, sm.ImgGif, 0}
	}

}
