package json_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"testing"
	"unsafe"

	"github.com/bytedance/sonic"
	lxt "github.com/lxt1045/json"
	"github.com/lxt1045/json/testdata"
	"github.com/tidwall/gjson"
)

func AppendFileToContent(data []string, files ...string) []string {
	for _, file := range files {
		bs, err := ioutil.ReadFile(file)
		if err != nil {
			panic(err)
		}
		data = append(data, BytesToString(bs))
	}
	return data
}

func BenchmarkUnmarshal(b *testing.B) {
	datas := AppendFileToContent(nil, "./testdata/twitter.json", "./testdata/twitterescaped.json")
	for i, data := range datas {
		b.Logf("i:%d,len:%d", i, len(data))
		bs := StringToBytes(data)
		runs := []struct {
			name string
			f    func()
		}{
			{"std-map",
				func() {
					var m map[string]interface{}
					json.Unmarshal(bs, &m)
				},
			},
			{"std-st",
				func() {
					var m testdata.Twitter
					json.Unmarshal(bs, &m)
				},
			},
			{"gjson-Parse",
				func() {
					gjson.Parse(data).Value()
				},
			},
			{"lxt-map",
				func() {
					var m map[string]interface{}
					lxt.UnmarshalString(data, &m)
				},
			},
			{"lxt-st",
				func() {
					var m testdata.Twitter
					lxt.UnmarshalString(data, &m)
				},
			},
			{
				"sonic-map",
				func() {
					var m map[string]interface{}
					sonic.UnmarshalString(data, &m)
				},
			},
			{
				"sonic-st",
				func() {
					var m testdata.Twitter
					sonic.UnmarshalString(data, &m)
				},
			},
		}
		for _, r := range runs {
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
}

func BenchmarkMmarshal(b *testing.B) {
	datas := AppendFileToContent(nil, "./testdata/twitter.json")
	var m map[string]interface{}
	data := datas[0]
	bs := []byte(data)
	json.Unmarshal(bs, &m)
	var st testdata.Twitter
	json.Unmarshal(bs, &st)
	runs := []struct {
		name string
		f    func()
	}{
		{"std-map",
			func() {
				json.Marshal(&m)
			},
		},
		{"std-st",
			func() {
				json.Marshal(&st)
			},
		},
		{"lxt-map",
			func() {
				lxt.Marshal(&m)
			},
		},
		{"lxt-st",
			func() {
				lxt.Marshal(&st)
			},
		},
		{
			"sonic-map",
			func() {
				sonic.Marshal(&m)
			},
		},
		{
			"sonic-st",
			func() {
				sonic.Marshal(&st)
			},
		},
	}
	for _, r := range runs {
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

//BytesToString ...
func BytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

//StringToBytes ...
func StringToBytes(s string) []byte {
	strH := (*reflect.StringHeader)(unsafe.Pointer(&s))
	p := reflect.SliceHeader{
		Data: strH.Data,
		Len:  strH.Len,
		Cap:  strH.Len,
	}
	return *(*[]byte)(unsafe.Pointer(&p))
}

/*

go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalType$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkUnmarshalType$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof


BenchmarkUnmarshalType/[]int8-10-lxt
BenchmarkUnmarshalType/[]int8-10-lxt-12         	  912650	      1255 ns/op	      60 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int8-10-sonic
BenchmarkUnmarshalType/[]int8-10-sonic-12       	 1809181	       648.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int-10-lxt
BenchmarkUnmarshalType/[]int-10-lxt-12          	  869449	      1217 ns/op	    1018 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int-10-sonic
BenchmarkUnmarshalType/[]int-10-sonic-12        	 1950355	       620.3 ns/op	       0 B/op	       0 allocs/op
*/

func BenchmarkUnmarshalType(b *testing.B) {
	type X struct {
		A string
		B string
	}
	type Y struct {
		A bool
		B bool
	}
	all := []struct {
		V     interface{}
		JsonV string
	}{
		{uint(0), `888888`},            // 0
		{(*uint)(nil), `888888`},       // 1
		{int8(0), `88`},                // 2
		{int(0), `888888`},             // 3
		{true, `true`},                 // 4
		{"", `"asdfghjkl"`},            // 5
		{[]int8{}, `[1,2,3]`},          // 6
		{[]int{}, `[1,2,3]`},           // 7
		{[]bool{}, `[true,true,true]`}, // 8
		{[]string{}, `["1","2","3"]`},  // 9
		{[]X{}, `[{"A":"aaaa","B":"bbbb"},{"A":"aaaa","B":"bbbb"},{"A":"aaaa","B":"bbbb"}]`},
		{[]Y{}, `[{"A":true,"B":true},{"A":true,"B":true},{"A":true,"B":true}]`},
		{(*int)(nil), `88`},             // 11
		{(*bool)(nil), `true`},          // 12
		{(*string)(nil), `"asdfghjkl"`}, // 13
	}
	N := 10
	idxs := []int{}
	// idxs = []int{8, 9, 10, 11}
	if len(idxs) > 0 {
		get := all[:0]
		for _, i := range idxs {
			get = append(get, all[i])
		}
		all = get
	}
	for _, obj := range all {
		builder := lxt.NewTypeBuilder()
		buf := bytes.NewBufferString("{")
		fieldType := reflect.TypeOf(obj.V)
		for i := 0; i < N; i++ {
			if i != 0 {
				buf.WriteByte(',')
			}
			key := fmt.Sprintf("Field_%d", i)
			builder.AddField(key, fieldType)
			buf.WriteString(fmt.Sprintf(`"%s":%v`, key, obj.JsonV))
		}
		buf.WriteByte('}')
		bs := buf.Bytes()
		str := string(bs)

		// b.Logf("json:%s", str)
		typ := builder.Build()
		value := reflect.New(typ).Interface()
		// b.Logf("\ntype:%T\n", value)

		runtime.GC()
		b.Run(fmt.Sprintf("%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := lxt.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		runtime.GC()
		b.Run(fmt.Sprintf("%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := sonic.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		// continue
		runtime.GC()
		// b.Run(fmt.Sprintf("%s-%d-std", fieldType, N), func(b *testing.B) {
		// 	for i := 0; i < b.N; i++ {
		// 		err := json.Unmarshal(bs, value)
		// 		if err != nil {
		// 			b.Fatal(err)
		// 		}
		// 	}
		// })
		// runtime.GC()

		var err error
		b.Run(fmt.Sprintf("Marshal-%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = lxt.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		runtime.GC()
		b.Run(fmt.Sprintf("Marshal-%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = sonic.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		// runtime.GC()
		// b.Run(fmt.Sprintf("Marshal-%s-%d-std", fieldType, N), func(b *testing.B) {
		// 	for i := 0; i < b.N; i++ {
		// 		bs, err = json.Marshal(value)
		// 		if err != nil {
		// 			b.Fatal(err)
		// 		}
		// 	}
		// })
	}
}

func BenchmarkUnmarshalInterface(b *testing.B) {

	all := []struct {
		V     interface{}
		JsonV string
	}{
		{int(0), `888888`},
		{true, `true`},
		{"", `"asdfghjkl"`},
	}
	N := 10
	var value = &struct {
		Field_0 interface{}
		Field_1 interface{}
		Field_2 interface{}
		Field_3 interface{}
		Field_4 interface{}
		Field_5 interface{}
		Field_6 interface{}
		Field_7 interface{}
		Field_8 interface{}
		Field_9 interface{}
	}{}
	for _, obj := range all {
		buf := bytes.NewBufferString("{")
		for i := 0; i < N; i++ {
			if i != 0 {
				buf.WriteByte(',')
			}
			key := fmt.Sprintf("Field_%d", i)
			buf.WriteString(fmt.Sprintf(`"%s":%s`, key, obj.JsonV))
		}
		buf.WriteByte('}')
		bs := buf.Bytes()
		str := string(bs)
		fieldType := reflect.TypeOf(obj.V)

		b.Run(fmt.Sprintf("%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := lxt.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := sonic.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("%s-%d-std", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := json.Unmarshal(bs, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		var err error
		b.Run(fmt.Sprintf("Marshal-%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = lxt.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("Marshal-%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = sonic.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("Marshal-%s-%d-std", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = json.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkUnmarshalMapInterface(b *testing.B) {
	all := []struct {
		V     interface{}
		JsonV string
	}{
		{int(0), `888888`},
		{true, `true`},
		{"", `"asdfghjkl"`},
	}
	N := 10
	var value = &map[string]interface{}{}
	for _, obj := range all {
		buf := bytes.NewBufferString("{")
		for i := 0; i < N; i++ {
			if i != 0 {
				buf.WriteByte(',')
			}
			key := fmt.Sprintf("Field_%d", i)
			buf.WriteString(fmt.Sprintf(`"%s":%s`, key, obj.JsonV))
		}
		buf.WriteByte('}')
		bs := buf.Bytes()
		str := string(bs)
		fieldType := reflect.TypeOf(obj.V)

		b.Run(fmt.Sprintf("%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := lxt.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := sonic.UnmarshalString(str, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("%s-%d-std", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				err := json.Unmarshal(bs, value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})

		var err error
		b.Run(fmt.Sprintf("Marshal-%s-%d-lxt", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = lxt.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("Marshal-%s-%d-sonic", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = sonic.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
		b.Run(fmt.Sprintf("Marshal-%s-%d-std", fieldType, N), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				bs, err = json.Marshal(value)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TODO: slice
/*
go test -benchmem -run=^$ -bench ^BenchmarkObj$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof -c
go test -benchmem -run=^$ -bench ^BenchmarkObj$ github.com/lxt1045/json -count=1 -v -memprofile cpu.prof -c
go tool pprof ./json.test cpu.prof
//   */
func BenchmarkStrings(b *testing.B) {
	// str := `{"X0":["1","2","3"]}`
	// str := `{"X0":["1","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3"],"X1":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X2":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X3":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X4":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X5":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X6":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X7":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X8":["1","2","3","2","3","2","3","2","3","2","3","2","3"],"X9":["1","2","3","2","3","2","3","2","3","2","3","2","3"]}`
	str := `{"X0":["1","2","3"],"X1":["1","2","3"],"X2":["1","2","3"],"X3":["1","2","3"],"X4":["1","2","3"],"X5":["1","2","3"],"X6":["1","2","3"],"X7":["1","2","3"],"X8":["1","2","3"],"X9":["1","2","3"]}`
	// str := `{"X0":["1","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3","2","3"]}`
	// str := `{"X0":["1","2","3","2","3","1","2","3","2","3"]}`
	d := struct {
		X0 []string
		X1 []string
		X2 []string
		X3 []string
		X4 []string
		X5 []string
		X6 []string
		X7 []string
		X8 []string
		X9 []string
	}{}
	b.Run("lxt-strings", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := lxt.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("sonic-strings", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}

/*

BenchmarkObj/lxt-obj
BenchmarkObj/lxt-obj-12         	  394314	      3049 ns/op	129330.79 MB/s	    1286 B/op	       0 allocs/op
BenchmarkObj/sonic-obj
BenchmarkObj/sonic-obj-12       	  357523	      2850 ns/op	125466.19 MB/s	       4 B/op	       0 allocs/op
*/
func BenchmarkObj(b *testing.B) {
	str := `{"X0":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X1":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X2":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X3":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X4":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X5":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X6":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X7":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X8":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}],"X9":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}]}`
	// str := `{"X0":[{"A":"1","B":"2"},{"A":"1","B":"2"},{"A":"1","B":"2"}]}`
	type X struct {
		A string
		B string
		C *int
	}
	d := struct {
		X0 []X
		X1 []X
		X2 []X
		X3 []X
		X4 []X
		X5 []X
		X6 []X
		X7 []X
		X8 []X
		X9 []X
	}{}
	// err := lxt.UnmarshalString(str, &d)
	// if err != nil {
	// 	b.Fatalf("[%d]:%v", 0, err)
	// }
	// b.Logf("json:%+v", d)
	b.Run("lxt-obj", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := lxt.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
	// return
	b.Run("sonic-obj", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			err := sonic.UnmarshalString(str, &d)
			if err != nil {
				b.Fatalf("[%d]:%v", i, err)
			}
		}
		b.StopTimer()
		b.SetBytes(int64(b.N))
	})
}
