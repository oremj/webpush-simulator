package simulator

type semaphore struct {
	ch chan struct{}
}

func newSemaphore(cnt int) *semaphore {
	return &semaphore{
		ch: make(chan struct{}, cnt),
	}
}

func (s *semaphore) Count() int {
	return len(s.ch)
}

func (s *semaphore) Acquire() {
	s.ch <- struct{}{}
}

func (s *semaphore) Release() {
	select {
	case <-s.ch:
	default:
	}
}
