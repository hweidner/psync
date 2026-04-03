# infchan - Channels with infinite capacity in Go

This Go package provides a Golang channel with infinite capacity
for values of a generic type.

The function InfChan returns a pair of _send_ and _receive_ channels.
It is possible to send data of a generic type, which can later be received.
In contrast to a builtin  Go channel, an unlimited number of values can
be sent without blocking.
Besides that, sending, receiving and closing the channel works exactly like
Go channels.

Internally, the values are stored in an underlying slice,
which will be resized as needed.

See the full package documentation on
https://pkg.go.dev/github.com/hweidner/psync/infchan
for details.

## License

This package is part of the [psync](https://github.com/hweidner/psync)
tool, licened under the GNU General Public License, version 3.
