package balancer

import (
	"context"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
)

type myStruct struct {
	num int
}

func TestAcquire(t *testing.T) {
	total := 5
	items := make([]*Item[myStruct], total)
	for i := range total {
		s := &myStruct{
			num: i,
		}
		items[i] = NewItem(s, 1) // 1 request per second
	}

	b := New(items)
	assert.Equal(t, b.NumItems(), total)
	assert.Equal(t, b.TotalMaxRequests(), total*1)
	assert.Equal(t, b.TotalFreeRequests(), total*1)

	synctest.Test(t, func(t *testing.T) {
		for i := range total {
			item := b.Acquire()
			assert.Equal(t, item.num, i)
		}

		assert.Equal(t, b.TotalMaxRequests(), total*1)
		assert.Equal(t, b.TotalFreeRequests(), 0)
		assert.Nil(t, b.Acquire())

		time.Sleep(time.Second)
		synctest.Wait()

		for i := range total {
			item := b.Acquire()
			assert.Equal(t, item.num, i)
		}
		assert.Nil(t, b.Acquire())
	})

	for i := range total {
		s := &myStruct{
			num: i,
		}
		items[i] = NewItem(s, 3) // 3 request per second
	}
	b = New(items)
	assert.Equal(t, b.TotalMaxRequests(), total*3)
	assert.Equal(t, b.TotalFreeRequests(), total*3)
	synctest.Test(t, func(t *testing.T) {
		for i := range total {
			for range 3 {
				item := b.Acquire()
				assert.Equal(t, item.num, i)
			}
		}
		assert.Equal(t, b.TotalMaxRequests(), total*3)
		assert.Equal(t, b.TotalFreeRequests(), 0)
		assert.Nil(t, b.Acquire())

		time.Sleep(time.Second)
		synctest.Wait()

		for i := range total {
			for range 3 {
				item := b.Acquire()
				assert.Equal(t, item.num, i)
			}
		}
		assert.Nil(t, b.Acquire())
	})
}

func TestAcquireWait(t *testing.T) {
	total := 5
	items := make([]*Item[myStruct], total)
	for i := range total {
		s := &myStruct{
			num: i,
		}
		items[i] = NewItem(s, 1) // 1 request per second
	}

	ctx := context.Background()
	b := New(items)
	synctest.Test(t, func(t *testing.T) {
		for i := range total {
			item := b.AcquireWait(ctx, 3, time.Second/2)
			assert.Equal(t, item.num, i)
		}

		item := b.AcquireWait(ctx, 3, time.Second/20) // Not enough attempts or too few pauses
		assert.Nil(t, item)

		start := time.Now()
		item = b.AcquireWait(ctx, 3, time.Second/3)
		assert.Greater(t, time.Since(start), time.Duration(0))
		assert.Equal(t, item.num, 0)

		for item != nil {
			item = b.Acquire()
		}

		// if the pause is <=0, the default is 1 second
		item = b.AcquireWait(ctx, 3, -10)
		assert.NotNil(t, item)
		item = b.AcquireWait(ctx, 3, 0)
		assert.NotNil(t, item)
	})
}

func TestShuffle(t *testing.T) {
	total := 5
	items := make([]*Item[myStruct], total)
	start := make([]int, total)
	for i := range total {
		s := &myStruct{
			num: i,
		}
		items[i] = NewItem(s, 1) // 1 request per second
		start[i] = i
	}

	b := New(items)
	b.SetShuffle(true)

	res := make([]int, total)
	for i := range total {
		item := b.Acquire()
		res[i] = item.num
	}
	assert.NotEqual(t, start, res)
}
