package bridge

func Init(start, reload, stop bridgeFunc, opts ...Option) {
	bridgeStart = start
	bridgeReload = reload
	bridgeStop = stop
	for _, opt := range opts {
		opt()
	}
	buf = make([]byte, int(bufCap))
}
