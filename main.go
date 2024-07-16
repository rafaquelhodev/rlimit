package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type TokenBucket struct {
	id           string
	mu           sync.Mutex
	tokens       int64
	maxTokens    int64
	refillPeriod int64
	cron         chan bool
	subs         []chan bool
}

type TokenBucketManager struct {
	tokenBuckets map[string]*TokenBucket
	mutex        sync.Mutex
}

func newTokenBucket(id string, maxTokens int64, refillPeriod int64) *TokenBucket {
	bucket := &TokenBucket{
		id:           id,
		tokens:       0,
		maxTokens:    maxTokens,
		refillPeriod: refillPeriod,
		cron:         make(chan bool),
		subs:         make([]chan bool, 0),
	}
	fmt.Printf("refill period  = %d\n", refillPeriod)
	bucket.startCron()
	return bucket
}

func (tb *TokenBucket) startCron() {
	ticker := time.NewTicker(time.Duration(tb.refillPeriod) * time.Millisecond)

	go func() {
		for {
			select {
			case <-tb.cron:
				ticker.Stop()
				return
			case <-ticker.C:
				if tb.tokens < tb.maxTokens {
					// fmt.Printf("[TOKEN REFIL] | beforeCurTokens = %f\n", tb.tokens)
					tb.tokens += 1
					fmt.Printf("[TOKEN REFIL] | currTokens = %d\n", tb.tokens)

					if len(tb.subs) > 0 {
						sub := tb.subs[0]
						tb.subs = tb.subs[1:]
						sub <- true
					}
				}
			}
		}
	}()
}

func (tb *TokenBucket) tokenSubscribe() chan bool {
	subChannel := make(chan bool)
	tb.subs = append(tb.subs, subChannel)
	return subChannel
}

func (tb *TokenBucket) waitAvailable() bool {
	tb.mu.Lock()

	if tb.tokens > 0 {
		fmt.Printf("[CONSUMING TOKEN] - id = %s\n", tb.id)
		tb.tokens -= 1
		tb.mu.Unlock()
		return true
	}

	fmt.Printf("[WAITING TOKEN] - id %s\n", tb.id)

	ch := tb.tokenSubscribe()

	tb.mu.Unlock()

	<-ch

	fmt.Printf("[NEW TOKEN AVAILABLED] - id %s\n", tb.id)

	tb.tokens -= 1

	return true
}

func newUserTokenBucketManager() *TokenBucketManager {
	return &TokenBucketManager{tokenBuckets: make(map[string]*TokenBucket)}
}

func (tm *TokenBucketManager) getTokenBucket(reqID string, maxTokens int64, refillRate int64) *TokenBucket {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	pk := fmt.Sprintf("%s-%d-%d", reqID, maxTokens, refillRate)

	if bucket, ok := tm.tokenBuckets[pk]; ok {
		return bucket
	}

	bucket := newTokenBucket(reqID, maxTokens, refillRate)
	tm.tokenBuckets[pk] = bucket
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

		options := strings.Split(temp, " ")

		resp := ""
		if len(options) < 3 {
			resp = "INVALID: reqID maxTokens refillRate are required\n"
		} else {
			reqID := options[0]

			maxToken, err := strconv.ParseInt(options[1], 10, 64)
			if err != nil {
				resp = "INVALID: maxTokens should be an integer"
			}

			refillRate, err := strconv.ParseInt(options[2], 10, 64)
			if err != nil {
				resp = "INVALID: refillRate should be an integer"
			}

			if resp == "" {
				bucket := tm.getTokenBucket(reqID, maxToken, refillRate)

				bucket.waitAvailable()

				resp = "AVAILABLE\n"
			}

		}

		c.Write([]byte(string(resp)))
	}
	c.Close()
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

	// var wg sync.WaitGroup

	// bucket := tokenManager.getTokenBucket("test")

	// for i := 0; i < 50; i++ {
	// 	wg.Add(1)
	// 	go func() {
	// 		bucket.waitAvailable()
	// 		defer wg.Done()
	// 	}()
	// }

	// wg.Wait()

}
