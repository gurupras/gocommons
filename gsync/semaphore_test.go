package gsync

import "testing"

func TestP(t *testing.T) {
	t.Parallel()

	sem := NewSem(1)

	sem.P()
}
