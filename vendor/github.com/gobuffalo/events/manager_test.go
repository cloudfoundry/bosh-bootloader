package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_manager_Listen(t *testing.T) {
	r := require.New(t)

	m := DefaultManager().(*manager)

	df, err := m.Listen("foo", func(e Event) {})
	r.NoError(err)

	l, ok := m.listeners.Load("foo")
	r.True(ok)
	r.NotNil(l)

	_, err = m.Listen("foo", func(Event) {})
	r.Error(err)

	df()

	l, ok = m.listeners.Load("foo")
	r.False(ok)
	r.Nil(l)
}

func Test_manager_Emit(t *testing.T) {
	r := require.New(t)

	m := DefaultManager().(*manager)

	max := 5
	wg := &sync.WaitGroup{}
	wg.Add(max)

	moot := &sync.Mutex{}
	var es []Event
	m.Listen("foo", func(e Event) {
		moot.Lock()
		defer moot.Unlock()
		es = append(es, e)
		wg.Done()
	})

	for i := 0; i < max; i++ {
		err := m.Emit(Event{
			Kind: "FOO",
		})
		r.NoError(err)
	}

	// because wg.Wait can potentially hang here if there's
	// a bug, let's make sure that doesn't happen
	ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
	var _ = cf // don't want the cf, but lint complains if i don't keep it

	go func() {
		<-ctx.Done()
		if ctx.Err() != nil {
			panic("test ran too long")
		}
	}()
	wg.Wait()
	r.Len(es, max)

	for _, e := range es {
		r.Equal("foo", e.Kind)
	}
}

func Test_manager_Emit_Panic(t *testing.T) {
	r := require.New(t)

	m := DefaultManager().(*manager)

	max := 5
	wg := &sync.WaitGroup{}
	wg.Add(max)

	moot := &sync.Mutex{}
	var es []Event
	m.Listen("foo", func(e Event) {
		moot.Lock()
		defer moot.Unlock()
		es = append(es, e)
		wg.Done()
		panic("x")
	})

	for i := 0; i < max; i++ {
		err := m.Emit(Event{
			Kind: "FOO",
		})
		r.NoError(err)
	}

	// because wg.Wait can potentially hang here if there's
	// a bug, let's make sure that doesn't happen
	ctx, cf := context.WithTimeout(context.Background(), 2*time.Second)
	var _ = cf // don't want the cf, but lint complains if i don't keep it

	go func() {
		<-ctx.Done()
		if ctx.Err() != nil {
			panic("test ran too long")
		}
	}()
	wg.Wait()
	r.Len(es, max)

	for _, e := range es {
		r.Equal("foo", e.Kind)
	}
}

func Test_Manager_List(t *testing.T) {
	r := require.New(t)

	m := DefaultManager().(*manager)
	m.Listen("b", func(Event) {})
	m.Listen("a", func(Event) {})
	m.Listen("c", func(Event) {})

	list, err := m.List()
	r.NoError(err)
	r.Len(list, 3)
	r.Equal([]string{"a", "b", "c"}, list)
}
