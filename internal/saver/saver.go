package saver

import (
	"context"
	"sync"
	"time"

	"ova-method-api/internal/flusher"
	"ova-method-api/internal/model"
)

type Saver interface {
	Save(item model.Method)
	Close()
}

type saver struct {
	sync.Mutex

	retry   uint
	delay   time.Duration
	buffer  []model.Method
	flusher flusher.Flusher
}

func New(ctx context.Context, capacity, retry, delay uint, flusher flusher.Flusher) Saver {
	s := &saver{
		retry:   retry,
		flusher: flusher,
		delay:   time.Duration(delay) * time.Second,
		buffer:  make([]model.Method, 0, capacity),
	}
	go s.runAutoFlush(ctx)

	return s
}

func (s *saver) runAutoFlush(ctx context.Context) {
	withLock := func(fn func()) {
		s.Lock()
		defer s.Unlock()
		fn()
	}

	for {
		select {
		case <-time.After(s.delay):
			withLock(s.flushBuffer)
		case <-ctx.Done():
			withLock(s.flushBuffer)
			return
		}
	}
}

func (s *saver) flushBuffer() {
	for i := s.retry; i > 0; i-- {
		if s.flushAndClear() {
			return
		}
	}

	panic("failed flush buffer")
}

func (s *saver) flushAndClear() bool {
	bufLen := len(s.buffer)
	if bufLen == 0 {
		return true
	}

	unsaved := s.flusher.Flush(s.buffer)

	s.buffer = s.buffer[:0]
	s.buffer = append(s.buffer, unsaved...)

	return bufLen != len(s.buffer)
}

func (s *saver) Close() {
	s.Lock()
	defer s.Unlock()

	s.flushBuffer()
}

func (s *saver) Save(item model.Method) {
	s.Lock()
	defer s.Unlock()

	if !s.canAddToBuffer() {
		s.flushBuffer()
	}

	s.buffer = append(s.buffer, item)
}

func (s *saver) canAddToBuffer() bool {
	return cap(s.buffer) >= len(s.buffer)+1
}
