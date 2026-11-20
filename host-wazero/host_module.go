package wazero_bridge

import (
	"context"
	"log"
	"sync"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"

	"github.com/pantopic/ext-cluster/host-wazero/pb"
	"github.com/pantopic/ext-util/host-wazero"
	"github.com/pantopic/wazero-pool"
)

// Name is the name of this host module.
const Name = "pantopic/ext-bridge"

var (
	ctxKeyMeta = Name + `/meta`
)

type meta struct {
	ptrTerm   uint32
	ptrBufCap uint32
	ptrBufLen uint32
	ptrBuf    uint32
}

type Manager interface {
	Start(ctx context.Context, term uint64, shard *pb.Shard) (err error)
	Stop(ctx context.Context, term uint64, shard *pb.Shard) (err error)
	Reload(ctx context.Context, term uint64, shard *pb.Shard) (err error)
}

type hostModule struct {
	sync.RWMutex
}

func New(opts ...Option) *hostModule {
	h := &hostModule{}
	for _, opt := range opts {
		opt(h)
	}
	return h
}

func (h *hostModule) Name() string {
	return Name
}

// InitContext retrieves the meta page from the wasm module
func (h *hostModule) InitContext(ctx context.Context, m api.Module) (context.Context, error) {
	fn := m.ExportedFunction(`__bridge`)
	if fn == nil {
		return ctx, nil
	}
	stack, err := fn.Call(ctx)
	if err != nil {
		return ctx, err
	}
	meta := &meta{}
	ptr := uint32(stack[0])
	for i, v := range []*uint32{
		&meta.ptrTerm,
		&meta.ptrBufCap,
		&meta.ptrBufLen,
		&meta.ptrBuf,
	} {
		*v = wazero_util.ReadUint32(m, ptr+uint32(4*i))
	}
	ctx = context.WithValue(ctx, ctxKeyMeta, meta)
	return ctx, nil
}

// ContextCopy populates dst context with the meta page from src context.
func (h *hostModule) ContextCopy(dst, src context.Context) context.Context {
	if v := src.Value(ctxKeyMeta); v != nil {
		dst = context.WithValue(dst, ctxKeyMeta, v.(*meta))
	}
	return dst
}

// Register instantiates the host module, making it available to all module instances in this runtime
func (h *hostModule) Register(ctx context.Context, r wazero.Runtime) (err error) {
	return
}

func (h *hostModule) Start(ctx context.Context, term uint64, shard *pb.Shard) (err error) {
	meta := get[*meta](ctx, ctxKeyMeta)
	wazeropool.FromContext(ctx).Run(func(mod api.Module) {
		setTerm(mod, meta, term)
		n, err := shard.MarshalToVT(dataBuf(mod, meta))
		if err != nil {
			log.Fatal(err)
		}
		wazero_util.WriteUint32(mod, meta.ptrBufLen, uint32(n))
		if _, err = mod.ExportedFunction("__bridge_start").Call(ctx); err != nil {
			log.Printf("Error calling __bridge_start: %v", err)
			return
		}
	})
	return
}

func (h *hostModule) Stop(ctx context.Context, term uint64, shard *pb.Shard) (err error) {
	meta := get[*meta](ctx, ctxKeyMeta)
	wazeropool.FromContext(ctx).Run(func(mod api.Module) {
		setTerm(mod, meta, term)
		n, err := shard.MarshalToVT(dataBuf(mod, meta))
		if err != nil {
			log.Fatal(err)
		}
		wazero_util.WriteUint32(mod, meta.ptrBufLen, uint32(n))
		if _, err = mod.ExportedFunction("__bridge_stop").Call(ctx); err != nil {
			log.Printf("Error calling __bridge_stop: %v", err)
			return
		}
	})
	return
}

func (h *hostModule) Reload(ctx context.Context, term uint64, shard *pb.Shard) (err error) {
	meta := get[*meta](ctx, ctxKeyMeta)
	wazeropool.FromContext(ctx).Run(func(mod api.Module) {
		setTerm(mod, meta, term)
		n, err := shard.MarshalToVT(dataBuf(mod, meta))
		if err != nil {
			log.Fatal(err)
		}
		wazero_util.WriteUint32(mod, meta.ptrBufLen, uint32(n))
		if _, err = mod.ExportedFunction("__bridge_reload").Call(ctx); err != nil {
			log.Printf("Error calling __bridge_reload: %v", err)
			return
		}
	})
	return
}

func get[T any](ctx context.Context, key string) T {
	v := ctx.Value(key)
	if v == nil {
		log.Panicf("Context item missing %s", key)
	}
	return v.(T)
}

func getData(m api.Module, meta *meta) []byte {
	return wazero_util.Read(m, meta.ptrBuf, meta.ptrBufLen, meta.ptrBufCap)
}

func getDataCopy(m api.Module, meta *meta) []byte {
	return append([]byte(nil), getData(m, meta)...)
}

func dataBuf(m api.Module, meta *meta) []byte {
	return wazero_util.Buffer(m, meta.ptrBuf, meta.ptrBufCap)
}

func setTerm(m api.Module, meta *meta, val uint64) {
	wazero_util.WriteUint64(m, meta.ptrTerm, val)
}

func getTerm(m api.Module, meta *meta) (val uint64) {
	return wazero_util.ReadUint64(m, meta.ptrTerm)
}

func setData(m api.Module, meta *meta, b []byte) {
	buf := dataBuf(m, meta)
	copy(buf[:cap(buf)], b)
	wazero_util.WriteUint32(m, meta.ptrBufLen, uint32(len(b)))
}
