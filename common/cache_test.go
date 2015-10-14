package common

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	Timeout    time.Duration = 1 * time.Minute
	InitalSize int           = 10
)

const (
	alphaKey   = "alphaKey"
	alphaValue = "a"
	bravoKey   = "bravoKey"
	bravoValue = "b"
)

// Current time as simulated by the fakeClock function.
var (
	currentTime time.Time
	fakeClock   clock = func() time.Time {
		return currentTime
	}
)

// RemovalListener callback.
var (
	callbackKey     Key
	callbackValue   Value
	removalListener RemovalListener = func(k Key, v Value) {
		callbackKey = k
		callbackValue = v
	}
)

// Test that the removal listener is invoked with the expired key/value.
func TestExpireWithRemovalListener(t *testing.T) {
	callbackKey = nil
	callbackValue = nil
	c := newCache(Timeout, InitalSize, removalListener, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	assert.Equal(t, 1, c.CleanUp())
	assert.Equal(t, alphaKey, callbackKey)
	assert.Equal(t, alphaValue, callbackValue)
}

// Test that the number of removed elements is returned by Expire.
func TestExpireWithoutRemovalListener(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	c.Put(bravoKey, bravoValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	assert.Equal(t, 2, c.CleanUp())
}

func TestPutIfAbsent(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	oldValue := c.PutIfAbsent(alphaKey, alphaValue, 0)
	assert.Nil(t, oldValue)
	oldValue = c.PutIfAbsent(alphaKey, bravoValue, 0)
	assert.Equal(t, alphaValue, oldValue)
}

func TestPut(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	oldValue := c.Put(alphaKey, alphaValue, 0)
	assert.Nil(t, oldValue)
	oldValue = c.Put(bravoKey, bravoValue, 0)
	assert.Nil(t, oldValue)

	oldValue = c.Put(alphaKey, bravoValue, 0)
	assert.Equal(t, alphaValue, oldValue)
	oldValue = c.Put(bravoKey, alphaValue, 0)
	assert.Equal(t, bravoValue, oldValue)
}

func TestReplace(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)

	// Nil is returned when the value does not exist and no element is added.
	assert.Nil(t, c.Replace(alphaKey, alphaValue, 0))
	assert.Equal(t, 0, c.Size())

	// alphaKey is replaced with the new value.
	assert.Nil(t, c.Put(alphaKey, alphaValue, 0))
	assert.Equal(t, alphaValue, c.Replace(alphaKey, bravoValue, 0))
	assert.Equal(t, 1, c.Size())
}

func TestGetUpdatesLastAccessTime(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)

	currentTime = currentTime.Add(Timeout / 2)
	assert.Equal(t, alphaValue, c.Get(alphaKey))
	currentTime = currentTime.Add(Timeout / 2)
	assert.Equal(t, alphaValue, c.Get(alphaKey))
}

func TestDeleteNonExistentKey(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	assert.Nil(t, c.Delete(alphaKey))
}

func TestDeleteExistingKey(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	assert.Equal(t, alphaValue, c.Delete(alphaKey))
}

func TestDeleteExpiredKey(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	assert.Nil(t, c.Delete(alphaKey))
}

// Test that Entries returns the non-expired map entries.
func TestEntries(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	c.Put(bravoKey, bravoValue, 0)
	m := c.Entries()
	assert.Equal(t, 1, len(m))
	assert.Equal(t, bravoValue, m[bravoKey])
}

// Test that Size returns a count of both expired and non-expired elements.
func TestSize(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	c.Put(bravoKey, bravoValue, 0)
	assert.Equal(t, 2, c.Size())
}

func TestGetExpiredValue(t *testing.T) {
	c := newCache(Timeout, InitalSize, nil, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	v := c.Get(alphaKey)
	assert.Equal(t, alphaValue, v)

	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	v = c.Get(alphaKey)
	assert.Nil(t, v)
}

// Test that the janitor invokes CleanUp on the cache and that the
// RemovalListener is invoked during clean up.
func TestJanitor(t *testing.T) {
	keyChan := make(chan Key)
	c := newCache(Timeout, InitalSize, func(k Key, v Value) {
		keyChan <- k
	}, fakeClock)
	c.Put(alphaKey, alphaValue, 0)
	currentTime = currentTime.Add(Timeout).Add(time.Nanosecond)
	c.StartJanitor(time.Millisecond)
	key := <-keyChan
	c.StopJanitor()
	assert.Equal(t, alphaKey, key)
}
