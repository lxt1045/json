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
	"sync"
	"unsafe"
)

// 构造器
type TypeBuilder struct {
	// 用于存储属性字段
	fields      []reflect.StructField
	Type        reflect.Type
	lazyOffsets []*uintptr

	goType *GoType
	size   int
	pool   sync.Pool
}

func NewTypeBuilder() *TypeBuilder {
	return &TypeBuilder{}
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) New() unsafe.Pointer {
	v := reflect.New(b.Type)
	p := reflectValueToPointer(&v)
	return p
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) NewSlice() unsafe.Pointer {
	p := unsafe_NewArray(b.goType, 1024)
	return p
}

func (b *TypeBuilder) Interface() interface{} {
	v := reflect.New(b.Type)
	return v.Interface()
}
func (b *TypeBuilder) PInterface() (unsafe.Pointer, interface{}) {
	v := reflect.New(b.Type)
	return reflectValueToPointer(&v), v.Interface()
}

// 根据预先添加的字段构建出结构体
func (b *TypeBuilder) Build() reflect.Type {
	if len(b.fields) == 0 {
		return nil
	}
	typ := b.Type
	if b.Type == nil {
		typ = reflect.StructOf(b.fields)
		b.Type = typ
		b.goType = UnpackType(typ)
	}
	for i := 0; i < typ.NumField(); i++ {
		if len(b.lazyOffsets) > i && b.lazyOffsets[i] != nil {
			*b.lazyOffsets[i] = typ.Field(i).Offset
		}
	}
	return typ
}

//Init 调用 Build() 并初始化缓存
func (b *TypeBuilder) Init() {
	b.Build()

	// 处理缓存
	if len(b.fields) == 0 {
		return
	}
	goType := b.goType
	b.size = int(goType.Size)
	N := (8 * 1024 / b.size) + 1
	l := N * b.size
	b.pool.New = func() any {
		// reflect.MakeSlice()
		ps := unsafe_NewArray(goType, N)
		pH := &SliceHeader{
			Data: ps,
			Len:  l,
			Cap:  l,
		}
		return (*[]uint8)(unsafe.Pointer(pH))
	}
}

// TODO: 此处也可以使用  NewBatch  来提高性能？ 提升 10ns ？
func (b *TypeBuilder) NewFromPool() unsafe.Pointer {
	if len(b.fields) == 0 {
		return nil
	}
	s := b.pool.Get().(*[]uint8)
	pp := unsafe.Pointer(&(*s)[0])
	if cap(*s) >= b.size*2 {
		*s = (*s)[b.size:]
		b.pool.Put(s)
	}
	return pp
}

/*
针对 slice，要添加一个 [4]type 的空间作为预分配的资源
*/
func (b *TypeBuilder) AppendTagField(typ reflect.Type, lazyOffset *uintptr) *TypeBuilder {
	name := fmt.Sprintf("F_%d", len(b.fields))
	b.fields = append(b.fields, reflect.StructField{Name: name, Type: typ})
	b.lazyOffsets = append(b.lazyOffsets, lazyOffset)
	return b
}

func (b *TypeBuilder) AppendField(name string, typ reflect.Type, lazyOffset *uintptr) *TypeBuilder {
	b.fields = append(b.fields, reflect.StructField{Name: name, Type: typ})
	b.lazyOffsets = append(b.lazyOffsets, lazyOffset)
	return b
}

func (b *TypeBuilder) AppendPointer(name string, lazyOffset *uintptr) *TypeBuilder {
	var p unsafe.Pointer
	return b.AppendField(name, reflect.TypeOf(p), lazyOffset)
}

func (b *TypeBuilder) AppendIntSlice(name string, lazyOffset *uintptr) *TypeBuilder {
	var s []int
	return b.AppendField(name, reflect.TypeOf(s), lazyOffset)
}

func (b *TypeBuilder) AppendString(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(""), lazyOffset)
}

func (b *TypeBuilder) AppendBool(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(true), lazyOffset)
}

func (b *TypeBuilder) AppendInt64(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(int64(0)), lazyOffset)
}

func (b *TypeBuilder) AppendFloat64(name string, lazyOffset *uintptr) *TypeBuilder {
	return b.AppendField(name, reflect.TypeOf(float64(1.2)), lazyOffset)
}

// 添加字段
func (b *TypeBuilder) AddField(field string, typ reflect.Type) *TypeBuilder {
	b.fields = append(b.fields, reflect.StructField{Name: field, Type: typ})
	return b
}

func (b *TypeBuilder) AddString(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(""))
}

func (b *TypeBuilder) AddBool(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(true))
}

func (b *TypeBuilder) AddInt64(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(int64(0)))
}

func (b *TypeBuilder) AddFloat64(name string) *TypeBuilder {
	return b.AddField(name, reflect.TypeOf(float64(1.2)))
}

func main() {
	b := NewTypeBuilder().
		AddString("Name").
		AddInt64("Age")

	p := b.New()
	i := b.Interface()
	pp := reflect.ValueOf(i).Elem().Addr().Interface()
	fmt.Printf("typ:%T, value:%+v, ponter1:%d,ponter1:%v\n", p, i, p, pp)
}
