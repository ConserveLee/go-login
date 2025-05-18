package tools

import (
	"math/rand"
	"time"
)

func FakeToWait() {
	time.Sleep(genRandom(350))
}

func FakeToMysql() {
	time.Sleep(genRandom(50))
}

func FakeToRedis() {
	time.Sleep(genRandom(10))
}

func genRandom(millisecond int) time.Duration {
	rand.Seed(time.Now().UnixNano())
	return time.Duration(rand.Intn(millisecond)) * time.Millisecond
}
