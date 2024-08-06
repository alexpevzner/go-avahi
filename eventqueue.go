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
	buf  []T
	chn  chan T
	lock sync.Mutex
}

// init initializes an eventqueue
func (q *eventqueue[T]) init() {
	q.buf = make([]T, 0, 8)
	q.chn = make(chan T)
}

// Push adds a new value to the eventqueue
func (q *eventqueue[T]) Push(v T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.buf = append(q.buf, v)
	if len(q.buf) == 1 {
		go q.proc()
	}
}

// Chan returns eventqueue's read channel.
func (q *eventqueue[T]) Chan() <-chan T {
	return q.chn
}

// Close closes the eventqueue. It purges all values still pending in
// the eventqueue and closes the eventqueue's read channel.
func (q *eventqueue[T]) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.buf = q.buf[:0]
	close(q.chn)
}

// proc runs in goroutine and copies items from the buffer into the
// eventqueue's read channel.
func (q *eventqueue[T]) proc() {
	q.lock.Lock()
	defer q.lock.Unlock()

	for len(q.buf) > 0 {
		v := q.buf[0]
		copy(q.buf, q.buf[1:])
		q.buf = q.buf[:len(q.buf)-1]

		q.lock.Unlock()
		q.chn <- v
		q.lock.Lock()
	}
}
