package relay

func Init(start, reload, stop relayFunc, opts ...Option) {
	relayStart = start
	relayReload = reload
	relayStop = stop
	for _, opt := range opts {
		opt()
	}
	buf = make([]byte, int(bufCap))
}
