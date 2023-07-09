package util

import (
	"context"
	"github.com/go-redis/redis/v8"
	"time"
)

type Config struct {
	ServerHost            string        `envconfig:"SERVER_HOST" yaml:"server_host"`
	ServerPort            int           `envconfig:"SERVER_PORT" yaml:"server_port"`
	CacheHost             string        `envconfig:"CACHE_HOST" yaml:"cache_host"` // host of redis-cache
	CachePort             int           `envconfig:"CACHE_PORT" yaml:"cache_port"` // port of redis-cache
	HashcashZerosCount    int           `yaml:"hashcash_zeros_count"`              // count of zeros that server requires from client in hash on PoW
	HashcashDuration      time.Duration `yaml:"hashcash_duration"`                 // lifetime of challenge
	HashcashMaxIterations int           `yaml:"hashcash_max_iterations"`
	SaltLen               int           `yaml:"salt_len"`
}

const (
	RequestChallenge  = iota // from client to server - request new challenge from server
	ResponseChallenge        // from server to client - message with challenge for client
	RequestResource          // from client to server - message with solved challenge
	ResponseResource         // from server to client - message with useful info is solution is correct, or with error if not
)

type Message struct {
	Header  int    //type of message
	Payload string //payload, could be json, quote or be empty
}

// https://en.wikipedia.org/wiki/Hashcash
type Hashcash struct {
	Version    int
	ZerosCount int
	Date       int64
	Resource   string
	Rand       string
	Counter    int
}

type RedisCache struct {
	ctx    context.Context
	client *redis.Client
}
