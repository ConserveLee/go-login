package tools

import (
	"sync"
)

type MemCache struct {
	sync.Map
}

var (
	instance  *MemCache
	cacheOnce sync.Once
)

func GetCache() *MemCache {
	cacheOnce.Do(func() {
		instance = &MemCache{}
	})
	return instance
}

func (m *MemCache) Has(u string, pwd string) bool {
	realPwd, ok := m.Load(u)
	if !ok {
		return false
	}
	return realPwd == pwd
}
