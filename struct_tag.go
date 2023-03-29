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
	"strings"
	"sync"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

//TagInfo 拥有tag的struct的成员的解析结果
type TagInfo struct {
	TagName      string       //
	BaseType     reflect.Type //
	BaseKind     reflect.Kind // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	Offset       uintptr      //偏移量
	TypeSize     int          //
	StringTag    bool         // `json:"field,string"`: 此情形下,需要把struct的int转成json的string
	OmitemptyTag bool         //  `json:"some_field,omitempty"`

	/*
		ChildList： 遍历 map 性能较差，加个 list
	*/
	Children  map[string]*TagInfo
	ChildList []*TagInfo // 遍历的顺序和速度

	cacheType  reflect.Type // pointer 的 cache
	batchCache *BatchObj    // struct 的指针对象组成的类型 的 pool

	sliceCacheType      reflect.Type // 除 cdynamicPool 以外的 slice 缓存; slice 动态增长，与 pointer 不一样
	sliceElemGoType     *GoType
	idxSliceObjPool     uintptr
	idxSlicePointerPool uintptr // ptrDeep > 1 时，需要使用

	fUnm unmFunc
	fM   mFunc

	SPool  sync.Pool // TODO：slice pool 和 store.pool 放在一起吧，通过 id 来获取获取 pool，并把剩余的”垃圾“放回 sync.Pool 中共下次复用
	SPoolN int32

	slicePool sync.Pool // &dynamicPool{} 的 pool，用于批量非配 slice
	// idxStackDynamic uintptr   // 在 store.pool 的 index 文字

	sPooloffset  int32 // slice pool 在 PoolStore的偏移量； TODO
	psPooloffset int32 // pointer slice pool  在 PoolStore的偏移量
	bsMarshalLen int32 // 缓存上次 生成的 bs 的大小，如果 cache 小于这个值，则丢弃
	bsHaftCount  int32 // 记录上次低于 bsMarshalLen/2 的次数
}

const SPoolN = 1024 // * 1024

func (t *TagInfo) GetChildFromMap(key string) *TagInfo {
	return t.Children[string(key)]
}

func (t *TagInfo) AddChild(c *TagInfo) (err error) {
	if len(t.Children) == 0 {
		t.Children = make(map[string]*TagInfo)
	}
	if _, ok := t.Children[c.TagName]; ok {
		err = fmt.Errorf("error, tag[%s]类型配置出错,字段重复", c.TagName)
		return
	}
	t.ChildList = append(t.ChildList, c)
	t.Children[c.TagName] = c
	return
}

// []byte 是一种特殊的底层数据类型，需要 base64 编码
func isBytes(typ reflect.Type) bool {
	bsType := reflect.TypeOf(&[]byte{})
	return UnpackType(bsType).Hash == UnpackType(typ).Hash
}
func isStrings(typ reflect.Type) bool {
	bsType := reflect.TypeOf([]string{})
	return UnpackType(bsType).Hash == UnpackType(typ).Hash
}

type ancestor struct {
	hash uint32
	tag  *TagInfo
}

func (ti *TagInfo) setFuncs(builder, sliceBuilder *TypeBuilder, typ reflect.Type, anonymous bool, ancestors []ancestor) (son *TagInfo, err error) {
	son = ti
	ptrDeep, baseType := 0, typ
	var pidx *uintptr
	for ; ; typ = typ.Elem() {
		if typ.Kind() == reflect.Ptr {
			ptrDeep++
			continue
		}
		baseType = typ
		break
	}
	if ptrDeep > 0 {
		pidx = &[]uintptr{0}[0]
		builder.AppendTagField(baseType, pidx)
	}

	// 先从最后一个基础类型开始处理
	switch baseType.Kind() {
	case reflect.Bool:
		ti.fUnm, ti.fM = boolMFuncs2(pidx)
	case reflect.Uint, reflect.Uint64, reflect.Uintptr:
		ti.fUnm, ti.fM = uint64MFuncs(pidx)
	case reflect.Int, reflect.Int64:
		ti.fUnm, ti.fM = int64MFuncs(pidx)
	case reflect.Uint32:
		ti.fUnm, ti.fM = uint32MFuncs(pidx)
	case reflect.Int32:
		ti.fUnm, ti.fM = int32MFuncs(pidx)
	case reflect.Uint16:
		ti.fUnm, ti.fM = uint16MFuncs(pidx)
	case reflect.Int16:
		ti.fUnm, ti.fM = int16MFuncs(pidx)
	case reflect.Uint8:
		ti.fUnm, ti.fM = uint8MFuncs(pidx)
	case reflect.Int8:
		ti.fUnm, ti.fM = int8MFuncs(pidx)
	case reflect.Float64:
		ti.fUnm, ti.fM = float64MFuncs(pidx)
	case reflect.Float32:
		ti.fUnm, ti.fM = float32MFuncs(pidx)
	case reflect.String:
		ti.fUnm, ti.fM = stringMFuncs2(pidx)
	case reflect.Slice: // &[]byte; Array
		if isBytes(baseType) {
			ti.fUnm, ti.fM = bytesMFuncs(pidx)
		} else {
			ti.BaseType = baseType
			sliceType := baseType.Elem()
			ti.sliceElemGoType = UnpackType(sliceType)
			sliceBuilder.AppendTagField(ti.BaseType, &ti.idxSliceObjPool)
			if ptrDeep > 0 {
				typ := reflect.TypeOf([]unsafe.Pointer{})
				sliceBuilder.AppendTagField(typ, &ti.idxSlicePointerPool)
			}
			if isStrings(baseType) {
				// 字符串数组
				ti.fUnm, ti.fM = sliceStringsMFuncs()
				// ti.fUnm, ti.fM = sliceMFuncs(pidx)
			} else if ptrDeep > 0 {
				//指针数组
				ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else if UnpackType(sliceType).Hash == UnpackType(reflect.TypeOf(int(0))).Hash && ptrDeep == 0 {
				// int 数组
				ti.fUnm, ti.fM = sliceIntsMFuncs(pidx)
				// ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else if UnpackType(sliceType).PtrData == 0 && ptrDeep == 0 {
				ti.fUnm, ti.fM = sliceNoscanMFuncs(pidx)
			} else if UnpackType(sliceType).Hash == UnpackType(reflect.TypeOf(interface{}(0))).Hash && ptrDeep == 0 {
				// interface{} 数组
				ti.fUnm, ti.fM = sliceMFuncs(pidx)
			} else {
				ti.fUnm, ti.fM = sliceMFuncs(pidx)
			}
			son = &TagInfo{
				TagName:  `"son"`,
				BaseType: sliceType,
				TypeSize: int(sliceType.Size()),
			}

			subSon, err := son.setFuncs(builder, sliceBuilder, sliceType, false /*anonymous*/, ancestors)
			if err != nil {
				return nil, lxterrs.Wrap(err, "Struct")
			}
			err = ti.AddChild(subSon)
			if err != nil {
				return nil, lxterrs.Wrap(err, "Struct")
			}

			ti.SPoolN = (1 << 20) / int32(ti.BaseType.Size())
			ti.SPool.New = func() any {
				v := reflect.MakeSlice(ti.BaseType, 0, int(ti.SPoolN)) // SPoolN)
				p := reflectValueToPointer(&v)
				pH := (*SliceHeader)(p)
				pH.Cap = pH.Cap * int(sliceType.Size())
				return (*[]uint8)(p)
			}
		}
	case reflect.Struct:
		var sonIdx uintptr = 0
		ti.fUnm, ti.fM = structMFuncs(pidx, &sonIdx)

		son, err = NewStructTagInfo(baseType, ancestors)
		if err != nil {
			return nil, lxterrs.Wrap(err, "Struct")
		}
		son.fUnm, son.fM = ti.fUnm, ti.fM
		if son != nil && son.cacheType != nil {
			builder.AppendTagField(son.cacheType, &sonIdx) //TODO： 如果是 slice 这里需要处理成 slice 模式
		}
		if son == nil {
			// TODO: fUnm中需要重建 store.pool，并从
			// tag, err := LoadTagNode(vi, goType.Hash) 获取 tag，需要延后处理？
			ti.fUnm, ti.fM = structMFuncs(pidx, &sonIdx)
		}
		// 匿名成员的处理; 这里只能处理费指针嵌入，指针嵌入逻辑在上一层
		if !anonymous {
			if son.sliceCacheType != nil {
				sliceBuilder.AppendTagField(son.sliceCacheType, &ti.idxSliceObjPool)
				// ti.slicePool.New = son.slicePool.New
				// ti.sliceCacheType = son.sliceCacheType
				// ti.batchCache = son.batchCache
				// ti.cacheType = son.cacheType
			}
			for _, c := range son.ChildList {
				err = ti.AddChild(c)
				if err != nil {
					return nil, lxterrs.Wrap(err, "AddChild")
				}
			}
			// ti.buildChildMap()
		} else {
			for _, c := range son.ChildList {
				if ptrDeep == 0 {
					c.Offset += ti.Offset
				} else {
					fUnm, fM := c.fUnm, c.fM
					c.fM = func(store Store, in []byte) (out []byte) {
						store.obj = *(*unsafe.Pointer)(store.obj)
						if store.obj != nil {
							return fM(store, in)
						}
						out = append(in, "null"...)
						return
					}
					c.fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
						store.obj = *(*unsafe.Pointer)(store.obj)
						if store.obj == nil {
							store.obj = store.Idx(*pidx)
						}
						return fUnm(idxSlash, store, stream)
					}
				}
				err = ti.AddChild(c)
				if err != nil {
					return nil, lxterrs.Wrap(err, "AddChild")
				}
			}
		}
	case reflect.Interface:
		// Interface 需要根据实际类型创建
		ti.fUnm, ti.fM = interfaceMFuncs(pidx)

	case reflect.Map:
		ti.fUnm, ti.fM = mapMFuncs(pidx)
		valueType := baseType.Elem()
		son = &TagInfo{
			TagName:  `"son"`,
			TypeSize: int(valueType.Size()), // TODO
			// Builder:  ti.Builder,
		}
		err = ti.AddChild(son)
		if err != nil {
			return nil, lxterrs.Wrap(err, "Struct")
		}
		subSon, err := son.setFuncs(builder, sliceBuilder, valueType, false /*anonymous*/, ancestors)
		if err != nil {
			return nil, lxterrs.Wrap(err, "Struct")
		}
		_ = subSon
	default:
		return nil, lxterrs.New("errors type:%s", baseType)
	}

	// 处理一下指针
	for i := 1; i < ptrDeep; i++ {
		var idxP *uintptr = &[]uintptr{0}[0]
		builder.AppendPointer(fmt.Sprintf("%s_%d", ti.TagName, i), idxP)
		fUnm, fM := ti.fUnm, ti.fM
		ti.fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			store.obj = store.Idx(*idxP)
			return fUnm(idxSlash, store, stream)
		}
		ti.fM = func(store Store, in []byte) (out []byte) {
			store.obj = *(*unsafe.Pointer)(store.obj)
			return fM(store, in)
		}
	}

	return
}

//NewStructTagInfo 解析struct的tag字段，并返回解析的结果
/*
   每个 struct 都搞一个 pointerCacheType，在使用的时候直接获取； 再搞一个 slicePool 在是 slice 时使用；
   二者不会同一时刻出现，是不是可以合并为同一个值？

*/
func NewStructTagInfo(typIn reflect.Type, ancestors []ancestor) (ti *TagInfo, err error) {
	if typIn.Kind() != reflect.Struct {
		err = lxterrs.New("NewStructTagInfo only accepts structs; got %v", typIn.Kind())
		return
	}

	builder := NewTypeBuilder()
	sliceBuilder := NewTypeBuilder()
	ti = &TagInfo{
		BaseType: typIn,
		TagName:  typIn.String(),
		BaseKind: typIn.Kind(), // 解析出最内层类型
		TypeSize: int(typIn.Size()),
	}

	goType := UnpackType(typIn)

	// 通过 ancestors 避免死循环
	for _, a := range ancestors {
		if a.hash == goType.Hash {
			ti = nil // 以返回 nil 来处理后续逻辑
			return   // 避免嵌套循环
		}
	}
	ancestors = append(ancestors, ancestor{goType.Hash, ti})

	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		son := &TagInfo{
			BaseType: field.Type,
			TagName:  `"` + field.Name + `"`,
			Offset:   field.Offset,
			BaseKind: field.Type.Kind(),
			TypeSize: int(field.Type.Size()),
		}

		if !field.IsExported() {
			continue // 非导出成员不处理
		}

		tagv := field.Tag.Get("json")  //从tag列表中取出下标为i的tag //json:"field,string"
		tagv = strings.TrimSpace(tagv) //去除两头的空格
		if len(tagv) > 0 && tagv == "-" {
			continue //如果tag字段没有内容，则不处理
		}
		if len(tagv) == 0 {
			tagv = field.Name // 没有 tag 则以成员名为 tag
		}

		tvs := strings.Split(tagv, ",")
		for i := range tvs {
			tvs[i] = strings.TrimSpace(tvs[i])
		}
		son.TagName = `"` + tvs[0] + `"` // 此处加上 双引号 是为了方便使用 改进后的 hash map
		for i := 1; i < len(tvs); i++ {
			if strings.TrimSpace(tvs[i]) == "string" {
				son.StringTag = true
				continue
			}
			if strings.TrimSpace(tvs[i]) == "omitempty" {
				son.OmitemptyTag = true
				continue
			}
		}

		_, err = son.setFuncs(builder, sliceBuilder, field.Type, field.Anonymous, ancestors)
		if err != nil {
			err = lxterrs.Wrap(err, "son.setFuncs")
			return
		}
		if !field.Anonymous {
			err = ti.AddChild(son)
			if err != nil {
				return
			}
		} else {
			// 如果是匿名成员类型，需要将其子成员插入为父节点的子成员；
			// 此外，get set 函数也要做相应修改
			for _, c := range son.ChildList {
				err = ti.AddChild(c)
				if err != nil {
					return
				}
			}
		}
	}

	ti.cacheType = builder.Build()
	ti.batchCache = NewBatchObj(ti.cacheType)

	ti.sliceCacheType = sliceBuilder.Build()
	if ti.sliceCacheType != nil {
		// batch := NewBatchObj(ti.sliceCacheType)
		goType := UnpackType(ti.sliceCacheType)
		ti.slicePool = sync.Pool{
			New: func() any {
				return unsafe_NewArray(goType, 1)
				// return batch.Get()
			},
		}
	}
	return
}
