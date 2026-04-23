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
	"encoding/json"
	"reflect"
	"unsafe"
)

// RawMessage is a raw encoded JSON object.
// It implements Marshaler and Unmarshaler and can
// be used to delay JSON decoding or precompute a JSON encoding.
type RawMessage []byte

var rawMessageType = reflect.TypeOf(RawMessage(nil))

func isRawMessage(typ reflect.Type) bool {
	return typ == rawMessageType
}

// rawMessageMFuncs returns marshal/unmarshal functions for RawMessage.
func rawMessageMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			n, is := parseRawValue(idxSlash, stream, store)
			iSlash = is
			i = n
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			raw := *(*RawMessage)(store.obj)
			if raw == nil {
				out = append(in, "null"...)
				return
			}
			out = append(in, raw...)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		n, is := parseRawValue(idxSlash, stream, store)
		iSlash = is
		i = n
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			out = append(in, "null"...)
			return
		}
		store.obj = p
		raw := *(*RawMessage)(store.obj)
		if raw == nil {
			out = append(in, "null"...)
			return
		}
		out = append(in, raw...)
		return
	}
	return
}

// parseRawValue parses a single JSON value and stores its raw bytes into a RawMessage field.
func parseRawValue(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	start := 0
	for stream[start] == ' ' || stream[start] == '\t' || stream[start] == '\n' || stream[start] == '\r' {
		start++
	}
	end := scanJSONValue(stream[start:])
	raw := stream[start : start+end]
	*(*RawMessage)(store.obj) = append([]byte(nil), raw...)
	i = start + end
	return
}

// scanJSONValue scans a single JSON value from the start of stream and returns its length.
func scanJSONValue(stream string) int {
	i := 0
	for stream[i] == ' ' || stream[i] == '\t' || stream[i] == '\n' || stream[i] == '\r' {
		i++
	}
	start := i
	c := stream[i]
	switch c {
	case '{':
		return scanBracket(stream[start:], '{', '}') + start
	case '[':
		return scanBracket(stream[start:], '[', ']') + start
	case '"':
		return scanString(stream[start:]) + start
	case 't': // true
		if stream[i:i+4] == "true" {
			return 4 + start
		}
	case 'f': // false
		if stream[i:i+5] == "false" {
			return 5 + start
		}
	case 'n': // null
		if stream[i:i+4] == "null" {
			return 4 + start
		}
	default:
		// number
		for i < len(stream) {
			c = stream[i]
			if c == ',' || c == '}' || c == ']' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
				break
			}
			i++
		}
		return i + start
	}
	return start
}

func scanBracket(stream string, left, right byte) int {
	depth := 0
	inString := false
	escape := false
	for i := 0; i < len(stream); i++ {
		c := stream[i]
		if inString {
			if escape {
				escape = false
				continue
			}
			if c == '\\' {
				escape = true
				continue
			}
			if c == '"' {
				inString = false
			}
			continue
		}
		if c == '"' {
			inString = true
			continue
		}
		if c == left {
			depth++
			continue
		}
		if c == right {
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return len(stream)
}

func scanString(stream string) int {
	escape := false
	for i := 1; i < len(stream); i++ {
		c := stream[i]
		if escape {
			escape = false
			continue
		}
		if c == '\\' {
			escape = true
			continue
		}
		if c == '"' {
			return i + 1
		}
	}
	return len(stream)
}

// MarshalJSON returns m as the JSON encoding of m.
func (m RawMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return []byte("null"), nil
	}
	return m, nil
}

// UnmarshalJSON sets *m to a copy of data.
func (m *RawMessage) UnmarshalJSON(data []byte) error {
	if m == nil {
		return json.Unmarshal(data, m)
	}
	*m = append((*m)[:0], data...)
	return nil
}
