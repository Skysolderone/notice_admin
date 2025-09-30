package data

import (
	"fmt"
	"strings"
	"time"

	"github.com/mmcdole/gofeed"
)

var StartNow int64

func BeatBlockData() (text string) {
	// resp, err := HttpClient.R().SetQueryParams(map[string]string{
	// 	"size":     "10",
	// 	"page":     "10",
	// 	"type":     "push",
	// 	"language": "cn",
	// }).Get(os.Getenv("blockapi"))
	// if err != nil {
	// 	log.Println(err)
	// }
	// data := BlockbeatData{}
	// sonic.Unmarshal(resp.Body(), &data)
	// fmt.Printf("%#v", data)
	fp := gofeed.NewParser()
	feed, _ := fp.ParseURL("https://api.theblockbeats.news/v1/open-api/home-xml")
	item := feed.Items[0]
	if TimeLocalToUnix(item.Published) > StartNow {
		text += "<b>" + item.Title + "</b>" + "\n"
		msg := strings.ReplaceAll(item.Description, "BlockBeats 消息，", "")
		// 保留 HTML 标签用于格式化显示
		text += msg

		text += "\n<i>" + item.Published + "</i>"
	}

	return
}

type BlockbeatData struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Page int `json:"page"`
		Data []struct {
			ID         int    `json:"id"`
			Title      string `json:"title"`
			Content    string `json:"content"`
			Pic        string `json:"pic"`
			Link       string `json:"link"`
			URL        string `json:"url"`
			CreateTime string `json:"create_time"`
		} `json:"data"`
	} `json:"data"`
}

func TimeLocalToUnix(ts string) int64 {
	// 日期格式
	layout := "Mon, 02 Jan 2006 15:04:05 -0700"
	// 解析日期字符串
	t, err := time.Parse(layout, ts)
	if err != nil {
		fmt.Println("Error parsing date:", err)
	}
	// 获取Unix时间戳
	unixTime := t.Unix()
	// 打印Unix时间戳
	return unixTime
}
