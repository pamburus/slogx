package mock

import (
	"slices"
	"sync"
)

// NewCallLog returns a new [CallLog].
func NewCallLog() *CallLog {
	return &CallLog{}
}

// CallLog is a log of method of function calls that can be used to build assertions.
type CallLog struct {
	mu    sync.Mutex
	calls CallList
}

// Calls returns all recorded calls in the log.
func (l *CallLog) Calls() CallList {
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

// ---

// CallList is a list of recorded calls.
type CallList []any

// WithoutTime returns a new [CallList] with all time fields set to zero.
func (l CallList) WithoutTime() CallList {
	l = slices.Clone(l)

	for i := range l {
		if item, ok := l[i].(anyCloner); ok {
			l[i] = item.cloneAny()
		}

		if item, ok := l[i].(timeRemover); ok {
			l[i] = item.withoutTime()
		}
	}

	return l
}
