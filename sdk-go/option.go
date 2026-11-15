package relay

type Option func()

func WithBufferCap(n uint32) Option {
	return func() {
		bufCap = n
	}
}
