package common

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (u *User) Login() (bool, error) {
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
			if username == u.Username && password == u.Password {
				return true, nil // 找到匹配项，返回 true
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return false, fmt.Errorf("failed to read file %s: %w", filename, err)
	}

	return false, nil // 遍历完文件没有找到匹配项，返回 false
}
