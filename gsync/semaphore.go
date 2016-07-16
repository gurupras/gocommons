package gsync

type empty struct{}
type Semaphore struct {
	channel chan empty
}

func NewSem(count int) *Semaphore {
	s := new(Semaphore)
	s.channel = make(chan empty, count)
	return s
}

func (s *Semaphore) P() {
	e := empty{}
	s.channel <- e
}

func (s *Semaphore) V() {
	<-s.channel
}
