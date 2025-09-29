package main

import (
	"testing"
	"time"

	data "notice/api/blockbeat"
	"notice/api/expo"
	"notice/api/listen"
)

func TestStartListenFunction(t *testing.T) {
	// 初始化 Expo 客户端
	client := expo.GetExpoClient()

	// 添加测试 Token
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"
	err := client.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
	}

	// 设置初始时间戳（确保能获取到新消息）
	data.StartNow = time.Now().Add(-1 * time.Hour).Unix()

	t.Log("=== 测试 StartListen 函数 ===")

	// 启动监听
	listen.StartListen()

	// 等待一个周期让 goroutine 执行
	t.Log("等待 15 秒让监听器执行...")
	time.Sleep(15 * time.Second)

	t.Log("监听器已启动，如果有新消息将会发送推送通知")
	t.Log("注意：这是一个长期运行的 goroutine，测试完成后会继续运行")
}

func TestBeatBlockDataFunction(t *testing.T) {
	t.Log("=== 测试 BeatBlockData 函数 ===")

	// 设置较早的时间戳确保能获取到消息
	data.StartNow = time.Now().Add(-24 * time.Hour).Unix()

	msg := data.BeatBlockData()

	if msg == "" {
		t.Log("当前没有新消息")
	} else {
		t.Logf("获取到消息: %s", msg)
	}
}

func TestSendNotificationWithCurrentToken(t *testing.T) {
	t.Log("=== 测试使用当前 Token 发送通知 ===")

	client := expo.GetExpoClient()
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"

	// 验证 Token
	err := client.CheckToken(testToken)
	if err != nil {
		t.Logf("Token 验证失败: %v", err)
		return
	}

	// 添加 Token
	err = client.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
		return
	}

	// 发送测试消息
	testMsg := "测试消息 - " + time.Now().Format("2006-01-02 15:04:05")
	err = client.Send(testMsg)
	if err != nil {
		t.Errorf("发送通知失败: %v", err)
	} else {
		t.Log("通知发送成功!")
	}
}

func TestManualListenCycle(t *testing.T) {
	t.Log("=== 手动测试监听周期 ===")

	client := expo.GetExpoClient()
	testToken := "ExponentPushToken[BJT_nyGTdP40fCCji-oTH-]"

	// 添加 Token
	err := client.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
	}

	// 模拟 StartListen 中的逻辑
	msg := data.BeatBlockData()
	t.Logf("获取到的消息: %s", msg)

	if msg != "" {
		err := client.Send(msg)
		if err != nil {
			t.Errorf("发送失败: %v", err)
		} else {
			t.Log("消息发送成功")
		}
		data.StartNow = time.Now().Unix()
		t.Logf("更新时间戳: %d", data.StartNow)
	} else {
		t.Log("没有新消息需要发送")
	}
}
