package json

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	asrt "github.com/stretchr/testify/assert"
)

func TestStruct(t *testing.T) {
	type Anonymous struct {
		Count int `json:"count"`
		X     string
	}
	fLine := func() string {
		_, file, line, _ := runtime.Caller(1)
		_, file = filepath.Split(file)
		return file + ":" + strconv.Itoa(line)
	}
	idx := -1

	datas := []struct {
		name   string
		bs     string
		target string
		data   interface{}
	}{
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": -0.1 , "struct_0": [{ "count":8,"Anonymous":{"X":"xxx","Count":1}}]}`,
			target: `{"out":-0.1,"struct_0":[{"count":8,"Anonymous":{"count":0,"X":"xxx"}}]}`,
			data: &struct {
				Out    float32 `json:"out"`
				Struct []struct {
					Count     int `json:"count"`
					Anonymous Anonymous
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": [{ "count":8,"Anonymous":{"X":"xxx","Count":1}}]}`,
			target: `{"out":11,"struct_0":[{"count":8,"Anonymous":{"count":0,"X":"xxx"}}]}`,
			data: &struct {
				Out    int `json:"out"`
				Struct []struct {
					Count     int `json:"count"`
					Anonymous Anonymous
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "interface:" + fLine(),
			bs:     `{"out":88,"struct_0":{"a":"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"}}`,
			target: `{"out":88,"struct_0":{"a":"\u003ca href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\"\u003eTwitter for Mac\u003c/a\u003e"}}`,
			data: &struct {
				Out    int         `json:"out"`
				Struct interface{} `json:"struct_0"`
			}{},
		},
		{
			name:   "interface:" + fLine(),
			bs:     `{"out":"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"}`,
			target: "{\"out\":\"\\u003ca href=\\\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\\\" rel=\\\"\\\\\\\"nofollow\\\\\\\"\\\"\\u003eTwitter for Mac\\u003c/a\\u003e\"}",
			data: &struct {
				Out string `json:"out"`
			}{},
		},
		{
			name:   "interface:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int         `json:"out"`
				Struct interface{} `json:"struct_0"`
			}{},
		},
		{
			name:   "map" + fLine(),
			bs:     `{"out": 11 , "map_0": { "count":8,"y":"yyy"}}`,
			target: `{"out":11,"map_0":{"count":8,"y":"yyy"}}`,
			data:   &map[string]interface{}{},
		},

		// 匿名类型; 指针匿名类型
		{
			name:   "struct-Anonymous:" + fLine(),
			bs:     `{"out": 11 , "count":8,"X":"xxx"}`,
			target: `{"out":11,"count":8,"X":"xxx"}`,
			data: &struct {
				Out int `json:"out"`
				Anonymous
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int `json:"count"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8,"slice":[1,2,3]}}`,
			target: `{"out":11,"struct_0":{"count":8,"slice":[1,2,3]}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int   `json:"count"`
					Slice []int `json:"slice"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "slice:" + fLine(),
			bs:     `{"count":8 , "slice":[1,2,3] }`,
			target: `{"count":8,"slice":[1,2,3]}`,
			data: &struct {
				Count int   `json:"count"`
				Slice []int `json:"slice"`
			}{},
		},
		{
			name:   "bool:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count bool `json:"count"`
				False bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-null:" + fLine(),
			bs:     `{"count":true , "false_0":null }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-empty:" + fLine(),
			bs:     `{"count":true }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "float64:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count float64 `json:"count"`
			}{},
		},
		{
			name:   "float64-ptr:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count *float64 `json:"count"`
			}{},
		},
		{
			name:   "int-ptr:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count *int `json:"count"`
			}{},
		},
		{
			name:   "int:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count int `json:"count"`
			}{},
		},
		{
			name:   "string-ptr:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN *string
			}{},
		},
		{
			name:   "string-notag:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN string
			}{},
		},
		{
			name:   "string:" + fLine(),
			bs:     `{ "ZH_CN":"chinese", "ENUS":"English", "count":8 }`,
			target: `{"ZH_CN":"chinese"}`,
			data: &struct {
				ZHCN string `json:"ZH_CN"`
			}{},
		},
	}
	if idx >= 0 {
		datas = datas[idx : idx+1]
	}

	for i, d := range datas {
		t.Run(d.name, func(t *testing.T) {
			err := Unmarshal([]byte(d.bs), d.data)
			if err != nil {
				t.Fatalf("[%d]%s, error:%v\n", i, d.name, err)
			}
			bs, err := json.Marshal(d.data)
			if err != nil {
				t.Fatalf("i:%d, %s:%v\n", i, d.name, err)
			}
			if m, ok := (d.data).(*map[string]interface{}); ok {
				var mm map[string]interface{}
				json.Unmarshal([]byte(d.target), &mm)
				for k, v := range *m {
					asrt.Equalf(t, mm[k], v, fmt.Sprintf("i:%d,%s", i, d.name))
				}
				for k, v := range mm {
					asrt.Equalf(t, (*m)[k], v, fmt.Sprintf("i:%d,%s", i, d.name))
				}

				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else if _, ok := (d.data).(*interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else {
				asrt.Equalf(t, d.target, string(bs), fmt.Sprintf("i:%d,%s", i, d.name))
			}

			runtime.GC()
			_ = fmt.Sprintf("d :%+v", d)
		})
	}
}

func TestStructMarshalInterface(t *testing.T) {
	type st struct {
		Count string `json:"count"`
		X     interface{}
	}
	s := &st{
		Count: "xxx",
		X: st{
			Count: "xxx",
			X:     "xxx",
		},
	}
	bs, err := json.Marshal(&s)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("json:%s", string(bs))
}

func TestStructMarshal(t *testing.T) {
	type Anonymous struct {
		Count int `json:"count"`
		X     string
	}
	fLine := func() string {
		_, file, line, _ := runtime.Caller(1)
		_, file = filepath.Split(file)
		return file + ":" + strconv.Itoa(line)
	}
	idx := -5

	datas := []struct {
		name    string
		bs      string
		target  string
		target2 string
		data    interface{}
	}{
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int                    `json:"out"`
				Struct map[string]interface{} `json:"struct_0"`
			}{},
		},
		{
			name:    "interface:" + fLine(),
			bs:      `{"out":88,"struct_0":{"a":"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"}}`,
			target:  `{"out":88,"struct_0":{"a":"\u003ca href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\"\u003eTwitter for Mac\u003c/a\u003e"}}`,
			target2: `{"out":88,"struct_0":{"a":"<a href="//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22" rel="\"nofollow\"">Twitter for Mac</a>"}}`,
			data: &struct {
				Out    int         `json:"out"`
				Struct interface{} `json:"struct_0"`
			}{},
		},
		{
			name:    "interface:" + fLine(),
			bs:      `{"out":"<a href=\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\" rel=\"\\\"nofollow\\\"\">Twitter for Mac</a>"}`,
			target:  "{\"out\":\"\\u003ca href=\\\"//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22\\\" rel=\\\"\\\\\\\"nofollow\\\\\\\"\\\"\\u003eTwitter for Mac\\u003c/a\\u003e\"}",
			target2: `{"out":"<a href="//itunes.apple.com/us/app/twitter/id409789998?mt=12%5C%22" rel="\"nofollow\"">Twitter for Mac</a>"}`,
			data: &struct {
				Out string `json:"out"`
			}{},
		},
		{
			name:   "interface:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int         `json:"out"`
				Struct interface{} `json:"struct_0"`
			}{},
		},
		{
			name:   "map" + fLine(),
			bs:     `{"out": 11 , "map_0": { "count":8,"y":"yyy"}}`,
			target: `{"out":11,"map_0":{"count":8,"y":"yyy"}}`,
			data:   &map[string]interface{}{},
		},

		// 匿名类型; 指针匿名类型
		{
			name:   "struct-Anonymous:" + fLine(),
			bs:     `{"out": 11 , "count":8,"X":"xxx"}`,
			target: `{"out":11,"count":8,"X":"xxx"}`,
			data: &struct {
				Out int `json:"out"`
				Anonymous
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8}}`,
			target: `{"out":11,"struct_0":{"count":8}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int `json:"count"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "struct:" + fLine(),
			bs:     `{"out": 11 , "struct_0": { "count":8,"slice":[1,2,3]}}`,
			target: `{"out":11,"struct_0":{"count":8,"slice":[1,2,3]}}`,
			data: &struct {
				Out    int `json:"out"`
				Struct struct {
					Count int   `json:"count"`
					Slice []int `json:"slice"`
				} `json:"struct_0"`
			}{},
		},
		{
			name:   "slice:" + fLine(),
			bs:     `{"count":8 , "slice":[1,2,3] }`,
			target: `{"count":8,"slice":[1,2,3]}`,
			data: &struct {
				Count int   `json:"count"`
				Slice []int `json:"slice"`
			}{},
		},
		{
			name:   "bool:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count bool `json:"count"`
				False bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr:" + fLine(),
			bs:     `{"count":true , "false_0":false }`,
			target: `{"count":true,"false_0":false}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-null:" + fLine(),
			bs:     `{"count":true , "false_0":null }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "bool-ptr-empty:" + fLine(),
			bs:     `{"count":true }`,
			target: `{"count":true,"false_0":null}`,
			data: &struct {
				Count *bool `json:"count"`
				False *bool `json:"false_0"`
			}{},
		},
		{
			name:   "float64:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count float64 `json:"count"`
			}{},
		},
		{
			name:   "float64-ptr:" + fLine(),
			bs:     `{"count":8.11 }`,
			target: `{"count":8.11}`,
			data: &struct {
				Count *float64 `json:"count"`
			}{},
		},
		{
			name:   "int-ptr:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count *int `json:"count"`
			}{},
		},
		{
			name:   "int:" + fLine(),
			bs:     `{"count":8 }`,
			target: `{"count":8}`,
			data: &struct {
				Count int `json:"count"`
			}{},
		},
		{
			name:   "string-ptr:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN *string
			}{},
		},
		{
			name:   "string-notag:" + fLine(),
			bs:     `{ "ZHCN":"chinese"}`,
			target: `{"ZHCN":"chinese"}`,
			data: &struct {
				ZHCN string
			}{},
		},
		{
			name:   "string:" + fLine(),
			bs:     `{ "ZH_CN":"chinese", "ENUS":"English", "count":8 }`,
			target: `{"ZH_CN":"chinese"}`,
			data: &struct {
				ZHCN string `json:"ZH_CN"`
			}{},
		},
	}
	if idx >= 0 {
		datas = datas[idx : idx+1]
	}

	for i, d := range datas {
		t.Run(d.name, func(t *testing.T) {
			err := Unmarshal([]byte(d.bs), d.data)
			if err != nil {
				t.Fatalf("[%d]%s, error:%v\n", i, d.name, err)
			}
			bs, err := json.Marshal(d.data)
			if err != nil {
				t.Fatalf("i:%d, %s:%v\n", i, d.name, err)
			}
			if _, ok := (d.data).(*map[string]interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else if _, ok := (d.data).(*interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else {
				asrt.Equalf(t, d.target, string(bs), fmt.Sprintf("i:%d,%s", i, d.name))
			}

			runtime.GC()
			_ = fmt.Sprintf("d :%+v", d)

			bsOut, err := Marshal(d.data)
			if err != nil {
				t.Fatalf("i:%d, %s:%v\n", i, d.name, err)
			}
			if _, ok := (d.data).(*map[string]interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else if _, ok := (d.data).(*interface{}); ok {
				t.Logf("\n%s\n%s", string(d.target), string(bs))
				// asrt.EqualValuesf(t, d.target, string(bs), d.name)
			} else {
				target := d.target2
				if target == "" {
					target = d.target
				}
				asrt.Equalf(t, target, string(bsOut), fmt.Sprintf("i:%d,%s", i, d.name))
			}

		})
	}
}

func BenchmarkTestLoop(b *testing.B) {
	loop := 10000000000
	b.Run("loop1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < loop; j++ {
			}
		}
	})
	b.Run("loop1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < loop; j++ {
			}
		}
	})
	b.Run("loop1", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for j := 0; j < loop; j++ {
			}
			for j := 0; j < loop; j++ {
			}
		}
	})
}

func TestStructST(t *testing.T) {
	type ST struct {
		Count int `json:"count"`
		X     string
		ST    *ST
	}
	st := ST{
		Count: 22,
		X:     "xxx",
		ST: &ST{
			Count: 22,
			X:     "xxx",
		},
	}
	bs, err := json.Marshal(&st)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf(string(bs))

	t.Run("qq", func(t *testing.T) {
		// err := Unmarshal([]byte(d.bs), d.data)
		// if err != nil {
		// 	t.Fatalf("[%d]%s, error:%v\n", i, d.name, err)
		// }
		st2 := ST{}
		err := json.Unmarshal(bs, &st2)
		if err != nil {
			t.Fatal(err)
		}
		bs, err := json.Marshal(&st2)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", string(bs))
	})

	t.Run("qq1", func(t *testing.T) {
		st2 := ST{}
		err := Unmarshal(bs, &st2)
		if err != nil {
			t.Fatal(err)
		}
		bs, err := json.Marshal(&st2)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%+v", string(bs))
	})
}
