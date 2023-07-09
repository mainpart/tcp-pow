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

	// loading config from file and env
	configInst, err := util.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("error load config:", err)
		return
	}

	// init context to pass config down
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)

	address := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)

	// run client
	err = runClient(ctx, address)
	if err != nil {
		fmt.Println("client error:", err)
	}
}

// Run - main function, launches client to connect and work with server on address
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

// handleConnectionClient - scenario for TCP-client
// 1. request challenge from server
// 2. compute hashcash to check Proof of Work
// 3. send hashcash solution back to server
// 4. get result quote from server
// readerConn and writerConn divided to more convenient mock on testing
func handleConnectionClient(ctx context.Context, address string) (string, error) {

	conn, err := net.Dial("tcp", address)
	if err != nil {
		return "", err
	}
	fmt.Println("connected to", address)

	// 1. requesting challenge
	err = util.SendMsg(util.Message{
		Header: util.RequestChallenge,
	}, conn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}

	// reading and parsing response
	msgStr, err := readConnMsg(conn)
	if err != nil {
		return "", fmt.Errorf("err read msg: %w", err)
	}
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

	// 2. got challenge, compute hashcash
	conf := ctx.Value("config").(*util.Config)
	hashcash, err = hashcash.ComputeHashcash(conf.HashcashMaxIterations)
	if err != nil {
		return "", fmt.Errorf("err compute hashcash: %w", err)
	}
	fmt.Println("hashcash computed:", hashcash)
	// marshal solution to json
	byteData, err := json.Marshal(hashcash)
	if err != nil {
		return "", fmt.Errorf("err marshal hashcash: %w", err)
	}

	conn, err = net.Dial("tcp", address)
	if err != nil {
		return "", err
	}

	// 3. send challenge solution back to server
	err = util.SendMsg(util.Message{
		Header:  util.RequestResource,
		Payload: string(byteData),
	}, conn)
	if err != nil {
		return "", fmt.Errorf("err send request: %w", err)
	}
	fmt.Println("challenge sent to server")

	// 4. get result quote from server
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

// readConnMsg - read string message from connection
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
