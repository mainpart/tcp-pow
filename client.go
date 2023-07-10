package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/mainpart/tcp-pow/util"
	"io"
	"net"
	"time"
)

func main() {
	fmt.Println("start client")

	configInst, err := util.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("error load config:", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)

	address := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)

	err = runClient(ctx, address)
	if err != nil {
		fmt.Println("client error:", err)
	}
}

func runClient(ctx context.Context, address string) error {

	// client will send new request every 5 seconds endlessly
	for {
		message, err := handleConnectionClient(ctx, address)
		if err != nil {
			return err
		}
		fmt.Println("quote result:", message)
		time.Sleep(5 * time.Second)
	}
}

// создаем два запроса - в одном из них запрашиваем pow задачу
// во втором - даем ответ
func handleConnectionClient(ctx context.Context, address string) (string, error) {

	// соединились с хостом
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	resource, err := util.GetRandSalt(10)
	if err != nil {
		return "", fmt.Errorf("err getting resource: %w", err)
	}

	// запросили задачку
	err = util.SendMsg(util.Message{
		Header:   util.RequestChallenge,
		Resource: resource,
	}, conn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	msgStr, err := readConnMsg(conn)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}

	// закрыли соединение
	_ = conn.Close()
	msg, err := util.ParseMessage(string(msgStr))
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}

	var hashcash util.Hashcash
	err = json.Unmarshal([]byte(msg.Payload), &hashcash)
	if err != nil {
		return "", fmt.Errorf("err parse hashcash: %w", err)
	}
	fmt.Println("got hashcash:", hashcash)

	// посчитали задачку
	conf := ctx.Value("config").(*util.Config)
	hashcash, err = hashcash.ComputeHashcash(conf.HashcashMaxIterations)
	if err != nil {
		return "", fmt.Errorf("err compute hashcash: %w", err)
	}
	fmt.Println("hashcash computed:", hashcash)
	byteData, err := json.Marshal(hashcash)
	if err != nil {
		return "", fmt.Errorf("err marshal hashcash: %w", err)
	}

	// снова соединились с хостом
	conn, err = net.Dial("tcp", address)
	if err != nil {
		return "", err
	}

	// отправили результат
	err = util.SendMsg(util.Message{
		Header:   util.RequestResource,
		Resource: resource,
		Payload:  string(byteData),
	}, conn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	// прочитали ответ
	msgStr, err = readConnMsg(conn)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
	msg, err = util.ParseMessage(string(msgStr))
	if err != nil {
		return "", fmt.Errorf("err parse msg: %w", err)
	}
	return msg.Payload, nil
}

func readConnMsg(connect net.Conn) ([]byte, error) {
	buf := make([]byte, 0, 4096) // big buffer
	tmp := make([]byte, 256)     // using small tmo buffer for demonstrating
	for {
		n, err := connect.Read(tmp)
		if err != nil {
			if err != io.EOF {
				return nil, err
			}
			break
		}
		buf = append(buf, tmp[:n]...)
	}
	return buf, nil
}
