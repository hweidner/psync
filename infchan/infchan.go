package infchan

/*
Function InfType implements a generic Go channel of infinite capacity.

Parameters are

  - the capacity _chanCap_ of the underlying channels,
  - the initial capacity _sliceCap_ of the underlying slice, and
  - a boolean flag _fifo_ which denotes ordering.

The function returns a pair of channels _send_ and _receive_.

The _send_ channel can then be used to send values to the InfChan in a
non-blocking fashion.
These values can be received from the _receive_ channel.
The sender must close the send channel after sending the last value.
The receiver can then receive the remaining values. After that, the
receive channel gets closed.

The internal data structures (a slice and a goroutine) will be destroyed
after the send channel is closed and the last value has been received.

When the _fifo_ flag is true, InfChan implements a strictly first-in-first-out
strategy. The values sent to the InfChan will be received in exactly the same order.
When the flag is false, the InfChan implements more or less a LIFO (last-in-first-out)
strategy, because this needs less allocations of the underlying slice. However,
both the capacity of the underlying channels and the concurrency of send/receive
operations affect the ordering. In practical applications, the order is quiet random
and indeterministic.

If the _chanCap_ parameter is 0, the _send_ and _receive_ channels are unbuffered.
This leads to much more context switches, especially with GOMAXPROCS=1 or on single
core machines. Sensible values for _chanCap_ are between 10 and 100.
*/
func InfChan[T any](chanCap, sliceCap uint, fifo bool) (send chan<- T, receive <-chan T) {
	in := make(chan T, chanCap)
	out := make(chan T, chanCap)
	data := make([]T, 0, sliceCap)

	go infchan(in, out, data, fifo)
	return in, out
}

// Func infchan is the internal infinite channel function that runs in
// a goroutine.
func infchan[T any](in <-chan T, out chan<- T, data []T, fifo bool) {
	// reader loop: read new values until the send channel gets closed
	// by the consumer
	index := 0
readloop:
	for {
		if len(data) == 0 {
			// data is empty, we can only read
			t, ok := <-in
			if !ok {
				break readloop
			}
			data = append(data, t)
		}

		// data is now present, we can read and write

		// Determine index of data to write. In FIFO mode,
		// always use first slice element to keep ordering.
		// Otherwise, take last element to avoid allocations.
		if !fifo {
			index = len(data) - 1
		}

		select {
		case t, ok := <-in:
			if !ok {
				break readloop
			}
			data = append(data, t)
		case out <- data[index]:
			// Drop the written value out of the slice
			if fifo {
				data = data[1:]
			} else {
				data = data[:index]
			}
		}
	}

	// writer loop: write the remaining values to the receiver.
	// This is always in FIFO order as there are no further
	// allocations needed.
	for _, t := range data {
		out <- t
	}
	close(out)
}
