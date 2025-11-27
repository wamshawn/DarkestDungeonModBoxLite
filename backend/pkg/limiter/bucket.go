package limiter

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"time"
)

func NewBucket(capacity int) *Bucket {
	if capacity < 1 {
		capacity = runtime.NumCPU() / 2
		if capacity < 1 {
			capacity = 1
		}
	}
	return &Bucket{
		locker:   sync.Mutex{},
		capacity: capacity,
		tokens:   0,
	}
}

type Bucket struct {
	locker   sync.Mutex
	capacity int
	tokens   int
}

func (bucket *Bucket) Take(ctx context.Context, n int) error {
	if n < 1 {
		return errors.New("n must be greater than zero")
	}
	bucket.locker.Lock()
	defer bucket.locker.Unlock()
LOOP:
	if err := ctx.Err(); err != nil {
		return err
	}
	remains := bucket.capacity - bucket.tokens
	if remains >= n {
		bucket.tokens += n
		return nil
	}
	time.Sleep(100 * time.Millisecond)
	goto LOOP
}

func (bucket *Bucket) Release(n int) {
	bucket.locker.Lock()
	defer bucket.locker.Unlock()
	bucket.tokens -= n
}
