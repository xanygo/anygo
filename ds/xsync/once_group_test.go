//  Copyright(C) 2025 github.com/hidu  All Rights Reserved.
//  Author: hidu <duv123+git@gmail.com>
//  Date: 2025-10-13

package xsync_test

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/xanygo/anygo/ds/xsync"
	"github.com/xanygo/anygo/safely"
	"github.com/xanygo/anygo/xt"
)

func TestOnceGroup_Do(t *testing.T) {
	var g xsync.OnceGroup[string, any]
	v, err, _ := g.Do("key", func() (any, error) {
		return "bar", nil
	})
	if got, want := fmt.Sprintf("%v (%T)", v, v), "bar (string)"; got != want {
		t.Errorf("Do = %v; want %v", got, want)
	}
	xt.NoError(t, err)
}

func TestOnceGroup_DoErr(t *testing.T) {
	var g xsync.OnceGroup[string, any]
	someErr := errors.New("some error")
	v, err, _ := g.Do("key", func() (any, error) {
		return nil, someErr
	})
	xt.SamePtr(t, someErr, err)
	xt.Empty(t, v)
}

func TestOnceGroup_Panic(t *testing.T) {
	var g xsync.OnceGroup[string, any]
	v, err, _ := g.Do("key", func() (any, error) {
		panic("hello")
	})
	xt.Empty(t, v)
	xt.Error(t, err)
	t.Logf("err= %s\n", err.Error())
	var te *safely.PanicErr
	ok := errors.As(err, &te)
	xt.True(t, ok)
	t.Logf("te: %#v", te.TraceData())
}

func TestOnceGroup_DoDupSuppress(t *testing.T) {
	var g xsync.OnceGroup[string, any]
	var wg1, wg2 sync.WaitGroup
	c := make(chan string, 1)
	var calls int32
	fn := func() (any, error) {
		if atomic.AddInt32(&calls, 1) == 1 {
			// First invocation.
			wg1.Done()
		}
		v := <-c
		c <- v // pump; make available for any future calls

		time.Sleep(10 * time.Millisecond) // let more goroutines enter Do

		return v, nil
	}

	const n = 10
	wg1.Add(1)
	for range n {
		wg1.Add(1)
		wg2.Go(func() {
			wg1.Done()
			v, err, _ := g.Do("key", fn)
			if err != nil {
				t.Errorf("Do error: %v", err)
				return
			}
			if s, _ := v.(string); s != "bar" {
				t.Errorf("Do = %T %v; want %q", v, v, "bar")
			}
		})
	}
	wg1.Wait()
	// At least one goroutine is in fn now and all of them have at
	// least reached the line before the Do.
	c <- "bar"
	wg2.Wait()
	if got := atomic.LoadInt32(&calls); got <= 0 || got >= n {
		t.Errorf("number of calls = %d; want over 0 and less than %d", got, n)
	}
}

func TestForget(t *testing.T) {
	var g xsync.OnceGroup[string, any]

	var (
		firstStarted  = make(chan struct{})
		unblockFirst  = make(chan struct{})
		firstFinished = make(chan struct{})
	)

	go func() {
		g.Do("key", func() (i any, e error) {
			close(firstStarted)
			<-unblockFirst
			close(firstFinished)
			return
		})
	}()
	<-firstStarted
	g.Forget("key")

	unblockSecond := make(chan struct{})
	secondResult := g.DoChan("key", func() (i any, e error) {
		<-unblockSecond
		return 2, nil
	})

	close(unblockFirst)
	<-firstFinished

	thirdResult := g.DoChan("key", func() (i any, e error) {
		return 3, nil
	})

	close(unblockSecond)
	<-secondResult
	r := <-thirdResult
	if r.Val != 2 {
		t.Errorf("We should receive result produced by second call, expected: 2, got %d", r.Val)
	}
}
