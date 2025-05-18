package tools

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"
)

type missCache struct {
	sync.RWMutex
	data      map[string]*MissUser
	MaxSize   int
	evictChan chan string
}

var (
	mc       *missCache
	missOnce sync.Once
)

type MissUser struct {
	Username   string
	Password   string
	MissCount  int8
	lastAccess time.Time
}

func GetMissCache() *missCache {
	missOnce.Do(func() {
		mc = &missCache{
			data:      make(map[string]*MissUser),
			MaxSize:   2000,
			evictChan: make(chan string, 100),
		}
		go mc.autoEvict()
	})
	return mc
}

func (m *missCache) autoEvict() {
	for username := range m.evictChan {
		m.Lock()
		delete(m.data, username)
		m.Unlock()
	}
}

func (mu *MissUser) SaveMissUser() {
	//count := mu.MissCount + 1
	//fmt.Printf("记录用户%s密码错误次数：%d \n", mu.Username, count)
	GetMissCache().Store(mu.Username, &MissUser{
		Username:  mu.Username,
		Password:  mu.Password, //更新最后一次错误密码
		MissCount: mu.MissCount + 1,
	})
}

func (m *missCache) Store(username string, user *MissUser) {
	m.Lock()
	defer m.Unlock()
	// 执行LFU淘汰
	if len(m.data) >= m.MaxSize {
		// 找到最少使用的条目
		var oldest string
		minAccess := int8(^uint8(0) >> 1)
		for k, v := range m.data {
			if v.MissCount < minAccess {
				oldest = k
				minAccess = v.MissCount
			}
		}
		m.evictChan <- oldest
	}

	user.lastAccess = time.Now()
	m.data[username] = user
}

func (m *missCache) GetMissUserInfo(username string, pwd string) (*MissUser, bool) {
	m.Lock()
	defer m.Unlock()
	mUser, ok := m.data[username]
	if !ok {
		return &MissUser{ //第一次初始化
			Username:  username,
			Password:  pwd,
			MissCount: 0,
		}, false
	}
	// 类型检查
	mUser.MissCount++
	mUser.lastAccess = time.Now()
	return mUser, true
}

func (mu *MissUser) LoginV2() (bool, error) {
	filename := "test/real_users"
	file, err := os.Open(filename)
	if err != nil {
		return false, fmt.Errorf("failed to open file %s: %w", filename, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			username := strings.TrimSpace(parts[0])
			password := strings.TrimSpace(parts[1])
			if username == mu.Username && password == mu.Password {
				return true, nil // 找到匹配项，返回 true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to read file %s: %w", filename, err)
	}
	go mu.SaveMissUser()
	return false, nil // 遍历完文件没有找到匹配项，返回 false
}
