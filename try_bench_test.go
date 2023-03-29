package json

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"unsafe"
)

// TODO: slice

var pgStr *string

func BenchmarkBatch(b *testing.B) {
	const N = 100
	batch := NewBatch[string]()
	for i := 0; i < 2; i++ {
		runtime.GC()
		b.Run("batch-BatchGet", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = BatchGet(batch)
			}
		})
		runtime.GC()
		b.Run("batch", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = batch.Get()
			}
		})
		runtime.GC()
		b.Run("batch-GetN", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = batch.GetN(1)
			}
		})
		runtime.GC()
		b.Run("batch-GetN4", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = batch.GetN(4)
			}
		})

		bObj := NewBatchObj(reflect.TypeOf(""))
		runtime.GC()
		b.Run("BatchObj-Get", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = (*string)(bObj.Get())
			}
		})
		runtime.GC()
		b.Run("BatchObj-GetN", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = (*string)(bObj.GetN(1))
			}
		})
		runtime.GC()
		b.Run("BatchObj-GetN4", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = (*string)(bObj.GetN(4))
			}
		})
	}
	stringPool := sync.Pool{
		New: func() any {
			s := make([]string, 512, 512) //8*1024/16
			return &s
		},
	}
	runtime.GC()
	b.Run("pool", func(b *testing.B) {
		for i := 0; i < b.N*N; i++ {
			pstrs := stringPool.Get().(*[]string)
			pgStr = &(*pstrs)[0]
			if len(*pstrs) > 1 {
				*pstrs = (*pstrs)[1:]
				stringPool.Put(pstrs)
			}
		}
	})
	runtime.GC()
	b.Run("new", func(b *testing.B) {
		for i := 0; i < b.N*N; i++ {
			pgStr = new(string)
		}
	})
	l := sync.Mutex{}
	runtime.GC()
	b.Run("Lock", func(b *testing.B) {
		for i := 0; i < b.N*N; i++ {
			l.Lock()
			l.Unlock()
		}
	})
	runtime.GC()
	b.Run("reflect.StringHeader", func(b *testing.B) {
		for i := 0; i < b.N*N; i++ {
			h := reflect.StringHeader{}
			pgStr = (*string)((unsafe.Pointer)(&h))
		}
	})
	runtime.GC()
	b.Run("lxt.StringHeader", func(b *testing.B) {
		for i := 0; i < b.N*N; i++ {
			h := StringHeader{}
			pgStr = (*string)((unsafe.Pointer)(&h))
		}
	})
	runtime.GC()
}

func BenchmarkNewGC(b *testing.B) {
	const N = 100
	type StringGC struct {
		Data *byte
		Len  int
	}
	type String struct {
		Data int
		Len  int
	}
	batch := NewBatch[string]()
	for i := 0; i < 5; i++ {
		runtime.GC()
		b.Run("batch-BatchGet", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				pgStr = BatchGet(batch)
			}
		})
		runtime.GC()
		b.Run("reflect.String", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				h := String{}
				pgStr = (*string)((unsafe.Pointer)(&h))
			}
		})
		runtime.GC()
		b.Run("lxt.StringGC", func(b *testing.B) {
			for i := 0; i < b.N*N; i++ {
				h := StringGC{}
				pgStr = (*string)((unsafe.Pointer)(&h))
			}
		})
		runtime.GC()
	}
}
