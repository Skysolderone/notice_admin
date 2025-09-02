package expo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/9ssi7/exponent"
)

func TestExpo_Send(t *testing.T) {
	expo := GetExpoClient()
	testToken := "ExponentPushToken[NfJX5OPserm2ZvCAKIMNok]"

	// 1. 验证 Token
	t.Log("=== 验证 Token 有效性 ===")
	err := expo.CheckToken(testToken)
	if err != nil {
		t.Logf("Token 验证失败: %v", err)
	} else {
		t.Log("Token 验证通过")
	}

	// 2. 添加 Token
	t.Log("=== 添加 Token ===")
	err = expo.AddToken(testToken)
	if err != nil {
		t.Logf("添加 Token 失败: %v", err)
	} else {
		t.Log("Token 添加成功")
	}

	// 3. 发送消息（带重试）
	t.Log("=== 发送推送消息 ===")
	err = expo.Send("这是一条测试通知 - " + time.Now().Format("15:04:05"))
	if err != nil {
		t.Errorf("推送失败: %v", err)
	} else {
		t.Log("推送成功")
	}

	// 4. 直接向特定 Token 发送
	t.Log("=== 向特定 Token 发送 ===")
	err = expo.SendToSpecificToken(testToken, "直接推送测试 - "+time.Now().Format("15:04:05"))
	if err != nil {
		t.Logf("直接推送失败: %v", err)
	} else {
		t.Log("直接推送成功")
	}
}

func TestNewPush(t *testing.T) {
	c := exponent.NewClient(exponent.WithAccessToken(""))

	tkn := exponent.MustParseToken("ExponentPushToken[NfJX5OPserm2ZvCAKIMNok]")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		res, err := c.PublishSingle(ctx, &exponent.Message{
			To:       []*exponent.Token{tkn},
			Body:     fmt.Sprintf("This is a test notification:%d", i),
			Data:     exponent.Data{"withSome": "data"},
			Sound:    "default",
			Title:    "Notification Title",
			Priority: exponent.DefaultPriority,
		})
		if err != nil {
			panic(err)
		}
		for _, receipt := range res {
			if receipt.IsOk() {
				println("Notification sent successfully")
			} else {
				println("Notification failed")
			}
		}
	}
}
