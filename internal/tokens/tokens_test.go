package tokens

import (
	"testing"
	"time"
)

func TestCreateNewManager(t *testing.T) {
	got := NewUserTokenBucketManager()
	if len(got.tokenBuckets) != 0 {
		t.Fatalf(`Initialized tokenBuckets must be empty, got: %+v`, got.tokenBuckets)
	}
}

func TestWaitAvailableShouldCreateNewBucket(t *testing.T) {
	tbm := NewUserTokenBucketManager()

	if len(tbm.tokenBuckets) != 0 {
		t.Fatalf(`Initialized tokenBuckets must be empty, got: %+v`, tbm.tokenBuckets)
	}

	tbm.WaitAvailable("test", 1, 1)

	if len(tbm.tokenBuckets) != 1 {
		t.Fatalf(`TokenBuckets must contain one bucket, got: %d`, len(tbm.tokenBuckets))
	}
}

func TestIsOnlyAvailableAfterTheTokenIsRefilled(t *testing.T) {
	tbm := NewUserTokenBucketManager()

	start := time.Now()

	tbm.WaitAvailable("test", 1, 300)

	elapsed := time.Since(start)

	if elapsed < 300*time.Millisecond {
		t.Fatalf(`The elapsed time must be lower than 300, got: %+v`, elapsed)
	}
}

func TestDoesNotNeedToWaitWhenTokenIsAvailable(t *testing.T) {
	tbm := NewUserTokenBucketManager()

	bucket := tbm.getTokenBucket("test", 1, 3000)
	bucket.tokens += 1

	start := time.Now()

	tbm.WaitAvailable("test", 1, 3000)

	elapsed := time.Since(start)

	if elapsed > 10*time.Millisecond {
		t.Fatalf(`The request took longer than 10ms.`)
	}
}
