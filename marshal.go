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
	"encoding"
	stdjson "encoding/json"
	"reflect"
	"strconv"

	lxterrs "github.com/lxt1045/errors"
)

/*
    fget 函数 list ，打入slice中，然后依次遍历slice，避免了遍历struct 的多个for循环！！ greate！
    golang不支持尾递归，可能嵌套调用性能没有fget list 方式好
*/

//bsGrow 在 slice marshal 的时候可以根据第一个的序列化结果计算出还需要的长度，如果 bs 不足则提前分配是比较好的选择
func bsGrow(in []byte, lNeed int) (out []byte) {
	lNew := bsPoolN
	if lNew < lNeed+len(in) {
		lNew += lNeed + len(in)
	}
	bsNew := make([]byte, 0, lNew)
	out = append(bsNew, in...)
	return
}

// 排列一个 fM list，优化掉多个 for 循环
func marshalStruct(store Store, in []byte) (out []byte) {
	out = append(in, '{')
	for _, tag := range store.tag.ChildList {
		out = append(out, tag.TagName...)
		out = append(out, ':')

		pObj := pointerOffset(store.obj, tag.Offset)
		out = tag.fM(Store{obj: pObj, tag: tag}, out)

		out = append(out, ',')
	}
	out[len(out)-1] = '}'
	return
}

//marshalT 序列化明确的类型
func marshalT(in []byte, store Store) (out []byte) {
	out = in
	panic(lxterrs.Errorf("error tag, fM is nil:%+v", store.tag))

	return
}

// TODO: 根据一个 marshal 的 len 乘以 len(m) ，达到还需要的空间，如果不足则 New？
func marshalSlice(bs []byte, store Store, l int) (out []byte) {
	out = append(bs, '[')
	if l <= 0 {
		out = append(out, ']')
		return
	}
	son := store.tag
	size := son.TypeSize

	lBefore := len(out)
	out = son.fM(Store{obj: store.obj, tag: son}, out)
	lObj := len(out) - lBefore + 1 + 16 // 16 随意取的值
	// 解析还需要的空间
	if lNeed := lObj * (l - 1); cap(out)-len(out) < lNeed {
		out = bsGrow(out, lNeed)
	}

	for i := 1; i < l; i++ {
		out = append(out, ',')
		pSon := pointerOffset(store.obj, uintptr(i*size))
		out = son.fM(Store{obj: pSon, tag: son}, out)
	}
	out = append(out, ']')
	return
}
func marshalInterface(bs []byte, iface interface{}) (out []byte) {
	if iface == nil {
		out = append(bs, "null"...)
		return
	}
	out = bs
	switch v := iface.(type) {
	case int8:
		out = strconv.AppendInt(out, int64(v), 10)
	case int16:
		out = strconv.AppendInt(out, int64(v), 10)
	case int32:
		out = strconv.AppendInt(out, int64(v), 10)
	case int64:
		out = strconv.AppendInt(out, int64(v), 10)
	case int:
		out = strconv.AppendInt(out, int64(v), 10)
	case uint8:
		out = strconv.AppendUint(out, uint64(v), 10)
	case uint16:
		out = strconv.AppendUint(out, uint64(v), 10)
	case uint32:
		out = strconv.AppendUint(out, uint64(v), 10)
	case uint64:
		out = strconv.AppendUint(out, uint64(v), 10)
	case uint:
		out = strconv.AppendUint(out, uint64(v), 10)
	case float32:
		out = strconv.AppendFloat(out, float64(v), 'f', -1, 32)
	case float64:
		out = strconv.AppendFloat(out, v, 'f', -1, 64)
	case string:
		out = append(out, '"')
		out = append(out, v...)
		out = append(out, '"')
	case stdjson.Number:
		// TODO: 检查 numStr 的有效性？
		out = append(out, v.String()...)
	case map[string]interface{}:
		out = marshalMapInterface(out, v)
	case []interface{}:
		// TODO
	case []byte:
		panic("TODO: []byte...")
	default:
		value := reflect.ValueOf(iface)
		out = marshalValue(out, value)
	}
	return
}

// TODO: 根据一个 marshal 的 len 乘以 len(m) ，达到还需要的空间，如果不足则 New？
func marshalMapInterface(bs []byte, m map[string]interface{}) (out []byte) {
	out = bs
	out = append(out, '{')
	if len(m) == 0 {
		out = append(out, '}')
		return
	}

	first := true
	for k, v := range m {
		if first {
			first = false
			lBefore := len(out)
			out = append(out, '"')
			out = append(out, k...)
			out = append(out, `":`...)
			out = marshalInterface(out, v)
			lObj := len(out) - lBefore + 1 + 16 // 16 随意取的值
			if lNeed := lObj * (len(m) - 1); cap(out)-len(out) < lNeed {
				out = bsGrow(out, lNeed)
			}
			continue
		}
		out = append(out, `,"`...)
		out = append(out, k...)
		out = append(out, `":`...)
		out = marshalInterface(out, v)
	}
	out = append(out, '}')
	return
}

func marshalValue(bs []byte, value reflect.Value) (out []byte) {
	out = bs

	// 针对指针
	for value.Kind() == reflect.Ptr {
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		value = value.Elem()
	}

	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		// UnpackEface(&value)
		// 从 iterface{} 里取出原始类型
		value := reflect.ValueOf(value.Interface())
		out = marshalValue(bs, value)
		return
	case reflect.Map:
		if value.IsNil() {
			out = append(out, "null"...)
			return
		}
		iter := value.MapRange()
		if iter == nil {
			return
		}
		out = append(out, '{')
		l := len(out)
		for iter.Next() {
			out = marshalKey(out, iter.Key())
			out = append(out, ':')
			out = marshalValue(out, iter.Value())
			out = append(out, ',')
		}
		if l < len(out) {
			out = out[:len(out)-1]
		}
		out = append(out, '}')
		return
	case reflect.Slice:
		if value.IsNil() {
			out = append(out, "[]"...)
			return
		}
		out = append(out, '[')
		lBefore := len(out)
		v := value.Index(0)
		out = marshalValue(out, v)
		lObj := len(out) - lBefore + 1 + 16 // 16 随意取的值
		// 解析还需要的空间
		if lNeed := lObj * (value.Len() - 1); cap(out)-len(out) < lNeed {
			out = bsGrow(out, lNeed)
		}
		for i := 1; i < value.Len(); i++ {
			out = append(out, ',')
			v := value.Index(i)
			out = marshalValue(out, v)
		}
		out = append(out, ']')
		return
	case reflect.Struct:
		prv := reflectValueToValue(&value)
		goType := prv.typ

		tag, err := LoadTagNode(value, goType.Hash)
		if err != nil {
			return
		}

		store := Store{
			tag: tag,
			obj: prv.ptr, // eface.Value,
		}
		out = marshalStruct(store, out)
		return
	case reflect.Bool:
		if value.Bool() {
			out = append(out, "true"...)
		} else {
			out = append(out, "false"...)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		out = strconv.AppendUint(out, value.Uint(), 10)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		out = strconv.AppendInt(out, value.Int(), 10)
	case reflect.Float64:
		out = strconv.AppendFloat(out, value.Float(), 'f', -1, 64)
	case reflect.Float32:
		out = strconv.AppendFloat(out, value.Float(), 'f', -1, 32)
	case reflect.String:
		if value.Type() == jsonNumberType {
			numStr := value.String()
			// TODO: 检查 numStr 的有效性？
			out = append(out, numStr...)
			return
		}
		out = append(out, '"')
		out = append(out, value.String()...)
		out = append(out, '"')
		return
	default:
		out = append(out, "null"...)
	}

	return
}

var jsonNumberType = reflect.TypeOf(stdjson.Number(""))

func marshalKey(in []byte, k reflect.Value) (out []byte) {
	out = in
	if k.Kind() == reflect.String {
		// key = k.String()
		out = append(out, '"')
		out = append(out, k.String()...)
		out = append(out, '"')
		return
	}
	if tm, ok := k.Interface().(encoding.TextMarshaler); ok {
		if k.Kind() == reflect.Pointer && k.IsNil() {
			return
		}
		bs, err := tm.MarshalText()
		if err != nil {
			err = lxterrs.Wrap(err, "MarshalText() got error")
			panic(err)
		}
		// key = string(bs)
		out = append(out, '"')
		out = append(out, bs...)
		out = append(out, '"')
		return
	}
	switch k.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// key = strconv.FormatInt(k.Int(), 10)
		out = append(out, '"')
		out = strconv.AppendInt(out, k.Int(), 10)
		out = append(out, '"')
		return
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		// key = strconv.FormatUint(k.Uint(), 10)
		out = append(out, '"')
		out = strconv.AppendUint(out, k.Uint(), 10)
		out = append(out, '"')
		return
	}
	err := lxterrs.New("unexpected map key type")
	panic(err)
}
