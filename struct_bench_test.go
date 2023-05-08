package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"unsafe"

	"github.com/bytedance/sonic"
	"github.com/lxt1045/json/testdata"
)

func Tt(value reflect.Value) {
	if value.Kind() != reflect.Ptr {
		panic("ss")
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		panic("ss")
	}
	field := value.FieldByName("SS")
	if !field.CanSet() {
		return
	}
	// str := "888"
	// field.Set(reflect.ValueOf(&str))
	field = field.Elem()
	if !field.CanSet() {
		return
	}
	field.SetString("999")
	// field.SetIterValue()

	m := value.FieldByName("M")
	if !m.CanSet() {
		panic("m")
	}
	iter := m.MapRange()
	if iter == nil {
		return
	}
	for iter.Next() {
		v := iter.Value().String()
		v += "-xxx"
		m.SetMapIndex(iter.Key(), reflect.ValueOf(v))
	}
}

func Test_String(t *testing.T) {

	runtime.GC()
	s := struct {
		SS *string
		M  map[string]string
	}{}
	x := ""
	s.SS = &x
	s.M = map[string]string{
		"a":  "b",
		"a1": "b11",
	}
	t.Run("1", func(t *testing.T) {
		Tt(reflect.ValueOf(&s))
		t.Logf("%+v", s)
		t.Logf("%+v", *s.SS)
	})
}

func Test_parseStr(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		src := `"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"`
		raw, i, n := parseStr(src, -1)
		t.Logf("%s, i:%d, n:%d", string(raw), i, n)
	})
}

// TODO:
//    1. SIMD 加速
//	  5. 全部 key 找出来之后，再排序，再从 bytes 中找出对应的 key?
//    11. bytes.IndexByte 和 map 文章

/*
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalPoniter$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalPoniter$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
*/
func BenchmarkMyUnmarshalPoniter(b *testing.B) {
	type Name struct {
		ZHCN  *string `json:"ZH_CN"`
		ZHCN1 *string `json:"ZH_CN1"`
		ZHCN2 *string `json:"xx"`
		ZHCN3 *string `json:"yy"`
		ZHCN4 *string `json:"DD"`
		ZHCN5 *string `json:"os"`
		ZHCN6 *string `json:"test"`
		ZHCN7 *string `json:"zxzv"`
		ZHCN8 *string `json:"XDS1"`
		ENUS  *string `json:"EN_US"`
	}
	type NameA struct {
		ZHCN  string `json:"ZH_CN"`
		ZHCN1 string `json:"ZH_CN1"`
		ZHCN2 string `json:"xx"`
		ZHCN3 string `json:"yy"`
		ZHCN4 string `json:"DD"`
		ZHCN5 string `json:"os"`
		ZHCN6 string `json:"test"`
		ZHCN7 string `json:"zxzv"`
		ZHCN8 string `json:"XDS1"`
		ENUS  string `json:"EN_US"`
	}
	bs := []byte(`{
		"ZH_CN":"chinesechinesec",
		"ZH_CN1":"chinesechinesec",
		"xx":"chinesechinesec",
		"yy":"chinesechinesec",
		"DD":"chinesechinesec",
		"os":"chinesechinesec",
		"test":"chinesechinesec",
		"zxzv":"chinesechinesec",
		"XDS1":"chinesechinesec",
		"EN_US":"English"
	}`)
	str := string(bs)

	d := Name{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	runtime.GC()
	sonic.UnmarshalString(str, &d)
	Unmarshal(bs, &d)

	//

	// return
	b.Run("Unmarshal-p", func(b *testing.B) {
		d := Name{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := Unmarshal(bs, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("sonic-p", func(b *testing.B) {
		d := Name{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("Unmarshal", func(b *testing.B) {
		d := NameA{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := Unmarshal(bs, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("sonic", func(b *testing.B) {
		d := NameA{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("Marshal-p", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := Marshal(&d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("sonic.Marshal-p", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := sonic.Marshal(&d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run("sonic.Marshal-p-string", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := sonic.MarshalString(&d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

func BenchmarkMyUnmarshal(b *testing.B) {
	name := "Unmarshal"
	bs := []byte(j0)
	d := J0{}
	err := Unmarshal(bs, &d)
	if err != nil {
		panic(err)
	}
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := Unmarshal(bs, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

func BenchmarkMyUnmarshal2(b *testing.B) {
	bs := []byte(j0)
	d := J0{}
	{
		err := Unmarshal(bs, &d)
		if err != nil {
			b.Fatal(err)
		}
	}

	name := "Unmarshal"
	b.Run(name, func(b *testing.B) {
		d := map[string]interface{}{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := UnmarshalString(j0, &d)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run(name, func(b *testing.B) {
		d := map[string]interface{}{}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(j0, &d)
			if err != nil {
				b.Fatal(err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

// sonic
func BenchmarkMyUnmarshal3(b *testing.B) {
	name := "Unmarshal"
	bs := []byte(j0)
	d := J0{}
	err := Unmarshal(bs, &d)
	if err != nil {
		panic(err)
	}
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(j0, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	b.Run(name, func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := UnmarshalString(j0, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

/*
go test -benchmem -run=^$  -benchtime=10000000x -bench ^BenchmarkUnmarshalStruct$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c

go test -benchmem -run=^$ -v -benchtime=10000000x -bench ^BenchmarkUnmarshalStruct$ github.com/lxt1045/json -count=1
BenchmarkUnMarshalStruct/lxt-st-12              10000000               129.2 ns/op      77375851.24 MB/s               0 B/op          0 allocs/op
BenchmarkUnMarshalStruct/sonic-st-12            10000000               155.8 ns/op      64166301.93 MB/s               0 B/op          0 allocs/op
BenchmarkUnMarshalStruct/lxt-st#01-12           10000000               127.5 ns/op      78409245.03 MB/s               0 B/op          0 allocs/op
BenchmarkUnMarshalStruct/sonic-st#01-12         10000000               148.5 ns/op      67361422.86 MB/s               0 B/op          0 allocs/op

BenchmarkUnmarshalStruct
BenchmarkUnmarshalStruct/lxt-st
BenchmarkUnmarshalStruct/lxt-st-12              10000000               137.1 ns/op      72926417.64 MB/s               0 B/op          0 allocs/op
BenchmarkUnmarshalStruct/sonic-st
BenchmarkUnmarshalStruct/sonic-st-12            10000000               156.1 ns/op      64064282.98 MB/s               0 B/op          0 allocs/op
BenchmarkUnmarshalStruct/lxt-st-string
BenchmarkUnmarshalStruct/lxt-st-string-12       10000000               137.0 ns/op      72995122.89 MB/s               0 B/op          0 allocs/op
BenchmarkUnmarshalStruct/lxt-st#01
BenchmarkUnmarshalStruct/lxt-st#01-12           10000000               137.7 ns/op      72632868.70 MB/s               0 B/op          0 allocs/op
BenchmarkUnmarshalStruct/sonic-st#01
BenchmarkUnmarshalStruct/sonic-st#01-12         10000000               153.9 ns/op      64996997.62 MB/s               0 B/op          0 allocs/op
BenchmarkUnmarshalStruct/lxt.marshal-st
BenchmarkUnmarshalStruct/lxt.marshal-st-12      10000000                77.56 ns/op     128934363.78 MB/s             56 B/op          0 allocs/op
BenchmarkUnmarshalStruct/sonic.marshal-st
BenchmarkUnmarshalStruct/sonic.marshal-st-12            10000000               193.1 ns/op      51773914.52 MB/s             133 B/op          4 allocs/op
*/
func BenchmarkUnmarshalStruct(b *testing.B) {
	type Name struct {
		ZHCN  string `json:"ZH_CN"`
		ENUS  string `json:"EN_US"`
		Count int    `json:"count"`
	}
	bs := []byte(`{
		"ZH_CN":"chinesechinesec",
		"EN_US":"English",
		"count":8
	}`)
	str := string(bs)
	var d Name
	Unmarshal(bs, &d)
	sonic.UnmarshalString(str, &d)

	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-st",
			func() {
				Unmarshal(bs, &d)
			},
		},
		{
			"sonic-st",
			func() {
				sonic.UnmarshalString(str, &d)
			},
		},
		{"lxt-st-string",
			func() {
				UnmarshalString(str, &d)
			},
		},
		{
			"sonic-st",
			func() {
				sonic.UnmarshalString(str, &d)
			},
		},
		{"lxt.marshal-st",
			func() {
				Marshal(&d)
			},
		},
		{
			"sonic.marshal-st",
			func() {
				sonic.Marshal(&d)
			},
		},
	}

	// b.RunParallel(func(pb *testing.PB) {
	// 	for pb.Next() {
	// 		Unmarshal(bs, &d)
	// 	}
	// })
	// b.RunParallel(func(pb *testing.PB) {
	// 	for pb.Next() {
	// 		sonic.UnmarshalString(str, &d)
	// 	}
	// })
	// return
	for _, r := range runs[:] {
		runtime.GC()
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.SetBytes(int64(b.N))
			b.StopTimer()
		})
	}
}

/*
go test -benchmem -run=^$ -bench ^BenchmarkUnMarshalStructMap$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkUnMarshalStructMap$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//
*/
func BenchmarkUnMarshalStructMap(b *testing.B) {
	bs := []byte(`{
		"ZH_CN1":"chinesechinese",
		"xx":"chinese",
		"yy":"chinese",
		"os":"chinesechinese",
		"test":"chinesechinese",
		"zxzv":"chinesechinese",
		"ZH_CN":"chinese",
		"EN_US":"English",
		"XDS1":"English"
	}`)
	type Name struct {
		ZHCN  *string `json:"ZH_CN"`
		ZHCN1 *string `json:"ZH_CN1"`
		ZHCN2 *string `json:"xx"`
		ZHCN3 *string `json:"yy"`
		ZHCN5 *string `json:"os"`
		ZHCN6 *string `json:"test"`
		ZHCN7 *string `json:"zxzv"`
		ZHCN8 *string `json:"XDS1"`
		ENUS  *string `json:"EN_US"`
	}
	st := Name{}
	str := string(bs)
	var d map[string]interface{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	err = sonic.UnmarshalString(str, &d)
	if err != nil {
		b.Fatal(err)
	}
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-map",
			func() {
				err := Unmarshal(bs, &d)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			"sonic-map",
			func() {
				err := sonic.UnmarshalString(str, &d)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{"lxt-map",
			func() {
				err := Unmarshal(bs, &d)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			"sonic-map",
			func() {
				err := sonic.UnmarshalString(str, &d)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{"lxt-st",
			func() {
				err := Unmarshal(bs, &st)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
	}

	for _, r := range runs[:] {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.SetBytes(int64(b.N))
			b.StopTimer()
		})
	}
}

/*
go test -benchmem -run=^$ -benchtime=1000000x -bench ^BenchmarkUnmarshalStruct1x$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalStruct1x$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
go test -benchmem -run=^$ -benchtime=10000000x -bench "^BenchmarkUnmarshalStruct1x$"
BenchmarkUnmarshalStruct1x/lxt-st-12             1000000              1221 ns/op             320 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/sonic-st-12           1000000              1571 ns/op             364 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/lxt-st#01-12          1000000              1202 ns/op             320 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/sonic-st#01-12        1000000              1569 ns/op             364 B/op          1 allocs/op

BenchmarkUnmarshalStruct1x/lxt-st-12             1000000              1161 ns/op             320 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/sonic-st-12           1000000              1593 ns/op             365 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/lxt-st#01-12          1000000              1171 ns/op             320 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x/sonic-st#01-12        1000000              1554 ns/op             359 B/op          1 allocs/op
*/
func BenchmarkUnmarshalStruct1x(b *testing.B) {
	bs := []byte(j0)
	data := string(bs)
	d := J0{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	_, err = json.Marshal(&d)
	if err != nil {
		b.Fatal(err)
	}

	runtime.GC()
	m := J0{}
	b.Run("lxt-Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := UnmarshalString(data, &m)
			if err != nil {
				panic(err)
			}
		}
	})
	// return
	runtime.GC()
	b.Run("lxt-Unmarshal", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(data, &m)
			if err != nil {
				panic(err)
			}
		}
	})
	// return
	runtime.GC()
	_ = fmt.Sprintf("d :%+v", d)
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-st",
			func() {
				m := J0{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				m := J0{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"lxt-st",
			func() {
				m := J0{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				m := J0{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
	}

	for _, r := range runs[:] {
		runtime.GC()
		b.Run(r.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				r.f()
			}
		})
	}
}

/*
go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalStruct1xMap$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalStruct1xMap$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//   */

func BenchmarkUnmarshalStruct1xMap(b *testing.B) {
	bs := []byte(j0)
	data := string(bs)
	d := J0{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	_, err = json.Marshal(&d)
	if err != nil {
		b.Fatal(err)
	}
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-map",
			func() {
				m := map[string]interface{}{}
				err := Unmarshal(bs, &m)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			"sonic-map",
			func() {
				m := map[string]interface{}{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{"lxt-map",
			func() {
				m := map[string]interface{}{}
				err := Unmarshal(bs, &m)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			"sonic-map",
			func() {
				m := map[string]interface{}{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{"lxt-st",
			func() {
				m := J0{}
				err := Unmarshal(bs, &m)
				if err != nil {
					b.Fatal(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				m := J0{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					b.Fatal(err)
				}
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
go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalStruct1x_medium $ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
*/
func BenchmarkUnmarshalStruct1x_small(b *testing.B) {
	bs := []byte(testdata.BookData)
	data := string(bs)
	d := testdata.Book{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	bsOut, err := json.Marshal(&d)
	if err != nil {
		b.Fatal(err)
	}
	m := testdata.Book{}
	sonic.UnmarshalString(data, &m)
	runtime.GC()
	_ = fmt.Sprintf("d :%+v", d)
	if string(bsOut) != string(testdata.BookDataOut) {
		str := string(bsOut)
		str2 := string(testdata.BookDataOut)
		for i := range str2 {
			if len(str) > i && str[i] != str2[i] {
				l := len(str2)
				if l-i > 8 {
					l = i + 8
				}
				if len(str) >= len(str2) {
					b.Logf("i:%d, c:%s,%s", i, str[i:l], str2[i:l])
				}
			}
		}
		b.Fatalf("len:%d,%d,bsOut:%s", len(str), len(str2), str)
	}
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-st",
			func() {
				// m := testdata.Book{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.Book{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"lxt-st",
			func() {
				// m := testdata.Book{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.Book{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std-st",
			func() {
				// m := testdata.Book{}
				err := json.Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic.marshal-st",
			func() {
				_, err := sonic.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"lxt.marshal-st",
			func() {
				_, err := Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std.marshal-st",
			func() {
				_, err := json.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
	}

	if false {
		runtime.GC()
		b.Run("RunParallel-lxt", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					UnmarshalString(data, &m)
				}
			})
		})
		runtime.GC()
		b.Run("RunParallel-sonic", func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					sonic.UnmarshalString(data, &m)
				}
			})
		})
	}

	for _, r := range runs[:] {
		runtime.GC()
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(data)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.f()
			}
		})
	}
}

func BenchmarkUnmarshalStruct1x_medium(b *testing.B) {
	bs := []byte(testdata.TwitterJson)
	data := string(bs)
	d := testdata.TwitterStruct{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	m := testdata.TwitterStruct{}
	err = sonic.UnmarshalString(data, &m)
	if err != nil {
		b.Fatal(err)
	}
	runtime.GC()
	_ = fmt.Sprintf("d :%+v", d)

	bsOut, err := json.Marshal(&d)
	if err != nil {
		b.Fatal(err)
	}
	if string(bsOut) != testdata.TwitterJsonOut {
		str := string(bsOut)
		for i := range testdata.TwitterJsonOut {
			if len(str) > i+8 && str[i] != testdata.TwitterJsonOut[i] {
				b.Logf("i:%d, c:%s,%s", i, str[i:i+8], testdata.TwitterJsonOut[i:i+8])
			}
		}
		b.Fatalf("len:%d,%d,bsOut:%s", len(str), len(testdata.TwitterJsonOut), str)
	}
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"lxt-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std-st",
			func() {
				// m := testdata.Book{}
				err := json.Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic.marshal-st",
			func() {
				_, err := sonic.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"lxt.marshal-st",
			func() {
				_, err := Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std.marshal-st",
			func() {
				_, err := json.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
	}

	for _, r := range runs[:] {
		runtime.GC()
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.f()
			}
		})
	}
}

func BenchmarkUnmarshalStruct1x_large(b *testing.B) {
	bs := []byte(testdata.TwitterJsonLarge)
	data := string(bs)
	d := testdata.TwitterStruct{}
	d2 := testdata.TwitterStruct{}
	err := json.Unmarshal(bs, &d2)
	if err != nil {
		b.Fatal(err)
	}

	m := testdata.TwitterStruct{}
	err = sonic.UnmarshalString(data, &m)
	if err != nil {
		panic(err)
	}

	// return
	err = Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	runtime.GC()
	_ = fmt.Sprintf("d :%+v", d)

	// bsOut, err := json.Marshal(&d)
	// if err != nil {
	// 	b.Fatal(err)
	// }
	// _ = bsOut
	// dGlobal = fmt.Sprintf("bsOut:%s", string(bsOut))
	// if string(bsOut) != testdata.TwitterJsonOut {
	// 	str := string(bsOut)
	// 	b.Fatalf("len:%d,%d,bsOut:%s", len(str), len(testdata.TwitterJsonOut), str)
	// 	for i := range str {
	// 		if str[i] != testdata.TwitterJsonOut[i] {
	// 			b.Logf("i:%d, c:%s,%s", i, str[i:i+8], testdata.TwitterJsonOut[i:i+8])
	// 		}
	// 	}
	// }
	runs := []struct {
		name string
		f    func()
	}{
		{"lxt-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"lxt-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := sonic.UnmarshalString(data, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std-st",
			func() {
				// m := testdata.TwitterStruct{}
				err := json.Unmarshal(bs, &m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"sonic.marshal-st",
			func() {
				_, err := sonic.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{
			"lxt.marshal-st",
			func() {
				_, err := Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
		{"std.marshal-st",
			func() {
				_, err := json.Marshal(&m)
				if err != nil {
					panic(err)
				}
			},
		},
	}

	for _, r := range runs {
		runtime.GC()
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			b.SetBytes(int64(len(bs)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				r.f()
			}
		})
	}
}

/*
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalLarge$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalLarge$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//
*/
func BenchmarkMyUnmarshalLarge(b *testing.B) {
	bs := []byte(testdata.TwitterJson)
	bs = []byte(testdata.TwitterJsonLarge)
	data := string(bs)
	_ = data
	d := testdata.TwitterStruct{}
	// return
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	b.Run("large", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			d := testdata.TwitterStruct{}
			// err := json.Unmarshal(bs, &d)
			// if err != nil {
			// 	b.Fatalf("[%d]:%v", i, err)
			// }
			err := Unmarshal(bs, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
			// 	err := sonic.UnmarshalString(data, &m)
			// 	if err != nil {
			// 		panic(err)
			// 	}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	return
	b.Run("spaceTable", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			for i := 0; i < len(bs); i++ {
				if spaceTable[bs[i]] {
					b0 = bs[i]
				}
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

/*
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalSmall$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkMyUnmarshalSmall$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//
*/
func BenchmarkMyUnmarshalSmall(b *testing.B) {
	bs := []byte(testdata.BookData)
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			m := testdata.Book{}
			err := Unmarshal(bs, &m)
			if err != nil {
				panic(err)
			}
		}
	})
}

/*
go test -benchmem -run=^$ -bench ^BenchmarkMarshalMedium $ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkMarshalLarge$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//
*/
func BenchmarkMarshalLarge(b *testing.B) {
	bs := []byte(testdata.TwitterJsonLarge)
	d := testdata.TwitterStruct{}
	err := json.Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	runtime.GC()
	b.Run("Large", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Marshal(&d)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
func BenchmarkMarshalMedium(b *testing.B) {
	bs := []byte(testdata.TwitterJson)
	d := testdata.TwitterStruct{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	runtime.GC()
	b.Run("Medium ", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Marshal(&d)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}
func BenchmarkMarshalSmall(b *testing.B) {
	var vBook = func() testdata.Book {
		bs := []byte(testdata.BookData)
		m := testdata.Book{}
		err := Unmarshal(bs, &m)
		if err != nil {
			panic(err)
		}
		return m
	}()
	b.Run("small", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, err := Marshal(&vBook)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

var pp unsafe.Pointer
var pm *map[string]interface{}
var m map[string]interface{}
var str *string
var pbs *[]byte
var bs []byte
var pbool *bool
var b0 byte
var iface *interface{}

func BenchmarkUnmarshalStruct20(b *testing.B) {
	bs := []byte(j0)
	data := string(bs)
	d := J0{}
	err := Unmarshal(bs, &d)
	if err != nil {
		b.Fatal(err)
	}
	bs, err = json.Marshal(&d)
	if err != nil {
		b.Fatal(err)
	}
	runs := []struct {
		name string
		f    func()
	}{
		{"std",
			func() {
				var m map[string]interface{}
				json.Unmarshal(bs, &m)
			},
		},
		{"std-st",
			func() {
				m := J0{}
				json.Unmarshal(bs, &m)
			},
		},
		{
			"sonic",
			func() {
				var m map[string]interface{}
				sonic.UnmarshalString(data, &m)
			},
		},
		{"lxt-st",
			func() {
				m := J0{}
				Unmarshal(bs, &m)
			},
		},
		{
			"sonic-st",
			func() {
				m := J0{}
				sonic.UnmarshalString(data, &m)
			},
		},
		{"lxt-st",
			func() {
				m := J0{}
				Unmarshal(bs, &m)
			},
		},
		{
			"sonic-st",
			func() {
				m := J0{}
				sonic.UnmarshalString(data, &m)
			},
		},
	}

	for _, r := range runs[3:] {
		b.Run(r.name, func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				r.f()
			}
			b.StopTimer()
			b.SetBytes(int64(b.N))
		})
	}
}

func Test_tagParse(t *testing.T) {
	t.Run("map", func(t *testing.T) {
		type BizName struct {
			ZHCN string `json:"ZH_CN"`
			ENUS string `json:"EN_US"`
		}
		type DataSt struct {
			ItemID   []int64   `json:"ItemID,string"`
			BizName  []BizName `json:"BizName"`
			BizCode  string    `json:"BizCode"`
			BizCode1 string
		}
		d := DataSt{}
		typ := reflect.TypeOf(&d)
		typ = typ.Elem()
		to, err := NewStructTagInfo(typ, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("to:%+v", to)
		return
	})
}

func TestMyUnmarshal(t *testing.T) {
	d := J0{}
	err := Unmarshal([]byte(j0), &d)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	bs, err := json.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("to:%s", string(bs))
}

var bGlobal = false

func BenchmarkIsSpace(b *testing.B) {
	bs := make([]byte, 10240)
	for i := range bs {
		bs[i] = byte(rand.Uint32())
	}
	const charSpace uint32 = 1<<('\t'-1) | 1<<('\n'-1) | 1<<('\v'-1) | 1<<('\f'-1) | 1<<('\r'-1) | 1<<(' '-1)

	b.Run("1", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b := bs[i%10240]
			bGlobal = b == 0x85 || b == 0xA0 || (charSpace>>(b-1)&0x1 > 0)
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b := bs[i%10240]
			bGlobal = b == 0x85 || b == 0xA0 || b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r' || b == ' '
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("3", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b := bs[i%10240]
			switch b {
			// toto: 用bitmap加速:
			case '\t', '\n', '\v', '\f', '\r', ' ', 0x85, 0xA0:
				bGlobal = true
			}
			bGlobal = false
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("4", func(b *testing.B) {
		table := [256]bool{}
		for i := range table {
			b := byte(i)
			if b == 0x85 || b == 0xA0 || b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r' || b == ' ' {
				table[i] = true
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b := bs[i%10240]
			bGlobal = table[b]
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("5", func(b *testing.B) {
		table := [4]uint64{}
		for i := range table {
			b := byte(i)
			if b == 0x85 || b == 0xA0 || b == '\t' || b == '\n' || b == '\v' || b == '\f' || b == '\r' || b == ' ' {
				idx := i / 64
				n := i % 64
				table[idx] |= 1 << n
			}
		}

		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			b := bs[i%10240]
			idx := b / 64
			n := b % 64
			bGlobal = table[idx]&(1<<n) > 0
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

}

func BenchmarkCron(b *testing.B) {
	bss := []string{
		":x",
		": x",
		"    :    x",
		" x",
	}
	var j int
	for x, bs := range bss {
		ss := fmt.Sprintf("-%d", x)
		b.Run("space"+ss, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				j = trimSpace(bs)
				if bs[j] != ':' {
					b.Fatal("err")
				}
				j = trimSpace(bs[j+1:])
			}
			b.StopTimer()
		})
		b.Run("cron"+ss, func(b *testing.B) {
			n := 0
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				j, n = parseByte(bs, ':')
				if n != 1 {
					b.Fatal("n!=1")
				}
				_ = j
			}
			b.StopTimer()
		})
	}
}

func TestMyUnmarshalStd(t *testing.T) {
	var j = `{
		"BizName": {
			"ZH_CN": "职级",
			"EN_US": "job-level"
		},
		"Description": {
			"ZH_CN": "",
			"EN_US": ""
		}
	}`
	type I18N struct {
		ZH_CN, EN_US string
	}
	// m := map[string]interface{}{}
	m := map[string]I18N{
		"test": {
			"1", "2",
		},
	}
	err := json.Unmarshal([]byte(j), &m)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%T:%+v", m, m)
}

var bsG string
var bsG1 []byte

func BenchmarkTrimSpace(b *testing.B) {
	bs0 := make([]byte, 16+len(j0))
	pbs := (*[1 << 31]byte)(unsafe.Pointer((uintptr(unsafe.Pointer(&bs0[0])) + 0xf) & (^uintptr(0xf))))
	bs := pbs[:len(j0x)]
	d := J0{}
	json.Unmarshal([]byte(j0), &d)
	bs, _ = json.Marshal(&d)
	b.Log(string(bs))
	// copy(bs, j0)
	countG := [7]int{}
	table := [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

	xs1 := "12345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890123456789012345678901234567890"
	for i := 0; i < 3; i++ {
		xs1 += xs1
	}
	b.Run("bytes.IndexByte", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			strings.Index(xs1, "asdfgjlnd")
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("1", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, bb := range bs {
				if table[bb] {
					count++
				}
			}
			if countG[0] == 0 {
				countG[0] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("2", func(b *testing.B) {
		xs := []byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r', ' '}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, x := range xs[:] {
				for i, c := 0, 0; ; {
					c = bytes.IndexByte(bs[i:], x)
					if c < 0 {
						break
					}
					count++
					i = i + c + 1
					for ; bs[i] == x; i++ {
						count++
					}
				}
			}
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-1", func(b *testing.B) {
		xs := []byte{0x85, 0xA0, '\t', '\n', '\v', '\f', '\r'}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for _, x := range xs[:] {
				for i, c := 0, 0; ; {
					c = bytes.IndexByte(bs[i:], x)
					if c < 0 {
						break
					}
					count++
					i = i + c + 1
					for ; bs[i] == x; i++ {
						count++
					}
				}
			}
			// for _, b := range bs {
			// 	if b == ' ' {
			// 		count++
			// 	}
			// }
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-3", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for i, c := 0, 0; ; {
				c = bytes.IndexByte(bs[i:], '\r')
				if c < 0 {
					break
				}
				count++
				i = i + c + 1
			}

		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
	b.Run("2-2", func(b *testing.B) {
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			count := 0
			for i := 0; i < len(bs); i++ {
				if bs[i] == ' ' {
					count++
				}
			}
			if countG[1] == 0 {
				countG[1] = count
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	b.Run("3", func(b *testing.B) {
		table := [256]bool{'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true}

		xs := [4]uint64{}

		for i := 0; i < 256; i++ {
			if table[i] {
				xs[i/64] |= 1 << (uint64(i) % 64)
			}
		}
		b.ReportAllocs()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			var count uint64
			for _, x := range bs {
				// x := uint64(bb)
				// if (xs[x/64] & (1 << (x % 64))) > 0 {
				// 	count++
				// }
				count += (xs[x/64] >> (x % 64)) & 1
			}
			if countG[2] == 0 {
				countG[2] = int(count)
			}
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})

	if countG[0] != countG[1] || countG[0] != countG[2] || countG[0] != countG[3] ||
		countG[0] != countG[4] || countG[0] != countG[5] || countG[0] != countG[6] {
		b.Fatalf("countG:%+v", countG)
	} else {
		b.Logf("countG:%+v", countG)
	}
}

func BenchmarkMapAcess(b *testing.B) {
	m := make(map[string][]byte)
	keys := make([]string, 200)
	for i := range keys {
		key := make([]byte, 20)
		for i := range key {
			key[i] = byte(rand.Uint32())
		}
		m[string(key)] = key
		keys[i] = string(key)
	}
	b.Run("1", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = m[keys[i%len(keys)]]
		}
		b.SetBytes(int64(b.N))
		b.StopTimer()
	})
}

var j0x = `{
	"id": "7028151259660092936",
    "name": {
        "zh-CN": "陈三",
        "en-US": ""
    },
    "avatar": {
        "avatar": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=noop&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar72": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=72x72&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar240": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=240x240&cut_type=&quality=&format=png&sticker_format=.webp",
        "avatar640": "https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=640x640&cut_type=&quality=&format=png&sticker_format=.webp"
    },
    "department": {
        "id": "6826585686905406989",
        "name": {
            "zh-CN": "研发部",
            "en-US": "RD Department"
        }
    },
    "email": "",
    "mobile": {
        "phone": "18030838810",
        "code": "86"
    },
    "status": {
        "accountStatus": false,
        "employmentStatus": false,
        "registerStatus": false
    },
    "employeeType": {
        "id": 1,
        "name": {
            "zh-CN": "正式",
            "en-US": "Regular"
        },
        "active": true
    },
    "isAdmin": false,
    "isLeader": false,
    "isManager": false,
    "isAppManager": false,
    "departmentList": {
        "id": "6826585686905406989",
        "name": {
            "zh-CN": "研发部",
            "en-US": "RD Department"
        }
    }
}`

var j0 = `{"id":"7028151259660092936","name":{"zh-CN":"陈三","en-US":""},"avatar":{"avatar":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=noop&cut_type=&quality=&format=png&sticker_format=.webp","avatar72":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=72x72&cut_type=&quality=&format=png&sticker_format=.webp","avatar240":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=240x240&cut_type=&quality=&format=png&sticker_format=.webp","avatar640":"https://7b90749e-a6be-4a86-9f9b-7dce3d2ecf5g~?image_size=640x640&cut_type=&quality=&format=png&sticker_format=.webp"},"department":{"id":"6826585686905406989","name":{"zh-CN":"研发部","en-US":"RD Department"}},"email":"","mobile":{"phone":"18030838810","code":"86"},"status":{"accountStatus":false,"employmentStatus":false,"registerStatus":false},"employeeType":{"id":1,"name":{"zh-CN":"正式","en-US":"Regular"},"active":true},"isAdmin":false,"isLeader":false,"isManager":false,"isAppManager":false,"departmentList":{"id":"6826585686905406989","name":{"zh-CN":"研发部","en-US":"RD Department"}}}`

type J0 struct {
	ID   string `json:"id"`
	Name struct {
		ZhCN string `json:"zh-CN"`
		EnUS string `json:"en-US"`
	} `json:"name"`
	Avatar struct {
		Avatar    string `json:"avatar"`
		Avatar72  string `json:"avatar72"`
		Avatar240 string `json:"avatar240"`
		Avatar640 string `json:"avatar640"`
	} `json:"avatar"`
	Department struct {
		ID   string `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
	} `json:"department"`
	Email  string `json:"email"`
	Mobile struct {
		Phone string `json:"phone"`
		Code  string `json:"code"`
	} `json:"mobile"`
	Status struct {
		AccountStatus    bool `json:"accountStatus"`
		EmploymentStatus bool `json:"employmentStatus"`
		RegisterStatus   bool `json:"registerStatus"`
	} `json:"status"`
	EmployeeType struct {
		ID   int `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
		Active bool `json:"active"`
	} `json:"employeeType"`
	IsAdmin        bool `json:"isAdmin"`
	IsLeader       bool `json:"isLeader"`
	IsManager      bool `json:"isManager"`
	IsAppManager   bool `json:"isAppManager"`
	DepartmentList struct {
		ID   string `json:"id"`
		Name struct {
			ZhCN string `json:"zh-CN"`
			EnUS string `json:"en-US"`
		} `json:"name"`
	} `json:"departmentList"`
}
