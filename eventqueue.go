// MFP - Miulti-Function Printers and scanners toolkit
// Cgo binding for Avahi
//
// Copyright (C) 2024 and up by Alexander Pevzner (pzz@apevzner.com)
// See LICENSE for license terms and conditions
//
// Event queue
//
//go:build linux || freebsd

package avahi

import "sync"

// eventqueue represents a queue of values of some type T.
//
// Values added to the eventqueue using Push method and can
// be retrieved from the eventqueue using a channel.
type eventqueue[T any] struct {
	buf       []T            // Buffered values
	outchan   chan T         // Output channel
	lock      sync.Mutex     // Access lock
	closechan chan struct{}  // Closed to signal goroutine to exit
	closewait sync.WaitGroup // Wait for goroutine to exit
}

// init initializes an eventqueue
func (q *eventqueue[T]) init() {
	q.buf = make([]T, 0, 8)
	q.outchan = make(chan T)
	q.closechan = make(chan struct{})
}

// Push adds a new value to the eventqueue
func (q *eventqueue[T]) Push(v T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.buf = append(q.buf, v)
	if len(q.buf) == 1 {
		q.closewait.Add(1)
		go q.proc()
	}
}

// Chan returns eventqueue's read channel.
func (q *eventqueue[T]) Chan() <-chan T {
	return q.outchan
}

// Close closes the eventqueue. It purges all values still pending in
// the eventqueue and closes the eventqueue's read channel.
func (q *eventqueue[T]) Close() {
	// Terminate goroutine
	q.lock.Lock()
	q.buf = q.buf[:0]
	close(q.closechan)
	q.lock.Unlock()
	q.closewait.Wait()

	// Not it is safe to close output channel
	close(q.outchan)
}

// proc runs in goroutine and copies items from the buffer into the
// eventqueue's read channel.
func (q *eventqueue[T]) proc() {
	defer q.closewait.Done()

	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.buf) > 0 {
		v := q.buf[0]
		copy(q.buf, q.buf[1:])
		q.buf = q.buf[:len(q.buf)-1]

		q.lock.Unlock()
		select {
		case <-q.closechan:
		case q.outchan <- v:
		}
		q.lock.Lock()
	}
}
