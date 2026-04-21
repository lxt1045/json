// MIT License
//
// Copyright (c) 2021 Xiantu Li

package json

import (
	stdjson "encoding/json"
	"strings"
	"testing"
)

// TestStringEscape 确保 Marshal 对特殊字符按 RFC 8259 转义。
// 这是此前的一个真实 bug, 旧实现只处理 `"` 与 `\\`, 控制字符 (\n \t \r) 未转义,
// 产生的 JSON 不合法也无法被 encoding/json 解析。
func TestStringEscape(t *testing.T) {
	type S struct {
		V string `json:"v"`
	}
	cases := []struct {
		name string
		in   string
	}{
		{"tab", "a\tb"},
		{"newline", "line1\nline2"},
		{"cr", "a\rb"},
		{"backspace", "a\bb"},
		{"formfeed", "a\fb"},
		{"quote", `a"b`},
		{"backslash", `a\b`},
		{"control-01", "a\x01b"},
		{"control-1f", "a\x1fb"},
		{"mixed", "he said: \"hi\"\nline2\tend\\"},
		{"utf8", "你好, мир, 🌍"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			bs, err := Marshal(&S{V: c.in})
			if err != nil {
				t.Fatalf("Marshal error: %v", err)
			}
			// 输出必须能被标准库解析, 并且反序列化回来值相同。
			var got S
			if err := stdjson.Unmarshal(bs, &got); err != nil {
				t.Fatalf("stdjson failed to parse our Marshal output %s: %v", string(bs), err)
			}
			if got.V != c.in {
				t.Fatalf("round-trip mismatch: got %q want %q (raw=%s)", got.V, c.in, string(bs))
			}
		})
	}
}

// TestValidString 验证 Valid 函数的基础用例。
func TestValidString(t *testing.T) {
	good := []string{
		`{}`,
		`[]`,
		`null`,
		`true`,
		`false`,
		`1`,
		`-1.5e10`,
		`"hello"`,
		`{"a":1,"b":[1,2,3],"c":null,"d":"s\n\t\u0041"}`,
		` { "a" : 1 } `,
		`[{"a":[1,2]},{"b":"c"}]`,
	}
	bad := []string{
		``,
		`{`,
		`}`,
		`{"a":}`,
		`[1,]`,
		`01`,
		`"abc`,
		`"a\x"`,
		`{"a":1 "b":2}`,
		`truefalse`,
		`{"a":1}x`,
	}
	for _, s := range good {
		if !ValidString(s) {
			t.Errorf("expected valid: %q", s)
		}
	}
	for _, s := range bad {
		if ValidString(s) {
			t.Errorf("expected invalid: %q", s)
		}
	}
}

// TestMarshalIndent 验证 MarshalIndent 的输出形状, 并与标准库做结构等价比较。
func TestMarshalIndent(t *testing.T) {
	type Inner struct {
		A int    `json:"a"`
		B string `json:"b"`
	}
	type Outer struct {
		Name  string   `json:"name"`
		Items []Inner  `json:"items"`
		Tags  []string `json:"tags"`
		Empty []int    `json:"empty"`
	}
	v := Outer{
		Name:  "demo",
		Items: []Inner{{1, "x"}, {2, "y"}},
		Tags:  []string{"go", "json"},
		Empty: []int{},
	}
	bs, err := MarshalIndent(&v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	// 基本形状检查
	if !strings.Contains(string(bs), "\n  ") {
		t.Fatalf("expected indent in output, got: %s", bs)
	}
	// 必须能被标准库解析
	var back Outer
	if err := stdjson.Unmarshal(bs, &back); err != nil {
		t.Fatalf("stdjson failed to parse MarshalIndent output: %v\n%s", err, bs)
	}
	if back.Name != v.Name || len(back.Items) != len(v.Items) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", back, v)
	}
}
