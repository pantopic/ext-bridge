package bridge

import (
	"unsafe"

	"github.com/pantopic/ext-cluster/sdk-go/pb"
)

type bridgeFunc func(term uint64, shard *pb.Shard)

var (
	meta   = make([]uint32, 4)
	term   uint64
	bufCap uint32 = 4 << 10 // 4 KiB
	bufLen uint32
	buf    []byte

	bridgeStart  bridgeFunc
	bridgeReload bridgeFunc
	bridgeStop   bridgeFunc

	shard = new(pb.Shard)
)

//export __bridge
func __bridge() uint32 {
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

//export __bridge_start
func __bridge_start() {
	bridgeStart(term, getShard())
}

//export __bridge_reload
func __bridge_reload() {
	bridgeReload(term, getShard())
}

//export __bridge_stop
func __bridge_stop() {
	bridgeStop(term, getShard())
}

func getShard() *pb.Shard {
	shard.Reset()
	if err := shard.UnmarshalVT(buf[:bufLen]); err != nil {
		panic(err)
	}
	return shard
}

var _ = __bridge
var _ = __bridge_start
var _ = __bridge_stop
var _ = __bridge_reload
