package message

import "testing"

func TestMsg(t *testing.T) {
	msg := NewMessage("1")
	println(msg.IsSync())
}
