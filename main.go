package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

type TokenBucket struct {
	mu           sync.Mutex
	tokens       float64
	maxTokens    float64
	refillPeriod float64
	cron         chan bool
	subs         []chan bool
}

type TokenBucketManager struct {
	tokenBuckets map[string]*TokenBucket
	mutex        sync.Mutex
}

func newTokenBucket(maxTokens float64, refillPeriod float64) *TokenBucket {
	bucket := &TokenBucket{
		tokens:       0,
		maxTokens:    maxTokens,
		refillPeriod: refillPeriod,
		cron:         make(chan bool),
		subs:         make([]chan bool, 0),
	}
	return bucket
}

func newUserTokenBucketManager() *TokenBucketManager {
	return &TokenBucketManager{tokenBuckets: make(map[string]*TokenBucket)}
}

func (tm *TokenBucketManager) getTokenBucket(reqID string) *TokenBucket {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if bucket, ok := tm.tokenBuckets[reqID]; ok {
		return bucket
	}

	bucket := newTokenBucket(1, 1000)
	tm.tokenBuckets[reqID] = bucket
	return bucket
}

func handleConnection(c net.Conn, tm *TokenBucketManager) {
	fmt.Printf("Serving %s\n", c.RemoteAddr().String())

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		temp := strings.TrimSpace(string(netData))
		if temp == "STOP" {
			fmt.Printf("Closing connection %s\n", c.RemoteAddr().String())
			break
		}

		fmt.Printf("Received %s from %s\n", temp, c.RemoteAddr().String())

		reqID := temp

		tm.getTokenBucket(reqID)

	}
}

func startServer(tm *TokenBucketManager) {
	args := os.Args
	if len(args) == 1 {
		fmt.Println("Please provide a port number")
	}

	port := ":" + args[1]

	l, err := net.Listen("tcp4", port)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer l.Close()

	for {
		c, err := l.Accept()
		if err != nil {
			fmt.Println(err)
			return
		}
		go handleConnection(c, tm)
	}
}

func main() {
	tokenManager := newUserTokenBucketManager()
	startServer(tokenManager)
}
