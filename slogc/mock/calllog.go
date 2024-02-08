package mock

import (
	"slices"
	"sync"
)

func NewCallLog() *CallLog {
	return &CallLog{}
}

type CallLog struct {
	mu    sync.Mutex
	calls []any
}

func (l *CallLog) Calls() []any {
	l.mu.Lock()
	defer l.mu.Unlock()

	return slices.Clone(l.calls)
}

func (l *CallLog) append(call any) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.calls = append(l.calls, call)

	return len(l.calls)
}
