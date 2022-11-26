package async

import (
	"errors"
	"testing"
	"time"

	"github.com/reugn/async/internal"
)

func TestFuture(t *testing.T) {
	p := NewPromise[bool]()
	go func() {
		time.Sleep(time.Millisecond * 100)
		p.Success(true)
	}()
	res, err := p.Future().Join()

	internal.AssertEqual(t, res, true)
	internal.AssertEqual(t, err, nil)
}

func TestFutureUtils(t *testing.T) {
	p1 := NewPromise[int]()
	p2 := NewPromise[int]()
	p3 := NewPromise[int]()
	go func() {
		time.Sleep(time.Millisecond * 100)
		p1.Success(1)
		time.Sleep(time.Millisecond * 200)
		p2.Success(2)
		time.Sleep(time.Millisecond * 300)
		p3.Success(3)
	}()
	arr := []Future[int]{p1.Future(), p2.Future(), p3.Future()}
	res := []interface{}{1, 2, 3}
	futRes, _ := FutureSeq(arr).Join()

	internal.AssertEqual(t, res, futRes)
}

func TestFutureFirstCompleted(t *testing.T) {
	p := NewPromise[bool]()
	go func() {
		time.Sleep(time.Millisecond * 1000)
		p.Success(true)
	}()
	timeout := FutureTimer[bool](time.Millisecond * 100)
	futRes, futErr := FutureFirstCompletedOf(p.Future(), timeout).Join()

	internal.AssertEqual(t, false, futRes)
	if futErr == nil {
		t.Fatalf("futErr is nil")
	}
}

func TestFutureTransform(t *testing.T) {
	p1 := NewPromise[int]()
	go func() {
		time.Sleep(time.Millisecond * 100)
		p1.Success(1)
	}()
	future := p1.Future().Map(func(v int) (int, error) {
		return v + 1, nil
	}).FlatMap(func(v int) (Future[int], error) {
		nv := v + 1
		p2 := NewPromise[int]()
		p2.Success(nv)
		return p2.Future(), nil
	}).Recover(func() (int, error) {
		return 5, nil
	})

	res, _ := future.Get(time.Second * 5)
	internal.AssertEqual(t, 3, res)

	res, _ = future.Join()
	internal.AssertEqual(t, 3, res)
}

func TestFutureFailure(t *testing.T) {
	p1 := NewPromise[int]()
	p2 := NewPromise[int]()
	go func() {
		time.Sleep(time.Millisecond * 100)
		p1.Failure(errors.New("Future error"))
		time.Sleep(time.Millisecond * 200)
		p2.Success(2)
	}()
	res, _ := p1.Future().RecoverWith(p2.Future()).Join()

	internal.AssertEqual(t, 2, res)
}

func TestFutureTimeout(t *testing.T) {
	p := NewPromise[bool]()
	go func() {
		time.Sleep(time.Millisecond * 200)
		p.Success(true)
	}()
	future := p.Future()

	_, err := future.Get(time.Millisecond * 50)
	internal.AssertErrorContains(t, err, "timeout")

	_, err = future.Join()
	internal.AssertErrorContains(t, err, "timeout")
}
