package saver

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	"ova-method-api/internal/flusher"
	"ova-method-api/internal/model"
)

var (
	ErrFlushBuffer = fmt.Errorf("failed flush buffer")
)

type Saver interface {
	Save(item model.Method) error
	Close() error
}

type saver struct {
	sync.Mutex

	retry uint
	delay time.Duration
	ctx   context.Context

	buffer  []model.Method
	flusher flusher.Flusher
}

func New(ctx context.Context, capacity, delay uint, flusher flusher.Flusher) Saver {
	s := &saver{
		retry:   2,
		ctx:     ctx,
		flusher: flusher,
		delay:   time.Duration(delay) * time.Second,
		buffer:  make([]model.Method, 0, capacity),
	}
	go s.runAutoFlush()

	return s
}

func (s *saver) runAutoFlush() {
	withLock := func(fn func() error) {
		s.Lock()
		defer s.Unlock()
		if err := fn(); err != nil {
			log.Error().Err(err)
		}
	}

	for {
		select {
		case <-time.After(s.delay):
			withLock(s.flushBuffer)
		case <-s.ctx.Done():
			withLock(s.flushBuffer)
			return
		}
	}
}

func (s *saver) flushBuffer() error {
	for i := s.retry; i > 0; i-- {
		if s.flushAndClear() {
			return nil
		}
	}

	return ErrFlushBuffer
}

func (s *saver) flushAndClear() bool {
	bufLen := len(s.buffer)
	if bufLen == 0 {
		return true
	}

	unsaved := s.flusher.Flush(s.ctx, s.buffer)

	s.buffer = s.buffer[:0]
	s.buffer = append(s.buffer, unsaved...)

	return bufLen != len(s.buffer)
}

func (s *saver) Close() error {
	s.Lock()
	defer s.Unlock()

	return s.flushBuffer()
}

func (s *saver) Save(item model.Method) error {
	s.Lock()
	defer s.Unlock()

	if !s.canAddToBuffer() {
		if err := s.flushBuffer(); err != nil {
			return err
		}
	}

	s.buffer = append(s.buffer, item)
	return nil
}

func (s *saver) canAddToBuffer() bool {
	return cap(s.buffer) >= len(s.buffer)+1
}
