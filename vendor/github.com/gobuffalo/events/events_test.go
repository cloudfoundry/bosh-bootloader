package events

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func Test_List(t *testing.T) {
	r := require.New(t)

	boss = nil
	_, err := List()
	r.Error(err)

	boss = DefaultManager()
	NamedListen("b", func(Event) {})
	NamedListen("c", func(Event) {})
	NamedListen("a", func(Event) {})

	names, err := List()
	r.NoError(err)
	r.Len(names, 3)
	r.Equal([]string{"a", "b", "c"}, names)
}

func Test_Emit_and_Listen(t *testing.T) {
	r := require.New(t)

	boss = DefaultManager()

	max := 5
	wg := &sync.WaitGroup{}
	wg.Add(max)

	moot := &sync.Mutex{}
	var es []Event
	Listen(func(e Event) {
		moot.Lock()
		defer moot.Unlock()
		es = append(es, e)
		wg.Done()
	})

	for i := 0; i < max; i++ {
		err := Emit(Event{
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

func Test_EmitError(t *testing.T) {
	r := require.New(t)

	boss = DefaultManager()

	max := 5
	wg := &sync.WaitGroup{}
	wg.Add(max)

	moot := &sync.Mutex{}
	var es []Event
	Listen(func(e Event) {
		moot.Lock()
		defer moot.Unlock()
		es = append(es, e)
		wg.Done()
	})

	for i := 0; i < max; i++ {
		err := EmitError("foo", errors.New("bar"), i)
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
		r.Equal("foo:err", e.Kind)
	}
}
