package listen

import (
	"log"
	"time"

	data "notice/api/blockbeat"
	"notice/api/notification"
)

func StartListen() {
	go func() {
		for range time.Tick(time.Second * 10) {
			msg := data.BeatBlockData()
			if msg == "" {
				continue
			}
			err := notification.SendNotification(msg, "news")
			if err != nil {
				log.Println(err)
			}
			data.StartNow = time.Now().Unix()
		}
	}()
}
