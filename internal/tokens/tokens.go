package tokens

import (
	"fmt"
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

func NewUserTokenBucketManager() *TokenBucketManager {
	return &TokenBucketManager{tokenBuckets: make(map[string]*TokenBucket)}
}

func (tm *TokenBucketManager) CreateTokenBucket(bucketID string, maxTokens int64, refillRate int64) *TokenBucket {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if bucket, ok := tm.tokenBuckets[bucketID]; ok {
		return bucket
	}

	bucket := newTokenBucket(bucketID, maxTokens, refillRate)
	tm.tokenBuckets[bucketID] = bucket
	return bucket
}

func (tm *TokenBucketManager) GetTokenBucket(bucketID string) *TokenBucket {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if bucket, ok := tm.tokenBuckets[bucketID]; ok {
		return bucket
	}

	return nil
}

func (tm *TokenBucketManager) WaitAvailable(bucketID string) error {
	bucket := tm.GetTokenBucket(bucketID)
	if bucket == nil {
		return fmt.Errorf("bucket %s not found", bucketID)
	}

	bucket.waitAvailable()
	return nil
}
