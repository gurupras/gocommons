package gsync

import "testing"

func TestP(t *testing.T) {
	sem := NewSem(1)

	sem.P()
}
