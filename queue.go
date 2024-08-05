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

// queue represents a queue of values of some type T.
//
// Values added to the queue using Push method and can
// be retrieved from the queue using a channel.
type queue[T any] struct {
	buf  []T
	chn  chan T
	lock sync.Mutex
}

// newQueue makes a new queue of type T.
func newQueue[T any]() *queue[T] {
	return &queue[T]{
		buf: make([]T, 0, 8),
		chn: make(chan T),
	}
}

// Push adds a new value to the queue
func (q *queue[T]) Push(v T) {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.buf = append(q.buf, v)
	if len(q.buf) == 1 {
		go q.proc()
	}
}

// Chan returns queue's read channel.
func (q *queue[T]) Chan() <-chan T {
	return q.chn
}

// Close closes the queue. It purges all values still pending in
// the queue and closes the queue's read channel.
func (q *queue[T]) Close() {
	q.lock.Lock()
	defer q.lock.Unlock()

	q.buf = q.buf[:0]
	close(q.chn)
}

// proc runs in goroutine and copies items from the buffer into the queue's
// read channel.
func (q *queue[T]) proc() {
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
