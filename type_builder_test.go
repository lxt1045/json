package json

import (
	"fmt"
	"reflect"
	"runtime"
	"testing"
	"unsafe"
)

func Test_NewSlice(t *testing.T) {
	top := func(in string) (p *string) {
		return &in
	}
	type X struct {
		x, y *string
	}
	printXs := func(xs []X) {
		fmt.Printf("[ ")
		for _, x := range xs {
			fmt.Printf("X{x:%s,y:%s} ", *x.x, *x.y)
		}
		fmt.Println("]")
	}

	f := func() []X {
		typ := UnpackEface(X{}).Type

		tt := reflect.TypeOf(X{})
		typ = UnpackType(tt)
		p := unsafe_NewArray(typ, 1)

		sh := reflect.SliceHeader{
			Data: uintptr(p),
			Len:  0,
			Cap:  2,
		}
		s := *(*[]X)(unsafe.Pointer(&sh))
		return s
	}
	s := f()
	runtime.GC()
	sa := append(s, X{top("1"), top("2")})
	sa = append(sa, X{top("11"), top("22")})

	printXs(sa)
	printXs(s[:2])
}

func Test_CacheStruct(t *testing.T) {
	b := NewTypeBuilder().
		AddString("Name").
		AddInt64("Age")

	b.Build()
	p, i := b.PInterface()
	pp := UnpackEface(i).Value
	fmt.Printf("typ:%T,\nvalue:%+v,\n ponter1:%d,\n ponter1:%d\n", p, i, p, pp)
}

func Test_CacheStructAppendField(t *testing.T) {
	lazyOffsets := make([]uintptr, 0, 8)
	b := NewTypeBuilder()
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendString("Name", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendIntSlice("Slice", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendInt64("Age", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendBool("Bool", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendBool("Bool1", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	b = b.AppendInt64("Age2", &lazyOffsets[len(lazyOffsets)-1])

	b.Build()
	p, i := b.PInterface()
	pp := UnpackEface(i).Value
	fmt.Printf("typ:%T,\nvalue:%+v,\n ponter1:%d,\n ponter1:%d,\nlazyOffsets:%+v\n",
		p, i, p, pp, lazyOffsets)
}

var pTypeBuilderST unsafe.Pointer
var ifaceTest interface{}

func Benchmark_CacheStruct(b *testing.B) {
	type TypeBuilderST struct {
		Name string
		Age  int
	}
	TypeBuilderSTType := reflect.TypeOf(TypeBuilderST{})

	lazyOffsets := make([]uintptr, 0, 8)
	builder := NewTypeBuilder()
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendString("Name", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendIntSlice("Slice", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendInt64("Age", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendBool("Bool", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendBool("Bool1", &lazyOffsets[len(lazyOffsets)-1])
	lazyOffsets = append(lazyOffsets, 0)
	builder = builder.AppendInt64("Age2", &lazyOffsets[len(lazyOffsets)-1])

	builder.Build()
	runs := []struct {
		name string
		f    func()
	}{
		{"std.New",
			func() {
				pTypeBuilderST = unsafe.Pointer(new(TypeBuilderST))
			},
		},
		{"reflect.New",
			func() {
				v := reflect.New(TypeBuilderSTType)
				pTypeBuilderST = reflectValueToPointer(&v)
			},
		},
		{"reflect.Interface",
			func() {
				ifaceTest = reflect.New(TypeBuilderSTType).Interface()
			},
		},
		{"builder.New",
			func() {
				pTypeBuilderST = builder.New()
			},
		},
		{"builder.Interface",
			func() {
				ifaceTest = builder.Interface()
			},
		},
		{"builder.Build",
			func() {
				builder.Build()
			},
		},
		{"builder.Build2",
			func() {
				builder.Type = nil
				builder.Build()
			},
		},
	}

	for _, r := range runs[:] {
		b.Run(r.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r.f()
			}
		})
	}
}

/*
func Test_Map(t *testing.T) {
	m := make(map[string]interface{})
	var in interface{} = m
	eface := UnpackEface(in)
	typ := (*maptype)(unsafe.Pointer(eface.Type))
	var hint int = 16
	pHmap := *(**hmap)(unsafe.Pointer(&m))
	B := uint8(0)
	for overLoadFactor(hint, B) {
		B++
	}
	pHmap.B = B

	pHmap.buckets, _ = makeBucketArray(typ, pHmap.B, nil)

	t.Logf("%+v\n", *pHmap)
	t.Logf("len(m):%+v\n", len(m))
}


// go test -benchmem -run=^$ -bench ^Benchmark_makeMap$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c

func Benchmark_makeMap(b *testing.B) {
	b.Run("pool", func(b *testing.B) {
		m := make(map[string]interface{})
		var in interface{} = m
		eface := UnpackEface(in)
		typ := (*maptype)(unsafe.Pointer(eface.Type))
		var hint int = 16
		pHmap := *(**hmap)(unsafe.Pointer(&m))
		B := uint8(0)
		for overLoadFactor(hint, B) {
			B++
		}
		pHmap.B = B
		size := typ.typ.Size

		pHmap.buckets, _ = makeBucketArray(typ, pHmap.B, nil)

		poolM := sync.Pool{
			New: func() any {
				B := uint8(8)
				buckets, _ := makeBucketArray(typ, B, nil)
				nbuckets := bucketShift(B)
				h := reflect.SliceHeader{
					Data: uintptr(buckets),
					Len:  int(nbuckets * typ.typ.Size),
					Cap:  int(nbuckets * typ.typ.Size),
				}
				uints := *(*[]uint8)(unsafe.Pointer(&h))
				return uints
			},
		}

		nbuckets := bucketShift(pHmap.B)
		l := int(nbuckets * size)

		m = make(map[string]interface{})
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			pHmap := *(**hmap)(unsafe.Pointer(&m))
			pHmap.B = B

			uints := poolM.Get().([]uint8)
			if cap(uints) < l {
				uints = poolM.New().([]uint8)
			}
			p := uints[:l]
			if cap(uints) > 2*l {
				uints = uints[l:]
				poolM.Put(uints)
			}
			pHmap.buckets = unsafe.Pointer(&p)
		}
	})
	b.Run("make", func(b *testing.B) {
		var m map[string]interface{}
		_ = m
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			m = make(map[string]interface{}, 16)
		}
	})
}
/*/
