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
						weatherState += "溫度: "+i.ElementValue +"°C\n"
				case "HUMD":
						hm,err := strconv.ParseFloat(i.ElementValue,64)
						if(err==nil){
							hm = hm*100
							weatherState += "相對溼度: "+fmt.Sprintf("%.0f", hm)+"%\n"
						}
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
						weatherState += "最高溫: "+i.ElementValue +"°C\n"
				case "D_TN":
						weatherState += "最低溫: "+i.ElementValue +"°C\n"
				default:
			}
		}	   	
	}
	return weatherState
}

func decodingmore(b []byte) string{
	var t StationObsResponse
	json.Unmarshal([]byte(b), &t)
	var weatherState string = ""
	nowWeather := t.Records.Location[0].WeatherElement

	for _,i := range nowWeather{
		if(i.ElementValue != "-99"){
			switch i.ElementName{
				case "WDIR":	
						weatherState += "風向: "+i.ElementValue +"度\n"
				case "WDSD":	
						weatherState += "風速: "+i.ElementValue +"m/s\n"
				case "PRES":
						weatherState += "測站氣壓: "+i.ElementValue+"hPA\n"
				case "H_FX":	
						weatherState += "小時最大陣風風速: "+i.ElementValue +"m/s\n"	
				case "H_XD":	
						weatherState += "小時最大陣風風向: "+i.ElementValue +"度\n"
				case "H_FXT":
						weatherState += "小時最大陣風時間: "+i.ElementValue[0:len([]rune(i.ElementValue))-2]+ ":" +i.ElementValue[len([]rune(i.ElementValue))-2:len([]rune(i.ElementValue))]+"\n"				
				case "D_TX":
						weatherState += "本日最高溫: "+i.ElementValue+"°C\n"					
				case "D_TXT":
						weatherState += "本日最高溫時間: "+i.ElementValue[0:len([]rune(i.ElementValue))-2]+ ":" +i.ElementValue[len([]rune(i.ElementValue))-2:len([]rune(i.ElementValue))]+"\n"				
				case "D_TN":
						weatherState += "本日最低溫: "+i.ElementValue+"°C\n"					
				case "D_TNT":
						weatherState += "本日最低溫時間: "+i.ElementValue[0:len([]rune(i.ElementValue))-2]+ ":" +i.ElementValue[len([]rune(i.ElementValue))-2:len([]rune(i.ElementValue))]+"\n"						
				case "D_TS":
						weatherState += "日照時數: "+i.ElementValue+"H\n"  
				default:
			}
		}	   	
	}	
	return weatherState
}

func getTime(b []byte) string{
	var t StationObsResponse
	json.Unmarshal([]byte(b), &t)
	var weatherState string = ""
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

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(myLat+"\n—————————————\n"+decoding(body)+getTime(body))).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "幫助") || (message.Text == "Help") || (message.Text == "help") || (message.Text == "HELP"){
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("幫助\n—————————————\n·輸入'天氣' → 查詢天氣\n·輸入'更多' → 查詢詳細天氣\n·輸入地區名→切換測站地點\n—————————————\n目前可查詢地區:\n基隆 新北 台北 新竹 台中 \n嘉義 高雄 宜蘭 花蓮 台東 \n澎湖 金門 連江 蘭嶼\n")).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "詳細") || (message.Text == "詳細資料") || (message.Text == "更多"){
					resp, _ := http.Get("https://opendata.cwb.gov.tw/api/v1/rest/datastore/O-A0003-001?Authorization=CWB-5392AACA-249F-4D87-9657-11BA88B990E8&locationName="+myLat)
					defer resp.Body.Close()  //關閉連線
					body, _ := ioutil.ReadAll(resp.Body) //讀取body的內容

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(myLat+"\n—————————————\n"+decoding(body)+decodingmore(body)+getTime(body))).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "基隆") || (message.Text == "基隆市") || (message.Text == "基隆縣"){
					myLat = "基隆"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "新北") || (message.Text == "新北市"){
					myLat = "板橋"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "台北") || (message.Text == "臺北") || (message.Text == "台北市") || (message.Text == "臺北市"){
					myLat = "臺北"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				/*}else if (message.Text == "桃園") || (message.Text == "桃園市"){
					myLat = "桃園"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				}else if (message.Text == "新竹") || (message.Text == "新竹縣") || (message.Text == "新竹市"){
					myLat = "新竹"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				/*}else if (message.Text == "苗栗") || (message.Text == "苗栗縣") || (message.Text == "苗栗國"){
					myLat = "苗栗"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				}else if (message.Text == "台中") || (message.Text == "臺中") || (message.Text == "台中市") || (message.Text == "臺中市"){
					myLat = "臺中"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				/*}else if (message.Text == "彰化") || (message.Text == "彰化縣"){
					myLat = "員林"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				/*}else if (message.Text == "南投") || (message.Text == "南投縣"){
					myLat = "南投"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				/*}else if (message.Text == "雲林") || (message.Text == "雲林縣"){
					myLat = "斗六"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				/*}else if (message.Text == "嘉義縣"){
					myLat = "民雄"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				}else if (message.Text == "嘉義") || (message.Text == "嘉義市"){
					myLat = "嘉義"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				/*}else if (message.Text == "台南") || (message.Text == "臺南") || (message.Text == "台南市") || (message.Text == "臺南市"){
					myLat = "臺南"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				}else if (message.Text == "高雄") || (message.Text == "高雄市"){
					myLat = "高雄"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				/*}else if (message.Text == "屏東") || (message.Text == "屏東縣"){
					myLat = "屏東"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}*/
				}else if (message.Text == "宜蘭") || (message.Text == "宜蘭縣"){
					myLat = "宜蘭"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "花蓮") || (message.Text == "花蓮縣"){
					myLat = "花蓮"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "台東") || (message.Text == "臺東") || (message.Text == "台東縣") || (message.Text == "臺東縣"){
					myLat = "臺東"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "澎湖") || (message.Text == "澎湖縣"){
					myLat = "澎湖"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "金門") || (message.Text == "金門縣"){
					myLat = "金門"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "連江") || (message.Text == "連江縣") || (message.Text == "馬祖"){
					myLat = "馬祖"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else if (message.Text == "蘭嶼"){
					myLat = "蘭嶼"
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("目前測站: "+myLat)).Do(); err != nil {
						log.Print(err)
					}
				}else{
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text+"\n—————————————\n需要幫助嗎？\n輸入'幫助'吧！")).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}
