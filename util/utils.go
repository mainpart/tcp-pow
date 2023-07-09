package util

import (
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"io"
	"math/rand"
	"net/http"
	"os"
)

func LoadConfig(path string) (*Config, error) {
	config := Config{}
	yfile, err := os.ReadFile(path)
	if err != nil {
		return &config, err
	}
	err = yaml.Unmarshal(yfile, &config)
	if err != nil {
		return &config, err
	}
	err = envconfig.Process("", &config)
	return &config, err
}

// SendMsg - Используется как на сервере, так и на клиенте. Отправляет в интерфейс сообщение
func SendMsg(msg Message, conn io.Writer) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}

// GetRandSalt - генерирует строчку из заданный символов (в base64) - используется для того чтобы
// проверить путем кэширования соли, что ранее такое задание выдавалось.
// Позволяет отсеять попытки многократно получить ресурс
func GetRandSalt(length int) (string, error) {
	buf := make([]byte, length)
	_, err := rand.Read(buf)
	if err != nil {
		return "", err
	}
	salt := base64.StdEncoding.EncodeToString(buf)
	return salt[:length], nil
}

func GetQuote() string {
	type Quote struct {
		Id     int
		Quote  string
		Author string
	}

	c := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	res, err := c.Get("https://dummyjson.com/quotes/random")

	if err != nil {
		return fmt.Sprintf("error getting quote %s", err)
	}

	body, err := io.ReadAll(res.Body)

	if err != nil {
		return fmt.Sprintf("error getting quote %s", err)
	}

	var data Quote
	json.Unmarshal(body, &data)
	return data.Quote
}
