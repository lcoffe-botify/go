package runtime

import (
	"runtime/internal/circbuf"
)

const maxEventLogSize = 128

type eventTraceElement struct {
	Time int64
	G    uint64
	M    int64
	log  [maxEventLogSize]byte
}

func (e *eventTraceElement) SetLog(log string) {
	var i int
	var c byte

	for i, c = range []byte(log) {
		if i >= maxEventLogSize-1 {
			return
		}

		e.log[i] = c
	}

	for i++; i < maxEventLogSize; i++ {
		e.log[i] = 0
	}
}

func (e *eventTraceElement) LogBytes() []byte {
	return e.log[:]
}

func initEventTrace() {
	if debug.eventtrace <= 0 {
		return
	}

	sched.eventTrace = circbuf.NewCircularBufferInit(int(debug.eventtrace),
		func() *eventTraceElement { return &eventTraceElement{} })
}

func pushEventTrace(log string) {
	if sched.eventTrace == nil {
		return
	}

	g := getg()

	now := nanotime()
	ok, e := sched.eventTrace.PushLater()
	if !ok {
		return
	}

	e.Time = now
	e.G = g.goid
	e.M = g.m.id
	e.SetLog(log)
}

func printEventTrace() {
	if sched.eventTrace == nil {
		return
	}

	sched.eventTrace.ForEachEvent(func(e *eventTraceElement) {
		print("eventtrace: ", "t=", e.Time, ", g=", e.G, " m=", e.M, ", ")
		gwrite(e.LogBytes())
		println()
	})
}
