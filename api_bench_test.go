package json

import (
	"reflect"
	"testing"
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
