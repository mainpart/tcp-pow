package main

import (
	"bufio"
	"context"
	"strings"

	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"

	"github.com/mainpart/tcp-pow/util"

	"time"
)

func main() {
	fmt.Println("start server")

	// loading config from file and env
	configInst, err := util.LoadConfig("config.yaml")
	if err != nil {
		fmt.Println("error load config:", err)
		return
	}

	// init context to pass config down
	ctx := context.Background()
	ctx = context.WithValue(ctx, "config", configInst)

	cacheInst, err := util.InitRedisCache(ctx, configInst.CacheHost, configInst.CachePort)
	if err != nil {
		fmt.Println("error init cache:", err)
		return
	}
	ctx = context.WithValue(ctx, "cache", cacheInst)

	// seed random generator to randomize order of quotes
	rand.Seed(time.Now().UnixNano())

	// run server
	serverAddress := fmt.Sprintf("%s:%d", configInst.ServerHost, configInst.ServerPort)
	err = runServer(ctx, serverAddress)
	if err != nil {
		fmt.Println("server error:", err)
	}
}

var ErrQuit = errors.New("client requests to close connection")

type Cache interface {
	Add(string, time.Duration) error
	Get(string) (bool, error)
	Delete(string)
}

func runServer(ctx context.Context, address string) error {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	defer listener.Close()
	fmt.Println("listening", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("error accept connection: %w", err)
		}
		go handleConnectionServer(ctx, conn)
	}
}

func handleConnectionServer(ctx context.Context, conn net.Conn) {
	fmt.Println("new client:", conn.RemoteAddr())
	defer conn.Close()

	reader := bufio.NewReader(conn)

	req, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("err read connection:", err)
		return
	}
	msg, err := ProcessRequest(ctx, req, strings.Split(conn.RemoteAddr().String(), ":")[0])
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

func ProcessRequest(ctx context.Context, msgStr string, resource string) (*util.Message, error) {
	msg, err := util.ParseMessage(msgStr)
	if err != nil {
		return nil, err
	}
	switch msg.Header {

	case util.RequestChallenge:
		fmt.Printf("client %s requests challenge\n", resource)
		// create new challenge for client
		conf := ctx.Value("config").(*util.Config)

		cache := ctx.Value("cache").(Cache)
		date := time.Now()

		// add new created rand value to cache to check it later on RequestResource stage
		// with duration in seconds
		randValue, err := util.GetRandSalt(conf.SaltLen)
		if err != nil {
			return nil, fmt.Errorf("err getting salt: %w", err)
		}
		err = cache.Add(randValue, conf.HashcashDuration)
		if err != nil {
			return nil, fmt.Errorf("err add rand to cache: %w", err)
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
			return nil, fmt.Errorf("err marshal hashcash: %v", err)
		}
		msg := util.Message{
			Header:  util.ResponseChallenge,
			Payload: string(hashcashMarshaled),
		}
		return &msg, nil
	case util.RequestResource:
		fmt.Printf("client %s requests resource with payload %s\n", resource, msg.Payload)
		// parse client's solution
		var hashcash util.Hashcash
		err := json.Unmarshal([]byte(msg.Payload), &hashcash)
		if err != nil {
			return nil, fmt.Errorf("err unmarshal hashcash: %w", err)
		}
		// validate hashcash params
		if hashcash.Resource != resource {
			return nil, fmt.Errorf("invalid hashcash resource")
		}
		conf := ctx.Value("config").(*util.Config)
		cache := ctx.Value("cache").(Cache)

		// if rand exists in cache, it means, that hashcash is valid and really challenged by this server in past
		exists, err := cache.Get(hashcash.Rand)
		if err != nil {
			return nil, fmt.Errorf("err get rand from cache: %w", err)
		}
		if !exists {
			return nil, fmt.Errorf("challenge expired or not sent")
		}

		// sent solution should not be outdated
		if time.Now().Unix()-hashcash.Date > conf.HashcashDuration.Milliseconds()/100 {
			return nil, fmt.Errorf("challenge expired")
		}
		//to prevent indefinite computing on server if client sent hashcash with 0 counter
		maxIter := hashcash.Counter
		if maxIter == 0 {
			maxIter = 1
		}
		_, err = hashcash.ComputeHashcash(maxIter)
		if err != nil {
			return nil, fmt.Errorf("invalid hashcash")
		}
		//get random quote
		fmt.Printf("client %s succesfully computed hashcash %s\n", resource, msg.Payload)
		msg := util.Message{
			Header:  util.ResponseResource,
			Payload: util.GetQuote(),
		}
		// delete rand from cache to prevent duplicated request with same hashcash value
		cache.Delete(hashcash.Rand)
		return &msg, nil
	default:
		return nil, fmt.Errorf("unknown header")
	}
}
