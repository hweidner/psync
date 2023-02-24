package infchan

import (
	"testing"
)

// How many values to send
const COUNT = 25

func TestFIFOChan(t *testing.T) {
	s, r := InfChan[int](10, COUNT, true)

	go func() {
		for i := 0; i < COUNT; i++ {
			s <- i
		}
		close(s)
	}()

	count, sum := 0, 0

	previous, ascending := 0, true
	for i := range r {
		count++
		sum += i

		// test FIFO (ascending) order
		if i < previous {
			ascending = false
		}
		previous = i
	}

	if count != COUNT || sum != COUNT*(COUNT-1)/2 {
		t.Errorf("Wrong values received from InfChan, expected %d values with sum %d, got %d values with sum %d.\n",
			COUNT, COUNT*(COUNT-1)/2, count, sum)
	}
	if !ascending {
		t.Errorf("Values received from InfChan in wrong order.\n")
	}
}

func TestLIFOChan(t *testing.T) {
	s, r := InfChan[int](10, COUNT, false)

	go func() {
		for i := 0; i < COUNT; i++ {
			s <- i
		}
		close(s)
	}()

	count, sum := 0, 0

	for i := range r {
		count++
		sum += i
	}

	if count != COUNT || sum != COUNT*(COUNT-1)/2 {
		t.Errorf("Wrong values received from InfChan, expected %d values with sum %d, got %d values with sum %d.\n",
			COUNT, COUNT*(COUNT-1)/2, count, sum)
	}
}
