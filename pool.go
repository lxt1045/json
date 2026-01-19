// MIT License
//
// Copyright (c) 2021 Xiantu Li
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package json

import (
	"reflect"
	"sync"
	"sync/atomic"
	"unsafe"
)

var (
	strEface    = UnpackEface("")
	isliceEface = UnpackEface([]interface{}{})
	hmapimp     = func() *hmap {
		m := make(map[string]interface{})
		return *(**hmap)(unsafe.Pointer(&m))
	}()
	mapGoType = func() *maptype {
		m := make(map[string]interface{})
		typ := reflect.TypeOf(m)
		return (*maptype)(unsafe.Pointer(UnpackType(typ)))
	}()

	chbsPool = func() (ch chan *[]byte) {
		ch = make(chan *[]byte, 4)
		go newBytes(ch)
		return
	}()
)

func newBytes(ch chan *[]byte) {
	for {
		s := make([]byte, 0, bsPoolN)
		ch <- &s
	}
}
func newMapArray(ch chan *[]byte) {
	N := 1 << 20
	size := int(mapGoType.bucket.Size)
	N = N / size
	cap := N * size
	for {
		p := unsafe_NewArray(mapGoType.bucket, N)
		s := &SliceHeader{
			Data: p,
			Len:  cap,
			Cap:  cap,
		}
		ch <- (*[]byte)(unsafe.Pointer(s))
	}
}

var (
	poolSliceInterface = sync.Pool{New: func() any {
		return make([]interface{}, 1024)
	}}

	pairPool = sync.Pool{
		New: func() any {
			s := make([]pair, 0, 128)
			return &s
		},
	}
	bsPool = sync.Pool{New: func() any {
		return <-chbsPool
		// s := make([]byte, 0, bsPoolN)
		// return &s
	}}
	poolMapArrayInterface = func() sync.Pool {
		ch := make(chan *[]byte, 4) // runtime.GOMAXPROCS(0))
		// go func() {
		// 	for {
		// 		N := 1 << 20
		// 		p := unsafe_NewArray(mapGoType.bucket, N)
		// 		s := &SliceHeader{
		// 			Data: p,
		// 			Len:  N * int(mapGoType.bucket.Size),
		// 			Cap:  N * int(mapGoType.bucket.Size),
		// 		}
		// 		ch <- (*[]byte)(unsafe.Pointer(s))
		// 	}
		// }()
		go newMapArray(ch)
		return sync.Pool{New: func() any {
			return <-ch
		}}
	}()

	cacheStructTagInfo = NewRCU[uint32, *TagInfo]()
	strPool            = NewBatch[string]()
	islicePool         = NewBatch[[]interface{}]()
	imapPool           = NewBatch[hmap]()
)

const (
	bsPoolN = 1 << 20
	batchN  = 1 << 12
)

type pair struct {
	k string
	v interface{}
}

// 获取 string 的起始地址
func strToUintptr(p string) uintptr {
	return *(*uintptr)(unsafe.Pointer(&p))
}

func LoadTagNode(v reflect.Value, hash uint32) (*TagInfo, error) {
	tag, ok := cacheStructTagInfo.Get(hash)
	if ok {
		return tag, nil
	}
	return LoadTagNodeSlow(v.Type(), hash)
}

func LoadTagNodeByType(typ reflect.Type, hash uint32) (*TagInfo, error) {
	tag, ok := cacheStructTagInfo.Get(hash)
	if ok {
		return tag, nil
	}
	return LoadTagNodeSlow(typ, hash)
}

func LoadTagNodeSlow(typ reflect.Type, hash uint32) (*TagInfo, error) {
	ti, err := NewStructTagInfo(typ, nil)
	if err != nil {
		return nil, err
	}

	cacheStructTagInfo.Set(hash, ti)
	return ti, nil
}

// RCU 依据 Read Copy Update 原理实现
type RCU[T uintptr | uint32 | string | int, V any] struct {
	m unsafe.Pointer
}

func NewRCU[T uintptr | uint32 | string | int, V any]() (c RCU[T, V]) {
	m := make(map[T]V, 1)
	c.m = unsafe.Pointer(&m)
	return
}

func (c *RCU[T, V]) Get(key T) (v V, ok bool) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	v, ok = m[key]
	return
}

func (c *RCU[T, V]) Set(key T, v V) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	if _, ok := m[key]; ok {
		return
	}
	m2 := make(map[T]V, len(m)+10)
	m2[key] = v
	for {
		p := atomic.LoadPointer(&c.m)
		m = *(*map[T]V)(p)
		if _, ok := m[key]; ok {
			return
		}
		for k, v := range m {
			m2[k] = v
		}
		swapped := atomic.CompareAndSwapPointer(&c.m, p, unsafe.Pointer(&m2))
		if swapped {
			break
		}
	}
}

func (c *RCU[T, V]) GetOrSet(key T, load func() (v V)) (v V) {
	m := *(*map[T]V)(atomic.LoadPointer(&c.m))
	v, ok := m[key]
	if !ok {
		v = load()
		m2 := make(map[T]V, len(m)+10)
		m2[key] = v
		for {
			p := atomic.LoadPointer(&c.m)
			m = *(*map[T]V)(p)
			for k, v := range m {
				m2[k] = v
			}
			swapped := atomic.CompareAndSwapPointer(&c.m, p, unsafe.Pointer(&m2))
			if swapped {
				break
			}
		}
	}
	return
}

/*
Pool 和 store.Pool 一起，一次 Unmashal 调用，补充一次 pool 填满，此次执行期间，不会其他进程争抢；
结束后再归还，剩下的下次还可以继续使用
*/
type sliceNode[T any] struct {
	s   []T
	idx uint32 // atomic
}
type Batch[T any] struct {
	pool   unsafe.Pointer // *sliceNode[T]
	backup *sliceNode[T]
	sync.Mutex
}

func NewBatch[T any]() *Batch[T] {
	ret := &Batch[T]{
		pool: unsafe.Pointer(&sliceNode[T]{
			s:   make([]T, batchN),
			idx: 0,
		}),
		backup: &sliceNode[T]{
			s:   make([]T, batchN),
			idx: 0,
		},
	}
	return ret
}

func BatchGet[T any](b *Batch[T]) *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		return &sn.s[idx-1]
	}
	return b.Make()
}

func (b *Batch[T]) Get() *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		return &sn.s[idx-1]
	}
	return b.Make()
}

func (b *Batch[T]) GetN(n int) *T {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, uint32(n))
	var p *T
	if int(idx) > len(sn.s) {
		p = b.MakeN(n)
	} else {
		p = &sn.s[int(idx)-n]
	}
	return p
}

func (b *Batch[T]) GetN2(n int) (s []T) {
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, uint32(n))
	var p *T
	if int(idx) > len(sn.s) {
		p = b.MakeN(n)
	} else {
		p = &sn.s[int(idx)-n]
	}

	sh := (*SliceHeader)(unsafe.Pointer(&s))
	sh.Data = unsafe.Pointer(p)
	sh.Cap = n
	return

}

func (b *Batch[T]) Make() (p *T) {
	b.Lock()
	defer b.Unlock()

	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}

	if b.backup != nil {
		sn = b.backup
		sn.idx = 1
		b.backup = nil
	} else {
		sn = &sliceNode[T]{
			s:   make([]T, batchN),
			idx: 1,
		}
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	b.backup = &sliceNode[T]{
		s:   make([]T, batchN),
		idx: 1,
	}
	p = &sn.s[0]
	return
}

func (b *Batch[T]) MakeN(n int) (p *T) {
	if n > batchN {
		strs := make([]T, n)
		return &strs[0]
	}
	b.Lock()
	defer b.Unlock()
	sn := (*sliceNode[T])(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, 1)
	if int(idx) <= len(sn.s) {
		p = &sn.s[idx-1]
		return
	}
	if b.backup != nil {
		sn = b.backup
		sn.idx = uint32(n)
		b.backup = nil
	} else {
		sn = &sliceNode[T]{
			s:   make([]T, batchN),
			idx: uint32(n),
		}
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	b.backup = &sliceNode[T]{
		s:   make([]T, batchN),
		idx: 0,
	}
	p = &sn.s[0]
	return
}

type sliceObj struct {
	p   unsafe.Pointer
	idx uint32 // atomic
	end uint32 // atomic
}

// BatchObj 通过批量创建的方式减少单个创建平均延时
type BatchObj struct {
	pool   unsafe.Pointer // *sliceObj[T]
	backup *sliceObj
	goType *GoType
	size   uint32
	sync.Mutex
}

func NewBatchObj(typ reflect.Type) *BatchObj {
	if typ == nil {
		return &BatchObj{
			pool: unsafe.Pointer(&sliceObj{}),
		}
	}
	goType := UnpackType(typ)
	ret := &BatchObj{
		pool:   unsafe.Pointer(&sliceObj{}),
		goType: goType,
		size:   uint32(goType.Size),
	}
	return ret
}

func (b *BatchObj) Get() unsafe.Pointer {
	sn := (*sliceObj)(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, b.size)
	if idx <= sn.end {
		return pointerOffset(sn.p, uintptr(idx-b.size))
	}
	return b.MakeN(1)
}

func (b *BatchObj) Make() unsafe.Pointer {
	b.Lock()
	defer b.Unlock()

	sn := (*sliceObj)(atomic.LoadPointer(&b.pool))
	idx := atomic.AddUint32(&sn.idx, b.size)
	if idx <= sn.end {
		return pointerOffset(sn.p, uintptr(idx-b.size))
	}

	const N = 1 << 10
	sn = &sliceObj{
		p:   unsafe_NewArray(b.goType, N),
		idx: uint32(b.goType.Size),
		end: uint32(N) * uint32(b.goType.Size),
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	return sn.p
}

func (b *BatchObj) GetN(n int) unsafe.Pointer {
	sn := (*sliceObj)(atomic.LoadPointer(&b.pool))
	offset := uint32(n) * b.size
	idx := atomic.AddUint32(&sn.idx, offset)
	if idx <= sn.end {
		return pointerOffset(sn.p, uintptr(idx-offset))
	}
	return b.MakeN(n)

}

func (b *BatchObj) MakeN(n int) (p unsafe.Pointer) {
	if n > batchN {
		return unsafe_NewArray(b.goType, n)
	}
	b.Lock()
	defer b.Unlock()
	sn := (*sliceObj)(atomic.LoadPointer(&b.pool))
	offset := uint32(n) * b.size
	idx := atomic.AddUint32(&sn.idx, offset)
	if idx <= sn.end {
		return pointerOffset(sn.p, uintptr(idx-offset))
	}
	const N = 1 << 12
	sn = &sliceObj{
		p:   unsafe_NewArray(b.goType, N),
		idx: offset,
		end: uint32(N) * uint32(b.goType.Size),
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))
	return sn.p
}

func (b *BatchObj) MakeN2(n int) (p unsafe.Pointer) {
	if n > batchN {
		return unsafe_NewArray(b.goType, n)
	}
	b.Lock()
	defer b.Unlock()
	sn := (*sliceObj)(atomic.LoadPointer(&b.pool))
	offset := uint32(n) * b.size
	idx := atomic.AddUint32(&sn.idx, offset)
	if idx <= sn.end {
		return pointerOffset(sn.p, uintptr(idx-offset))
	}
	const N = 1 << 12

	if b.backup != nil {
		sn = b.backup
		sn.idx = offset
		b.backup = nil
	} else {
		sn = &sliceObj{
			p:   unsafe_NewArray(b.goType, N),
			idx: offset,
			end: uint32(N) * uint32(b.goType.Size),
		}
	}
	atomic.StorePointer(&b.pool, unsafe.Pointer(sn))

	b.backup = &sliceObj{
		p:   unsafe_NewArray(b.goType, N),
		idx: 0,
		end: uint32(N) * uint32(b.goType.Size),
	}

	return sn.p
}

type Store struct {
	obj unsafe.Pointer
	tag *TagInfo
}
type PoolStore struct {
	obj         unsafe.Pointer
	tag         *TagInfo
	pointerPool unsafe.Pointer // 也放 tag？放这里能省多少性能？
	slicePool   unsafe.Pointer // 这 slicePool 放 tag，dynPool 放全局，是不是可以减少 copy ？ 顺便较少函数调用时栈的开销？
}

var (
	gStringPool    = NewBatch[string]()      // string 类型太常见了
	gNoscanPool    = NewBatch[byte]()        // 不含指针的类型都可以用这个分配
	gIntPool       = NewBatch[int]()         //
	gInterfacePool = NewBatch[interface{}]() //
)

func (ps PoolStore) Idx(idx uintptr) (p unsafe.Pointer) {
	p = pointerOffset(ps.pointerPool, idx)
	*(*unsafe.Pointer)(ps.obj) = p
	return
}

func GrowBytes(in []byte, need int) []byte {
	l := need + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 8*1024 {
		l = 8 * 1024
	} else {
		l *= 16
	}
	out := make([]byte, 0, l)
	out = append(out, in...)
	return out[:l]
}
func GrowStrings(in []string, need int) []string {
	l := need + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 1024 {
		l = 1024
	} else {
		l *= 16
	}
	out := make([]string, 0, l)
	out = append(out, in...)
	return out[:need+len(in)]
}

func GrowInts(in []int) []int {
	l := 1 + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 1024 {
		l = 1024
	} else {
		l *= 16
	}
	out := make([]int, 0, l)
	out = append(out, in...)
	return out[:1+len(in)]
}

func (ps PoolStore) GetObjs(goType *GoType) []uint8 {
	pObj := ps.slicePool
	p := (*[]uint8)(pObj)
	pool := *p
	*p = nil
	if cap(pool)-len(pool) > 0 {
		return pool
	}
	l := 1024
	parr := unsafe_NewArray(goType, l)
	l = l * int(goType.Size)
	// s := SliceHeader{
	// 	Data: parr,
	// 	Len:  0,
	// 	Cap:  l,
	// }
	// return *(*[]uint8)(unsafe.Pointer(&s))
	out := (*(*[1 << 30]uint8)(parr))[:0:l]
	return out
}

func (ps PoolStore) SetObjs(in []uint8) PoolStore {
	p := (*[]uint8)(ps.slicePool)
	*p = in
	return ps
}

func GrowObjs(in []uint8, need int, goType *GoType) []uint8 {
	l := need + len(in)
	if l <= cap(in) {
		return in[:l]
	}
	if l < 1024 {
		l = 1024
	} else {
		l *= 16
	}
	parr := unsafe_NewArray(goType, l)
	l = l * int(goType.Size)
	out := (*(*[1 << 30]uint8)(parr))[:0:l]

	// s := SliceHeader{
	// 	Data: parr,
	// 	Len:  0,
	// 	Cap:  l,
	// }
	// out := *(*[]uint8)(unsafe.Pointer(&s))

	out = append(out, in...)
	return out[:need+len(in)]
}
