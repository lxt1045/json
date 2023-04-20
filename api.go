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
	"fmt"
	"reflect"
	"sync/atomic"

	lxterrs "github.com/lxt1045/errors"
)

//Unmarshal 转成struct
func Unmarshal(bsIn []byte, in interface{}) (err error) {
	return UnmarshalString(bytesString(bsIn), in)
}

//UnmarshalString Unmarshal string
func UnmarshalString(bs string, in interface{}) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if err1, ok := e.(*lxterrs.Code); ok {
				err = err1
			} else {
				err = lxterrs.New("%+v", e)
			}
			return
		}
	}()
	i := trimSpace(bs)
	if mIn, ok := in.(*interface{}); ok {
		if bs[i] != '{' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		m, _, _ := parseMapInterface(-1, bs[i+1:])
		*mIn = m
		return nil
	}
	if mIn, ok := in.(*map[string]interface{}); ok {
		if bs[i] != '{' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		m, _, _ := parseMapInterface(-1, bs[i+1:])
		*mIn = m
		return nil
	}
	if _, ok := in.(*[]interface{}); ok {
		if bs[i] != '[' {
			err = fmt.Errorf("json must start with '{' or '[', %s", ErrStream(bs[i:]))
			return
		}
		out := make([]interface{}, 0, 32)
		parseObjToSlice(bs[i+1:], out)
		return
	}

	// 解引用； TODO: 用 Value 的方式提高性能
	vi := reflect.Indirect(reflect.ValueOf(in))
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		return
	}
	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, err := LoadTagNode(vi, goType.Hash)
	if err != nil {
		return
	}

	store := PoolStore{
		tag: tag,
		obj: prv.ptr, // eface.Value,
		// pointerPool: tag.ptrCache.Get(),
	}
	if tag.ptrCache != nil {
		store.pointerPool = tag.ptrCache.Get()
	}

	err = parseRoot(bs[i:], store)

	return
}

//Marshal []byte
func Marshal(in interface{}) (bs []byte, err error) {
	defer func() {
		if e := recover(); e != nil {
			if err1, ok := e.(*lxterrs.Code); ok {
				err = err1
			} else {
				err = lxterrs.New("%+v", e)
			}
			return
		}
	}()

	pbs := bsPool.Get().(*[]byte)
	bs = *pbs
	var lLeft int32 = 1024
	defer func() {
		if cap(bs)-len(bs) >= int(lLeft) {
			*pbs = bs[len(bs):]
			bs = bs[:len(bs):len(bs)]
			bsPool.Put(pbs)
		}
	}()

	if mIn, ok := in.(*interface{}); ok {
		bs = marshalInterface(bs[:0], *mIn)
		return
	}
	if mIn, ok := in.(*map[string]interface{}); ok {
		bs = marshalMapInterface(bs[:0], *mIn)
		return
	}
	if _, ok := in.(*[]interface{}); ok {

		return
	}

	vi := reflect.Indirect(reflect.ValueOf(in))
	if !vi.CanSet() {
		err = fmt.Errorf("%T cannot set", in)
		return
	}
	prv := reflectValueToValue(&vi)
	goType := prv.typ
	tag, err := LoadTagNode(vi, goType.Hash)
	if err != nil {
		return
	}

	store := Store{
		tag: tag,
		obj: prv.ptr, // eface.Value,
	}

	bs = marshalStruct(store, bs[:0])

	l := int32(len(bs))
	lLeft = atomic.LoadInt32(&tag.bsMarshalLen)
	if lLeft > l*2 {
		bsHaftCount := atomic.AddInt32(&tag.bsHaftCount, -1)
		if bsHaftCount < 1000 {
			atomic.StoreInt32(&tag.bsMarshalLen, l)
			lLeft = l
		}
	} else if lLeft < l {
		atomic.StoreInt32(&tag.bsMarshalLen, l)
		lLeft = l
	} else {
		atomic.AddInt32(&tag.bsHaftCount, 1)
	}
	return
}
