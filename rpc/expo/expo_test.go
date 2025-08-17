package expo

import "testing"

func TestExpo_Send(t *testing.T) {
	expo := GetExpoClient()
	expo.AddToken("ExponentPushToken[rza8G6KK5rKfCbb8g8lrt8]")
	err := expo.Send("This is a test notification")
	if err != nil {
		t.Errorf("failed to send message: %v", err)
	}
}
