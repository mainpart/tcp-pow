package util

import (
	"crypto/sha1"
	"fmt"
)

func (h Hashcash) Stringify() string {
	return fmt.Sprintf("%d:%d:%d:%s::%s:%d", h.Version, h.ZerosCount, h.Date, h.Resource, h.Rand, h.Counter)
}

func makehash(data string) string {
	hashSha := sha1.New()
	hashSha.Write([]byte(data))
	sum := hashSha.Sum(nil)
	return fmt.Sprintf("%x", sum)
}

// сравнение строки с лидирующими нулями
func IsHashCorrect(hash string, zerosCount int) bool {
	if zerosCount > len(hash) {
		return false
	}
	for _, ch := range hash[:zerosCount] {
		if ch != 48 {
			return false
		}
	}
	return true
}

// алгоритм подсчета hashcash
func (h Hashcash) ComputeHashcash(maxIterations int) (Hashcash, error) {
	for h.Counter <= maxIterations {
		header := h.Stringify()
		hash := makehash(header)
		if IsHashCorrect(hash, h.ZerosCount) {
			return h, nil
		}
		h.Counter++
	}
	return h, fmt.Errorf("max limit")
}
