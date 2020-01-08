// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"io/ioutil"
	"encoding/json"
	"github.com/line/line-bot-sdk-go/linebot"
)

var bot *linebot.Client

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)
	http.HandleFunc("/callback", callbackHandler)
	port := os.Getenv("PORT")
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}

var myLat string = "新竹"

type StationObsResponse struct {
	Success string `json:"success"`
	Records struct {
		Location []StationObsLocation `json:"location"`
	} `json:"records"`
}

type StationObsLocation struct {
	Lat          string `json:"lat"`
	Lon          string `json:"lon"`
	LocationName string `json:"locationName"`
	StationId    string `json:"stationId"`
	Time         struct {
		ObsTime string `json:"obsTime"`
	} `json:"time"`
	WeatherElement []StationObsElement `json:"weatherElement"`
}

type StationObsElement struct {
	ElementName  string `json:"elementName"`
	ElementValue string `json:"elementValue"`
}

func decoding(b []byte) string{
	var t StationObsResponse
	json.Unmarshal([]byte(b), &t)
	var weatherState string = ""
	nowWeather := t.Records.Location[0].WeatherElement

	for _,i := range nowWeather{
		if(i.ElementValue != "-99"){
			switch i.ElementName{
				case "TEMP":	
						weatherState += "溫度: "+i.ElementValue+"°C\n"
				case "HUMD":
						hm,err := strconv.ParseFloat(i.ElementValue,64)
						if(err==nil){
							hm = hm*100
							weatherState += "相對溼度: "+fmt.Sprintf("%.0f", hm)+"%\n"
						}						
				case "SUN":
						weatherState += "日照時數: "+i.ElementValue+"H\n"  
				case "H_UVI": 
						uvi,err := strconv.ParseFloat(i.ElementValue,64)
						if(err==nil){
							if(uvi == 0){
							}else if(uvi <= 2){
								weatherState += "紫外線指數: "+i.ElementValue+" (低量)\n"
							}else if(uvi <= 5){
								weatherState += "紫外線指數: "+i.ElementValue+" (中量)\n"
							}else if(uvi <= 7){
								weatherState += "紫外線指數: "+i.ElementValue+" (高量)\n"
							}else if(uvi <= 10){
								weatherState += "紫外線指數: "+i.ElementValue+" (過量)\n"
							}else{
								weatherState += "紫外線指數: "+i.ElementValue+" (危險)\n"
							}
						}
				case "24R":
					rain,err := strconv.ParseFloat(i.ElementValue,64)
						if(err==nil && rain != 0){
							weatherState += "累積雨量:"+ i.ElementValue + " ml\n" 
						}
						
				case "D_TX":
						weatherState += "最高溫: "+i.ElementValue[0:len([]rune(i.ElementValue))-1]+"°C\n"
				case "D_TN":
						weatherState += "最低溫: "+i.ElementValue[0:len([]rune(i.ElementValue))-1]+"°C\n"
				default:
			}
		}	   	
	}
	getTime := t.Records.Location[0].Time.ObsTime
	weatherState += "\n更新時間: "+getTime[0:len([]rune(getTime))-3]
	return weatherState
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				
				if (message.Text == "天氣"){
					resp, _ := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/O-A0003-001?Authorization=CWB-5392AACA-249F-4D87-9657-11BA88B990E8&locationName="+myLat)
					defer resp.Body.Close()  //關閉連線
					body, _ := ioutil.ReadAll(resp.Body) //讀取body的內容

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(myLat+"\n----------\n"+decoding(body))).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "台中"){
					myLat = "臺中"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已轉換測站至 '"+myLat+"'")).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "高雄"){
					myLat = "高雄"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("已轉換測站至 '"+myLat+"'")).Do(); err != nil {
						log.Print(err)
					}
				}else{
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}
