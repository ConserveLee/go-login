package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const (
	credentialsFile = "user_test"
	concurrency     = 500
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var targetURL string

func init() {
	flag.StringVar(&targetURL, "url", "http://192.168.50.211:3001/v1", "The target URL for the authentication service")
}

func sendRequest(ctx context.Context, username, password string, client *http.Client, requestCount *int64, mu *sync.Mutex) error {
	credentials := Credentials{
		Username: username,
		Password: password,
	}
	requestBody, err := json.Marshal(credentials)
	if err != nil {
		fmt.Printf("Error marshaling JSON for user %s: %v\n", username, err)
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("Error creating request for user %s to %s: %v\n", username, targetURL, err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request for user %s to %s: %v\n", username, targetURL, err)
		return err
	}
	defer resp.Body.Close()

	// 成功请求增加计数器
	mu.Lock()
	*requestCount++
	mu.Unlock()

	return nil
}

func main() {
	flag.Parse()

	file, err := os.Open(credentialsFile)
	if err != nil {
		fmt.Println("Error opening credentials file:", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var users []Credentials
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		if len(parts) == 2 {
			users = append(users, Credentials{Username: parts[0], Password: parts[1]})
		} else {
			fmt.Println("Skipping invalid line:", line)
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading credentials file:", err)
		return
	}

	startTime := time.Now()
	var requestCount int64
	var mu sync.Mutex
	client := &http.Client{}
	ctx := context.Background()
	group, ctx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, concurrency)

	for _, user := range users {
		select {
		case <-ctx.Done(): // 如果 context 被取消，则停止启动新的 goroutine
			break
		default:
			semaphore <- struct{}{} // 获取信号量
			group.Go(func() error {
				defer func() { <-semaphore }() // 释放信号量
				return sendRequest(ctx, user.Username, user.Password, client, &requestCount, &mu)
			})
		}
	}

	if err := group.Wait(); err != nil {
		fmt.Printf("Encountered errors during requests: %v\n", err)
	}

	elapsedTime := time.Since(startTime)
	qps := float64(requestCount) / elapsedTime.Seconds()

	fmt.Printf("Total requests: %d\n", requestCount)
	fmt.Printf("Elapsed time: %s\n", elapsedTime)
	fmt.Printf("QPS: %.2f\n", qps)
}
