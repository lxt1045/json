// MIT License
//
// Copyright (c) 2021 Xiantu Li

package json

import (
	"fmt"

	lxterrs "github.com/lxt1045/errors"
)

// MarshalIndent 调用 Marshal 序列化对象后, 按照 prefix/indent 进行缩进美化。
// 语义与 encoding/json.MarshalIndent 保持一致:
//
//	MarshalIndent(v, "", "  ") => 使用两个空格缩进
//	MarshalIndent(v, "> ", "\t") => 每行前缀 "> ", 级缩进为制表符
//
// 该实现基于字节扫描, 对 Marshal 结果做流式缩进处理, 不额外解析。
// 时间复杂度 O(n), 额外分配一次 []byte。
func MarshalIndent(in interface{}, prefix, indent string) (out []byte, err error) {
	raw, err := Marshal(in)
	if err != nil {
		return nil, err
	}
	out = AppendIndent(nil, raw, prefix, indent)
	return
}

// AppendIndent 把紧凑的 JSON src 缩进后追加到 dst 并返回新 slice。
// 行为与 encoding/json.Indent 一致但使用 append 风格方便内嵌到缓冲区中。
func AppendIndent(dst, src []byte, prefix, indent string) []byte {
	depth := 0
	inStr := false
	esc := false

	writeNewline := func() {
		dst = append(dst, '\n')
		dst = append(dst, prefix...)
		for i := 0; i < depth; i++ {
			dst = append(dst, indent...)
		}
	}

	for i := 0; i < len(src); i++ {
		c := src[i]
		if inStr {
			dst = append(dst, c)
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}

		switch c {
		case ' ', '\t', '\r', '\n':
			// 跳过紧凑 JSON 中的空白 (Marshal 不会产出, 这里也防御外部输入)
			continue
		case '"':
			inStr = true
			dst = append(dst, c)
		case '{', '[':
			dst = append(dst, c)
			// 空对象/空数组保持紧凑
			j := i + 1
			for j < len(src) && (src[j] == ' ' || src[j] == '\t' || src[j] == '\r' || src[j] == '\n') {
				j++
			}
			if j < len(src) && (src[j] == '}' || src[j] == ']') {
				dst = append(dst, src[j])
				i = j
				continue
			}
			depth++
			writeNewline()
		case '}', ']':
			depth--
			writeNewline()
			dst = append(dst, c)
		case ',':
			dst = append(dst, c)
			writeNewline()
		case ':':
			dst = append(dst, c, ' ')
		default:
			dst = append(dst, c)
		}
	}
	return dst
}

// Valid 用只读扫描判断给定字节是否是合法的 JSON (RFC 8259)。
// 相比 Unmarshal, 不做实际的对象解析, 也不分配内存, 更适合用作"快速校验"。
// 忽略首尾空白。
func Valid(data []byte) bool {
	return ValidString(bytesString(data))
}

// ValidString 同 Valid, 接收字符串。
func ValidString(s string) bool {
	v := validator{s: s}
	v.skipSpace()
	if !v.value() {
		return false
	}
	v.skipSpace()
	return v.i == len(v.s)
}

type validator struct {
	s string
	i int
}

func (v *validator) skipSpace() {
	for v.i < len(v.s) {
		c := v.s[v.i]
		if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
			return
		}
		v.i++
	}
}

func (v *validator) value() bool {
	v.skipSpace()
	if v.i >= len(v.s) {
		return false
	}
	switch v.s[v.i] {
	case '{':
		return v.object()
	case '[':
		return v.array()
	case '"':
		return v.string()
	case 't':
		return v.literal("true")
	case 'f':
		return v.literal("false")
	case 'n':
		return v.literal("null")
	default:
		return v.number()
	}
}

func (v *validator) literal(lit string) bool {
	if v.i+len(lit) > len(v.s) {
		return false
	}
	if v.s[v.i:v.i+len(lit)] != lit {
		return false
	}
	v.i += len(lit)
	return true
}

func (v *validator) object() bool {
	v.i++ // '{'
	v.skipSpace()
	if v.i < len(v.s) && v.s[v.i] == '}' {
		v.i++
		return true
	}
	for {
		v.skipSpace()
		if v.i >= len(v.s) || v.s[v.i] != '"' {
			return false
		}
		if !v.string() {
			return false
		}
		v.skipSpace()
		if v.i >= len(v.s) || v.s[v.i] != ':' {
			return false
		}
		v.i++
		if !v.value() {
			return false
		}
		v.skipSpace()
		if v.i >= len(v.s) {
			return false
		}
		switch v.s[v.i] {
		case ',':
			v.i++
		case '}':
			v.i++
			return true
		default:
			return false
		}
	}
}

func (v *validator) array() bool {
	v.i++ // '['
	v.skipSpace()
	if v.i < len(v.s) && v.s[v.i] == ']' {
		v.i++
		return true
	}
	for {
		if !v.value() {
			return false
		}
		v.skipSpace()
		if v.i >= len(v.s) {
			return false
		}
		switch v.s[v.i] {
		case ',':
			v.i++
		case ']':
			v.i++
			return true
		default:
			return false
		}
	}
}

func (v *validator) string() bool {
	if v.i >= len(v.s) || v.s[v.i] != '"' {
		return false
	}
	v.i++
	for v.i < len(v.s) {
		c := v.s[v.i]
		if c == '"' {
			v.i++
			return true
		}
		if c < 0x20 {
			return false
		}
		if c != '\\' {
			v.i++
			continue
		}
		v.i++
		if v.i >= len(v.s) {
			return false
		}
		switch v.s[v.i] {
		case '"', '\\', '/', 'b', 'f', 'n', 'r', 't':
			v.i++
		case 'u':
			if v.i+5 > len(v.s) {
				return false
			}
			for k := 1; k <= 4; k++ {
				c := v.s[v.i+k]
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
					return false
				}
			}
			v.i += 5
		default:
			return false
		}
	}
	return false
}

func (v *validator) number() bool {
	start := v.i
	if v.i < len(v.s) && v.s[v.i] == '-' {
		v.i++
	}
	// int part
	if v.i >= len(v.s) {
		return false
	}
	if v.s[v.i] == '0' {
		v.i++
	} else if v.s[v.i] >= '1' && v.s[v.i] <= '9' {
		for v.i < len(v.s) && v.s[v.i] >= '0' && v.s[v.i] <= '9' {
			v.i++
		}
	} else {
		return false
	}
	// frac
	if v.i < len(v.s) && v.s[v.i] == '.' {
		v.i++
		if v.i >= len(v.s) || v.s[v.i] < '0' || v.s[v.i] > '9' {
			return false
		}
		for v.i < len(v.s) && v.s[v.i] >= '0' && v.s[v.i] <= '9' {
			v.i++
		}
	}
	// exp
	if v.i < len(v.s) && (v.s[v.i] == 'e' || v.s[v.i] == 'E') {
		v.i++
		if v.i < len(v.s) && (v.s[v.i] == '+' || v.s[v.i] == '-') {
			v.i++
		}
		if v.i >= len(v.s) || v.s[v.i] < '0' || v.s[v.i] > '9' {
			return false
		}
		for v.i < len(v.s) && v.s[v.i] >= '0' && v.s[v.i] <= '9' {
			v.i++
		}
	}
	return v.i > start
}

// ensure lxterrs / fmt 被引用, 方便未来 API 扩展 (例如 Compact 的错误返回值)。
var _ = fmt.Sprintf
var _ = lxterrs.New
