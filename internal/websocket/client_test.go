package websocket

import (
	"testing"
)

func TestClientBasic(t *testing.T) {
	client := &Cliente{
		Nickname: "test",
		SalaId:   "room1",
		DeviceId: "dev1",
		Ip:       "1.1.1.1",
	}

	if client.Nickname != "test" {
		t.Errorf("Expected nickname test, got %s", client.Nickname)
	}
}
