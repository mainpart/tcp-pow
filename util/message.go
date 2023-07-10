package util

import (
	"fmt"
	"strconv"
	"strings"
)

// Обмен сообщениями идет с разделением \n  заголовка, ресурса и содержимого.
func (m *Message) Stringify() string {
	return fmt.Sprintf("%d\n%s\n%s", m.Header, m.Resource, m.Payload)
}

func ParseMessage(str string) (*Message, error) {
	str = strings.TrimSpace(str)
	var msgType int

	parts := strings.Split(str, "\n")
	if len(parts) < 1 {
		return nil, fmt.Errorf("unknown protocol")
	}

	msgType, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("error parsing header")
	}
	msg := Message{
		Header: msgType,
	}

	if len(parts) > 1 {
		msg.Resource = parts[1]
	}
	if len(parts) > 2 {
		msg.Payload = strings.Join(parts[2:], "\n")
	}

	return &msg, nil
}
