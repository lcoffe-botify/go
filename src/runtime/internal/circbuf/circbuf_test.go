package circbuf_test

import (
	"runtime/internal/circbuf"
	"testing"
)

func TestPush(t *testing.T) {
	b := circbuf.NewCircularBuffer[struct{}](0)
	if b != nil {
		t.FailNow()
	}

	b = circbuf.NewCircularBuffer[struct{}](10)
	if b == nil {
		t.FailNow()
	}

	for i := 0; i < 100; i++ {
		if !b.Push(struct{}{}) {
			t.FailNow()
		}
	}
}

func TestFindHead(t *testing.T) {
	b := circbuf.NewCircularBuffer[struct{}](10)
	if b == nil {
		t.FailNow()
	}

	if b.FindHead() != 0 {
		t.FailNow()
	}

	b.Push(struct{}{})
	if b.FindHead() != 0 {
		t.FailNow()
	}

	b.Push(struct{}{})
	if b.FindHead() != 0 {
		t.FailNow()
	}

	for i := 0; i < 10; i++ {
		b.Push(struct{}{})
	}
	if b.FindHead() != 2 {
		t.FailNow()
	}

	b.Push(struct{}{})
	if b.FindHead() != 3 {
		t.FailNow()
	}

	for i := 1; i <= 42; i++ {
		b.Push(struct{}{})
		if b.FindHead() != (3+i)%10 {
			t.FailNow()
		}
	}
}

func TestRead(t *testing.T) {
	type testValue struct {
		x int
	}

	b := circbuf.NewCircularBuffer[*testValue](5)
	if b == nil {
		t.FailNow()
	}

	// empty buffer, empty read
	x := b.Read()
	if len(x) != 0 {
		t.FailNow()
	}

	// push one value, should get it
	b.Push(&testValue{x: 42})
	x = b.Read()
	if len(x) != 1 {
		t.FailNow()
	}
	if x[0].x != 42 {
		t.FailNow()
	}

	// now empty
	x = b.Read()
	if len(x) != 0 {
		t.FailNow()
	}

	// push a couple values without crossing
	b.Push(&testValue{x: 42})
	b.Push(&testValue{x: 43})
	x = b.Read()
	if len(x) != 2 {
		t.FailNow()
	}
	if x[0].x != 42 {
		t.FailNow()
	}
	if x[1].x != 43 {
		t.FailNow()
	}

	// now empty
	x = b.Read()
	if len(x) != 0 {
		t.FailNow()
	}

	// do exactly a full circle
	for i := 0; i < 5; i++ {
		b.Push(&testValue{x: 42 + i})
	}
	x = b.Read()
	if len(x) != 5 {
		t.FailNow()
	}
	if x[0].x != 42 {
		t.FailNow()
	}
	if x[1].x != 43 {
		t.FailNow()
	}
	if x[2].x != 44 {
		t.FailNow()
	}
	if x[3].x != 45 {
		t.FailNow()
	}
	if x[4].x != 46 {
		t.FailNow()
	}

	// now empty
	x = b.Read()
	if len(x) != 0 {
		t.FailNow()
	}

	// go past frontier
	for i := 0; i < 11; i++ {
		b.Push(&testValue{x: 42 + i})
	}
	x = b.Read()
	if len(x) != 5 {
		t.FailNow()
	}
	if x[0].x != 48 {
		t.FailNow()
	}
	if x[1].x != 49 {
		t.FailNow()
	}
	if x[2].x != 50 {
		t.FailNow()
	}
	if x[3].x != 51 {
		t.FailNow()
	}
	if x[4].x != 52 {
		t.FailNow()
	}
}
