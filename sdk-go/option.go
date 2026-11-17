package bridge

type Option func()

func WithBufferCap(n uint32) Option {
	return func() {
		bufCap = n
	}
}
