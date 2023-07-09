package util

import (
	"fmt"
	"strconv"
	"strings"
)

func (m *Message) Stringify() string {
	return fmt.Sprintf("%d|%s", m.Header, m.Payload)
}

func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)
	var msgType int

	parts := strings.Split(str, "|")
	if len(parts) < 1 || len(parts) > 2 { //only 1 or 2 parts allowed
		return nil, fmt.Errorf("message doesn't match protocol")
	}

	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("cannot parse header")
	}
	msg := Message{
		Header: msgType,
	}

	if len(parts) == 2 {
		msg.Payload = parts[1]
	}
	return &msg, nil
}
