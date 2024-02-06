package trunk

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"
)

func TestStressSetGet(t *testing.T) {
	cache, err := NewCache[int](10*time.Second, 10)
	if err != nil {
		t.Fatalf("Failed to create cache: %v", err)
	}

	for i := 0; i < 100; i++ {
		err := cache.Add(fmt.Sprint(i), i)
		if err != nil {
			t.Fatalf("Failed to add item to cache: %v", err)
		}
	}

	var (
		wg     sync.WaitGroup
		errMsg string
		errMu  sync.Mutex
	)

	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for a := 0; a < 1000; a++ {
				k := r.Int() % 100
				val, ok := cache.Get(fmt.Sprint(k))
				if !ok {
					errMu.Lock()
					errMsg = fmt.Sprintf("expected %d but got nil", k)
					errMu.Unlock()
					break
				} else if val != k {
					errMu.Lock()
					errMsg = fmt.Sprintf("expected %d but got %d", k, val)
					errMu.Unlock()
					break
				}
			}
		}()
	}

	wg.Wait()

	if errMsg != "" {
		t.Errorf(errMsg)
	}
}
