package circbuf

import (
	"runtime/internal/atomic"
)

type element[T any] struct {
	id    uint64
	value T
}

type CircularBuffer[T any] struct {
	reading uint32
	idx     uint64
	buf     []*element[T]
}

func NewCircularBuffer[T any](size int) *CircularBuffer[T] {
	if size <= 0 {
		return nil
	}

	return &CircularBuffer[T]{
		buf: make([]*element[T], size),
	}
}

func NewCircularBufferInit[T any](size int, zeroT func() T) *CircularBuffer[T] {
	if size <= 0 {
		return nil
	}

	buf := make([]*element[T], size)
	for i := 0; i < size; i++ {
		buf[i] = &element[T]{value: zeroT()}
	}

	return &CircularBuffer[T]{
		buf: buf,
	}
}

func (b *CircularBuffer[T]) reserveID() uint64 {
	return atomic.Xadd64(&b.idx, 1)
}

func (b *CircularBuffer[T]) PushLater() (ok bool, ret T) {
	if v := atomic.Load(&b.reading); v == 1 {
		ok = false
		return
	}

	id := b.reserveID()
	idx := int((id - 1) % uint64(cap(b.buf)))

	b.buf[idx].id = id

	return true, b.buf[idx].value
}

func (b *CircularBuffer[T]) Push(value T) bool {
	if v := atomic.Load(&b.reading); v == 1 {
		return false
	}

	id := b.reserveID()
	idx := int((id - 1) % uint64(cap(b.buf)))

	if b.buf[idx] == nil || b.buf[idx].id == 0 {
		b.buf[idx] = &element[T]{}
	}

	b.buf[idx].id = id
	b.buf[idx].value = value

	return true
}

func (b *CircularBuffer[T]) Read() []T {
	if !atomic.Cas(&b.reading, 0, 1) {
		return nil
	}
	defer atomic.Store(&b.reading, 0)

	var out []T
	b.forEachEvent(func(e T) {
		out = append(out, e)
	})
	return out
}

func (b *CircularBuffer[T]) ForEachEvent(fn func(e T)) {
	if fn == nil || !atomic.Cas(&b.reading, 0, 1) {
		return
	}
	defer atomic.Store(&b.reading, 0)

	b.forEachEvent(fn)
}

func (b *CircularBuffer[T]) forEachEvent(fn func(e T)) {
	h := b.FindHead()
	i := h
	for {
		e := b.buf[i]
		if e == nil || e.id == 0 {
			break
		}

		fn(e.value)
		b.buf[i].id = 0

		i = (i + 1) % cap(b.buf)

		if i == h {
			break
		}
	}

	b.idx = 0
}

func (b *CircularBuffer[T]) FindHead() int {
	var smallId uint64 = 1<<64 - 1
	var smallIdx int

	for idx, e := range b.buf {
		if e == nil || e.id == 0 {
			break
		}

		if e.id < smallId {
			smallId = e.id
			smallIdx = idx
		}
	}

	return smallIdx
}
