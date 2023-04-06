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
	// 常用的放前面，在缓存的概率大
	fUnm       unmFunc
	fM         mFunc
	sliceCache *BatchObj

	TagName      string       //
	BaseType     reflect.Type //
	BaseKind     reflect.Kind // 次成员可能是 **string,[]int 等这种复杂类型,这个 用来指示 "最里层" 的类型
	Offset       uintptr      //偏移量
	TypeSize     int          //
	StringTag    bool         // `json:"field,string"`: 此情形下,需要把struct的int转成json的string
	OmitemptyTag bool         //  `json:"some_field,omitempty"`

	Children  map[string]*TagInfo // 支持快速获取子节点
	ChildList []*TagInfo          // 支持遍历的顺序；加快遍历速度

	ptrCacheType reflect.Type // pointer 的 cache
	ptrCache     *BatchObj    // ptrCacheType 类型的 pool

	slicePoolType       reflect.Type // 除 cdynamicPool 以外的 slice 缓存; slice 动态增长，与 pointer 不一样
	sliceElemGoType     *GoType
	idxSliceObjPool     uintptr
	idxSlicePointerPool uintptr // ptrDeep > 1 时，需要使用

	SPool  sync.Pool // TODO：slice pool 和 store.pool 放在一起吧，通过 id 来获取获取 pool，并把剩余的”垃圾“放回 sync.Pool 中共下次复用
	SPoolN int32
	SPool2 *BatchObj // 新的 slice cache

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

//AddChild 添加下级子节点
func (t *TagInfo) AddChild(c *TagInfo) (err error) {
	if len(t.Children) == 0 {
		t.Children = make(map[string]*TagInfo)
	}
	if _, ok := t.Children[c.TagName]; ok {
		err = fmt.Errorf("error, type[%s].tag[%s]类型配置出错,字段重复", t.TagName, c.TagName)
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

func (ti *TagInfo) setFuncs(ptrBuilder, sliceBuilder *TypeBuilder, typ reflect.Type, anonymous bool, ancestors []ancestor) (son *TagInfo, err error) {
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
	son.BaseType = baseType
	if ptrDeep > 0 {
		pidx = &[]uintptr{0}[0]
		ptrBuilder.AppendTagField(baseType, pidx)
	}

	// 先从最后一个基础类型开始处理
	switch baseType.Kind() {
	case reflect.Bool:
		ti.fUnm, ti.fM = boolMFuncs(pidx)
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

			subSon, err := son.setFuncs(ptrBuilder, sliceBuilder, sliceType, false /*anonymous*/, ancestors)
			if err != nil {
				return nil, lxterrs.Wrap(err, "Struct")
			}
			subSon.sliceCache = NewBatchObj(sliceType)

			err = ti.AddChild(subSon) //TODO: err = ti.AddChild(son) ?
			// err = ti.AddChild(son)
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
			ti.SPool2 = NewBatchObj(subSon.BaseType)
		}
	case reflect.Struct:
		var sonIdx uintptr = 0
		ti.fUnm, ti.fM = structMFuncs(pidx, &sonIdx)

		son, err = NewStructTagInfo(baseType, ancestors)
		// goType := UnpackType(baseType)
		// son, err = LoadTagNodeByType(baseType, goType.Hash)
		if err != nil {
			return nil, lxterrs.Wrap(err, "Struct")
		}
		son.fUnm, son.fM = ti.fUnm, ti.fM
		if son != nil && son.ptrCacheType != nil {
			ptrBuilder.AppendTagField(son.ptrCacheType, &sonIdx) //TODO： 如果是 slice 这里需要处理成 slice 模式
		}
		if son == nil {
			// TODO: fUnm中需要重建 store.pool，并从
			// tag, err := LoadTagNode(vi, goType.Hash) 获取 tag，需要延后处理？
			ti.fUnm, ti.fM = structMFuncs(pidx, &sonIdx)
		}
		// 匿名成员的处理; 这里只能处理费指针嵌入，指针嵌入逻辑在上一层
		if !anonymous {
			if son.slicePoolType != nil {
				sliceBuilder.AppendTagField(son.slicePoolType, &ti.idxSliceObjPool)
				// ti.slicePool.New = son.slicePool.New
				// ti.slicePoolType = son.slicePoolType
				// ti.ptrCache = son.ptrCache
				// ti.ptrCacheType = son.ptrCacheType
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
		subSon, err := son.setFuncs(ptrBuilder, sliceBuilder, valueType, false /*anonymous*/, ancestors)
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
		ptrBuilder.AppendPointer(fmt.Sprintf("%s_%d", ti.TagName, i), idxP)
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

	ptrBuilder := NewTypeBuilder()
	sliceBuilder := NewTypeBuilder()
	ti = &TagInfo{
		BaseType: typIn,
		TagName:  typIn.String(),
		BaseKind: typIn.Kind(), // 解析出最内层类型
		TypeSize: int(typIn.Size()),
	}

	// 通过 ancestors 避免死循环
	goType := UnpackType(typIn)
	isNestedLoop := false // 是否嵌套循环
	for _, a := range ancestors {
		if a.hash == goType.Hash {
			ti = nil // 以返回 nil 来处理后续逻辑
			panic("Nested loops are not yet supported")
			return // 避免嵌套循环
			isNestedLoop = true
			_ = isNestedLoop
			break
			/*
				// TODO: 针对循环类型
				// fUnm 和 fM 里重新创建缓存和对象，再获取 tag 继续往下执行
				store := PoolStore{
						tag:         tag,
						obj:         prv.ptr, // eface.Value,
						pointerPool: tag.ptrCache.Get(),
					}
					//slice 才需要的缓存
					if tag.slicePool.New != nil {
						store.slicePool = tag.slicePool.Get().(unsafe.Pointer)
						store.dynPool = dynPool.Get().(*dynamicPool)

						err = parseRoot(bs[i:], store)

						tag.slicePool.Put(store.slicePool)
						dynPool.Put(store.dynPool)
					} else {
						err = parseRoot(bs[i:], store)
					}
			*/
		}
	}
	ancestors = append(ancestors, ancestor{goType.Hash, ti})

	// 解析 struct 成员类型
	for i := 0; i < typIn.NumField(); i++ {
		field := typIn.Field(i)
		son := &TagInfo{
			BaseType: field.Type,
			Offset:   field.Offset,
			BaseKind: field.Type.Kind(),
			TypeSize: int(field.Type.Size()),
			TagName:  field.Name,
			// TagName:  `"` + field.Name + `"`,
		}

		if !field.IsExported() {
			continue // 非导出成员不处理
		}
		fieldTag := newtag(field.Tag.Get("json")) //从tag列表中取出下标为i的tag //json:"field,string"
		if fieldTag.negligible() {
			continue
		}
		if !fieldTag.empty() {
			son.TagName, son.StringTag, son.OmitemptyTag = fieldTag.parse()
		}

		_, err = son.setFuncs(ptrBuilder, sliceBuilder, field.Type, field.Anonymous, ancestors)
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

	// 缓存处理
	ti.ptrCacheType = ptrBuilder.Build()
	ti.ptrCache = NewBatchObj(ti.ptrCacheType)

	ti.slicePoolType = sliceBuilder.Build()
	if ti.slicePoolType != nil {
		// batch := NewBatchObj(ti.slicePoolType)
		goType := UnpackType(ti.slicePoolType)
		// TODO: slicePool 忘了怎么实际的了，但是看上去没起作用，需要重新设计
		// 实际上：如果每次分配 slice 都从共的 slice 中分配 2 倍空间，
		// 只是空间利用率可能要点，但是最多也就低一半而已，看上去完全可以接受
		ti.slicePool = sync.Pool{
			New: func() any {
				return unsafe_NewArray(goType, 1)
				// return batch.Get()
			},
		}
	}
	return
}

type tag string

func newtag(tagv string) tag {
	tagv = strings.TrimSpace(tagv)
	return tag(tagv)
}

// 应该忽视的
func (t tag) negligible() bool {
	if t == "-" {
		return true
	}
	return false
}

func (t tag) empty() bool {
	if t == "" {
		return true
	}
	return false
}

func (t tag) parse() (name string, bString, bOmitempty bool) {
	tvs := strings.Split(string(t), ",")
	// name = `"` + strings.TrimSpace(tvs[0]) + `"` // 此处加上 双引号 是为了方便使用 改进后的 hash map
	name = strings.TrimSpace(tvs[0]) // 此处加上 双引号 是为了方便使用 改进后的 hash map

	for i := 1; i < len(tvs); i++ {
		v := strings.TrimSpace(tvs[i])
		if v == "string" {
			bString = true
			continue
		}
		if v == "omitempty" {
			bOmitempty = true
			continue
		}
	}
	return
}
