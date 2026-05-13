package websocket

import (
	"testing"
)

func TestHubBasic(t *testing.T) {
	hub := &Hub{
		Salas: make(map[string]map[*Cliente]bool),
	}

	roomID := "room1"
	client1 := &Cliente{Nickname: "user1", SalaId: roomID, DeviceId: "dev1", Ip: "1.1.1.1"}
	client2 := &Cliente{Nickname: "user2", SalaId: roomID, DeviceId: "dev2", Ip: "2.2.2.2"}

	hub.Salas[roomID] = make(map[*Cliente]bool)
	hub.Salas[roomID][client1] = true
	hub.Salas[roomID][client2] = true

	// Test GetRoomUserCount
	count := hub.GetRoomUserCount(roomID)
	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}

	// Test GetRoomUsers
	users := hub.GetRoomUsers(roomID)
	if len(users) != 2 {
		t.Errorf("Expected 2 users in list, got %d", len(users))
	}

	// Test IsSessionBlocked
	if !hub.IsSessionBlocked("dev1", "any", "any") {
		t.Errorf("Expected dev1 to be blocked")
	}
	if !hub.IsSessionBlocked("any", "user2", "any") {
		t.Errorf("Expected user2 to be blocked")
	}
	if !hub.IsSessionBlocked("any", "any", "1.1.1.1") {
		t.Errorf("Expected IP 1.1.1.1 to be blocked")
	}
	if hub.IsSessionBlocked("dev3", "user3", "3.3.3.3") {
		t.Errorf("Expected dev3/user3/3.3.3.3 NOT to be blocked")
	}

	// Test k6 exception
	if hub.IsSessionBlocked("any", "K6_VU_1", "any") {
		t.Errorf("Expected K6_VU_1 NOT to be blocked")
	}
}
