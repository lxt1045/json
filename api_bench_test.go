package json

import (
	"reflect"
	"testing"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

func BenchmarkLoadTagNode(b *testing.B) {
	in := &J0{}
	// 解引用； TODO: 用 Value 的方式提高性能
	vi := reflect.Indirect(reflect.ValueOf(in))
	if !vi.CanSet() {
		b.Fatal("err")
	}
	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, err := LoadTagNode(vi, goType.Hash)
	if err != nil {
		return
	}
	_ = tag
	b.Run("LoadTagNode", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tag, err = LoadTagNode(vi, goType.Hash)
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("LoadTagNodeSlow", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			tag, err = LoadTagNodeSlow(vi.Type(), goType.Hash)
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
}

func BenchmarkIF(b *testing.B) {
	b.Run("IF 1", func(b *testing.B) {
		b.ReportAllocs()
		stream0 := []byte("true,false,true")
		stream := stream0
		result := false
		for i := 0; i < b.N; i++ {
			if stream[0] == 't' && stream[1] == 'r' && stream[2] == 'u' && stream[3] == 'e' {
				stream = stream[5:]
				result = true
			} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
				stream = stream0
				result = false
			} else {
				err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(string(stream)))
				panic(err)
			}
		}
		_ = result
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("IF 2", func(b *testing.B) {
		b.ReportAllocs()
		stream0 := []byte("true,false,true")
		iTrue := *(*uint64)(unsafe.Pointer(&([]byte("true")[0]))) & 0x00000000ffffffff
		iFalse := *(*uint64)(unsafe.Pointer(&([]byte("false")[0]))) & 0x000000ffffffffff
		stream := stream0
		result := false
		for i := 0; i < b.N; i++ {
			if x := *(*uint64)(unsafe.Pointer(&stream[0])) & 0x00000000ffffffff; x == iTrue {
				stream = stream[5:]
				result = true
			} else if y := *(*uint64)(unsafe.Pointer(&stream[0])) & 0x000000ffffffffff; y == iFalse {
				stream = stream0
				result = false
			} else {
				err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(string(stream)))
				panic(err)
			}
		}
		_ = result
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
}
