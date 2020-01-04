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
	"time"
	//"strconv"

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
				now := time.Now()
				if ((message.Text == "Time") || (message.Text == "time") || (message.Text == "TIME")){
					local1, err1 := time.LoadLocation("UTC")
					if err1 != nil {
						fmt.Println(err1)
					}
					local2, err2 := time.LoadLocation("Asia/Taipei")
					if err2 != nil {
						fmt.Println(err2)
					}
					local3, err3 := time.LoadLocation("America/Los_Angeles")
					if err3 != nil {
						fmt.Println(err3)
					}
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("Current Time: "+
													      "\nUTC time: "+now.In(local1).Format(time.UnixDate)+
													      "\nTaipei time: "+now.In(local2).Format(time.UnixDate)+
													      "\nUSA time: "+now.In(local3).Format(time.UnixDate))).Do(); err != nil {
						log.Print(err)
					}
				} else {
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.Text)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	}
}
