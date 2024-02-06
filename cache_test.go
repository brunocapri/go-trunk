package trunk

import (
	"fmt"
	"testing"
	"time"
)

// tests the creation of a new cache.
func TestNewCache(t *testing.T) {
	cache, _ := NewCache[int](time.Minute, 32)
	if cache == nil {
		t.Error("NewCache returned nil")
	}
}

// tests adding and getting items from the cache.
func TestAddAndGet(t *testing.T) {
	cache, _ := NewCache[int](time.Minute, 32)
	key := "key1"
	value := 100
	err := cache.Add(key, value)
	if err != nil {
		t.Errorf("Failed to add item to cache: %s", err)
	}

	gotValue, found := cache.Get(key)
	if !found {
		t.Errorf("Expected to find key '%s'", key)
	}
	if gotValue != value {
		t.Errorf("Expected value '%d', got '%d'", value, gotValue)
	}
}

// tests if items are correctly evicted after the interval.
func TestEviction(t *testing.T) {
	evictionInterval := 50 * time.Millisecond
	cache, _ := NewCache[int](evictionInterval, 32)
	key := "key1"
	value := 100

	_ = cache.Add(key, value)

	_, found := cache.Get(key)
	if !found {
		t.Errorf("Expected key '%s' to not be evicted yet", key)
	}

	time.Sleep(evictionInterval + 10*time.Millisecond)

	_, found = cache.Get(key)
	if found {
		t.Errorf("Expected key '%s' to be evicted", key)
	}
}

func TestConcurrentAccess(t *testing.T) {
	cache, _ := NewCache[int](5*time.Minute, 32)
	iterations := 1000
	done := make(chan bool)

	for i := 0; i < iterations; i++ {
		go func(i int) {
			key := "key" + fmt.Sprint(i)
			_ = cache.Add(key, i)
			done <- true
		}(i)
	}

	for i := 0; i < iterations; i++ {
		<-done
	}

	for i := 0; i < iterations; i++ {
		go func(i int) {
			key := "key" + fmt.Sprint(i)
			_, found := cache.Get(key)
			if !found {
				t.Errorf("Key '%s' not found", key)
			}
			done <- true
		}(i)
	}
	for i := 0; i < iterations; i++ {
		<-done
	}
}

// tests adding an item with an empty key.
func TestEmptyKey(t *testing.T) {
	cache, _ := NewCache[int](time.Minute, 32)
	err := cache.Add("", 100)
	if err == nil {
		t.Error("Expected error when adding item with empty key, got nil")
	}
}
