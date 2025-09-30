package main

import (
	"testing"
	"time"

	data "notice/api/blockbeat"
	"notice/api/expo"
)

func TestHTMLFormatNotification(t *testing.T) {
	t.Log("=== 测试 HTML 格式推送通知 ===")

	client := expo.GetExpoClient()
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"

	// 添加 Token
	err := client.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
	}

	// 发送 HTML 格式的测试消息
	htmlMessage := `<b>测试标题</b>
<p>这是一段带有 <strong>粗体</strong> 和 <em>斜体</em> 的文本。</p>
<ul>
<li>列表项 1</li>
<li>列表项 2</li>
</ul>
<p><a href="https://example.com">链接示例</a></p>
<i>` + time.Now().Format("2006-01-02 15:04:05") + `</i>`

	err = client.Send(htmlMessage)
	if err != nil {
		t.Errorf("HTML 格式推送失败: %v", err)
	} else {
		t.Log("HTML 格式推送成功!")
	}
}

func TestBlockbeatHTMLMessage(t *testing.T) {
	t.Log("=== 测试 Blockbeat HTML 消息格式 ===")

	// 设置较早的时间戳确保能获取到消息
	data.StartNow = time.Now().Add(-24 * time.Hour).Unix()

	// 获取带 HTML 格式的消息
	msg := data.BeatBlockData()

	if msg == "" {
		t.Log("当前没有新消息")
		return
	}

	t.Logf("获取到 HTML 格式消息: %s", msg)

	// 发送到推送服务
	client := expo.GetExpoClient()
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"

	err := client.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
	}

	err = client.Send(msg)
	if err != nil {
		t.Errorf("发送 HTML 消息失败: %v", err)
	} else {
		t.Log("HTML 消息发送成功!")
	}
}

func TestDirectHTMLNotification(t *testing.T) {
	t.Log("=== 测试直接发送 HTML 通知 ===")

	client := expo.GetExpoClient()
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"

	// 创建 HTML 格式消息
	htmlMsg := `<b>🚀 重要通知</b>
<p>这是一条包含 HTML 格式的推送通知：</p>
<ul>
<li><strong>加粗文本</strong></li>
<li><em>斜体文本</em></li>
<li><u>下划线文本</u></li>
</ul>
<p><span style="color: #ff0000;">红色文字</span></p>
<hr>
<i>发送时间：` + time.Now().Format("2006-01-02 15:04:05") + `</i>`

	err := client.SendToSpecificToken(testToken, htmlMsg)
	if err != nil {
		t.Errorf("直接发送 HTML 通知失败: %v", err)
	} else {
		t.Log("直接发送 HTML 通知成功!")
	}
}
