package relay

import (
	"unsafe"

	"github.com/pantopic/ext-cluster/sdk-go/pb"
)

type relayFunc func(term uint64, shard *pb.Shard)

var (
	meta   = make([]uint32, 4)
	term   uint64
	bufCap uint32 = 4 << 10 // 4 KiB
	bufLen uint32
	buf    []byte

	relayStart  relayFunc
	relayReload relayFunc
	relayStop   relayFunc

	shard = new(pb.Shard)
)

//export __relay
func __relay() uint32 {
	for i, p := range []unsafe.Pointer{
		unsafe.Pointer(&bufCap),
		unsafe.Pointer(&bufLen),
		unsafe.Pointer(&buf[0]),
		unsafe.Pointer(&term),
	} {
		meta[i] = uint32(uintptr(p))
	}
	return uint32(uintptr(unsafe.Pointer(&meta[0])))
}

//export __relay_start
func __relay_start() {
	relayStart(term, getShard())
}

//export __relay_reload
func __relay_reload() {
	relayReload(term, getShard())
}

//export __relay_stop
func __relay_stop() {
	relayStop(term, getShard())
}

func getShard() *pb.Shard {
	shard.Reset()
	if err := shard.UnmarshalVT(buf[:bufLen]); err != nil {
		panic(err)
	}
	return shard
}

var _ = __relay
var _ = __relay_start
var _ = __relay_stop
var _ = __relay_reload
