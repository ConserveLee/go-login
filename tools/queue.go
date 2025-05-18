package tools

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

//1.插入队列
//2.errGroup处理队列
//3.存入memCache

const (
	queueSize  = 3000
	maxWorkers = 1500
)

type User struct {
	Username string
	Password string
}

var (
	timeoutThreshold = 600 * time.Millisecond
	workerCount      int32
	userQueue        chan User
	queueOnce        sync.Once
	workersOnce      sync.Once
)

func initQueue() {
	userQueue = make(chan User, queueSize)
}

func GetQueue() chan User {
	queueOnce.Do(initQueue)
	return userQueue
}

func StartWorkersOnce() {
	workersOnce.Do(func() {
		for i := 0; i < maxWorkers; i++ {
			atomic.AddInt32(&workerCount, 1)
			go worker()
		}
	})
}

func worker() {
	defer atomic.AddInt32(&workerCount, -1)

	for {
		select {
		case user, ok := <-userQueue:
			if !ok {
				return
			}
			_ = ProcessAndCache(context.Background(), user)
		case <-time.After(5 * time.Second):
			return
		}
	}
}

func ProcessAndCache(ctx context.Context, user User) error {
	// 创建带超时的第三方调用上下文
	_, cancel := context.WithTimeout(ctx, timeoutThreshold)
	defer cancel()
	FakeToWait()
	missUser := MissUser{
		Username: user.Username,
		Password: user.Password,
	}
	// 保存到缓存（线程安全)
	//fmt.Printf("worker开始处理用户%s\n", user.Username)
	if hasUser, _ := missUser.LoginV2(); hasUser {
		GetCache().Store(user.Username, user.Password)
		//fmt.Println("存入用户", user.Username)
	}
	return nil
}
