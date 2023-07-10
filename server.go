package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"

	"github.com/mainpart/tcp-pow/util"

	"time"
)

func main() {
	fmt.Println("start server")

	configInst, err := util.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("error load config:", err)
		return
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)

	cacheInst, err := util.InitRedisCache(ctx, configInst.CacheHost, configInst.CachePort)
	if err != nil {
		fmt.Println("error init cache:", err)
		return
	}
	ctx = context.WithValue(ctx, "cache", cacheInst)

	rand.Seed(time.Now().UnixNano())
	serverAddress := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)

	err = runServer(ctx, serverAddress)
	if err != nil {
		fmt.Println("server error:", err)
	}
}

func runServer(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer listener.Close()
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accept connection: %w", err)
		}
		go handleConnectionServer(ctx, conn)
	}
}

// обработка ответа сервера
func handleConnectionServer(ctx context.Context, conn net.Conn) {
	fmt.Println("new client:", conn.RemoteAddr())
	defer conn.Close()

	reader := bufio.NewReader(conn)

	buf := make([]byte, 10000)
	len, err := reader.Read(buf)
	if err != nil {
		fmt.Printf("Error reading: %#v\n", err)
		return
	}

	req := string(buf[:len])
	msg, err := ProcessRequest(ctx, req, conn)
	if err != nil {
		fmt.Println("err process request:", err)
		return
	}
	if msg != nil {
		err := util.SendMsg(*msg, conn)
		if err != nil {
			fmt.Println("err send message:", err)
		}
	}
	return

}

// смотрим на тип запроса. Типа два - запрос задачи для ресурса и запрос ресурса
func ProcessRequest(ctx context.Context, msgStr string, conn net.Conn) (*util.Message, error) {
	msg, err := util.ParseMessage(msgStr)
	resource := msg.Resource
	conf := ctx.Value("config").(*util.Config)
	cache := ctx.Value("cache").(util.Cache)

	if err != nil {
		return nil, err
	}
	switch msg.Header {

	case util.RequestChallenge:

		date := time.Now()
		randValue, err := util.GetRandSalt(conf.SaltLen)
		if err != nil {
			return nil, err
		}
		err = cache.Add(randValue+resource, conf.HashcashDuration)
		if err != nil {
			return nil, err
		}

		hashcash := util.Hashcash{
			Version:    1,
			ZerosCount: conf.HashcashZerosCount,
			Date:       date.Unix(),
			Resource:   resource,
			Rand:       randValue,
			Counter:    0,
		}
		hashcashMarshaled, err := json.Marshal(hashcash)
		if err != nil {
			return nil, err
		}
		msg := util.Message{
			Header:   util.ResponseChallenge,
			Resource: resource,
			Payload:  string(hashcashMarshaled),
		}
		return &msg, nil

	case util.RequestResource:
		fmt.Printf("client %s asks resource %s, payload %s\n", conn.RemoteAddr(), resource, msg.Payload)
		var hashcash util.Hashcash
		err := json.Unmarshal([]byte(msg.Payload), &hashcash)
		if err != nil {
			return nil, err
		}
		// смотрим была ли такая задача выдана на ресурс ранее
		// защита от DDOS
		exists, err := cache.Get(hashcash.Rand + hashcash.Resource)
		if err != nil {
			return nil, err
		}
		if !exists {
			return nil, fmt.Errorf("challenge not exists")
		}

		// смотрим чтобы не протух срок запроса ресурса
		if time.Now().Unix()-hashcash.Date > conf.HashcashDuration.Milliseconds()/100 {
			return nil, fmt.Errorf("challenge not exists")
		}

		_, err = hashcash.ComputeHashcash(hashcash.Counter)
		if err != nil {
			return nil, fmt.Errorf("invalid hashcash sum")
		}
		// если все окей - посылаем цитату
		msg := util.Message{
			Header:  util.ResponseResource,
			Payload: util.GetQuote(),
		}
		cache.Delete(hashcash.Rand + hashcash.Resource)
		return &msg, nil

	default:
		return nil, fmt.Errorf("unknown header")
	}
}
