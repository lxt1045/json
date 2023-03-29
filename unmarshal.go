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
	"errors"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

//go:noinline
func ErrStream(stream string) string {
	if len(stream[:]) > 128 {
		stream = stream[:128]
	}
	str := string(stream)
	return str
}

var spaceTable = [256]bool{
	'\t': true, '\n': true, '\v': true, '\f': true, '\r': true, ' ': true, 0x85: true, 0xA0: true,
}

func trimSpace(stream string) (i int) {
	for ; spaceTable[stream[i]]; i++ {
	}
	return
}

// 为了 inline 部分共用逻辑让调用者完成; 逻辑 解析：冒号 和 逗号 等单字符
// n 表示在空字符串中找到多少个 b
func parseByte(stream string, b byte) (i, n int) {
	for ; ; i++ {
		if stream[i] == b {
			n++
			continue
		}
		if !spaceTable[stream[i]] {
			return
		}
	}
}
func parseByte0(stream string, b byte) (i, n int) {
loop:
	if stream[i] == b {
		n++
		i++
		goto loop
	}
	if !spaceTable[stream[i]] {
		return
	}
	i++
	goto loop
}

func parseObjToSlice(stream string, s []interface{}) (i int) {
	return 0
}

// 解析 {}
// func parseObj(sts status, stream string, store PoolStore,  tag *TagInfo) (i int) {
func parseObj(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i += trimSpace(stream[i:])
	if stream[i] == '}' {
		i++
		return
	}
	n, nB := 0, 0
	key := ""
	for {
		// 手动内联
		{
			start := i
			n = strings.IndexByte(stream[i+1:], '"') // 默认 stream[i] == '"'， 不做检查
			if n >= 0 {
				i += n + 2
				key = stream[start:i]
			}
		}
		// 解析 冒号
		n, nB = parseByte(stream[i:], ':')
		i += n
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		son := store.tag.Children[string(key)]

		if son != nil {
			storeSon := PoolStore{
				tag:         son,
				pointerPool: store.pointerPool,
				slicePool:   store.slicePool,
				dynPool:     store.dynPool,
				obj:         store.obj,
			}
			n, iSlash = son.fUnm(iSlash-i, storeSon, stream[i:])
			iSlash += i
		} else {
			n = parseEmpty(stream[i:])
		}
		i += n
		// 解析 逗号
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

func parseMapInterface(idxSlash int, stream string) (m map[string]interface{}, i, iSlash int) {
	iSlash = idxSlash
	n, nB := 0, 0
	key := ""
	ppairs := pairPool.Get().(*[]pair)
	pairs := *ppairs
	for {
		i += trimSpace(stream[i:])
		// 手动内联
		{
			i++
			n = strings.IndexByte(stream[i:], '"')
			if n >= 0 {
				n += i
				key = stream[i:n]
				i = n + 1
			}
		}
		n, nB = parseByte(stream[i:], ':')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
		pairs = append(pairs, pair{
			k: key,
		})
		n, iSlash = parseInterface(iSlash-i, stream[i:], &pairs[len(pairs)-1].v)
		iSlash += i
		i += n
		// m[string(key)] = *value
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++

				// map
				// m = make(map[string]interface{}, len(pairs))
				m = makeMapEface(len(pairs))

				for i := range pairs {
					m[*(*string)(unsafe.Pointer(&pairs[i].k))] = pairs[i].v
				}
				*ppairs = pairs[:0]
				pairPool.Put(ppairs)
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

// map[string]T
func parseMapValue(idxSlash int, stream string) (m map[string]interface{}, i, iSlash int) {
	iSlash = idxSlash
	n, nB := 0, 0
	key := ""
	ppairs := pairPool.Get().(*[]pair)
	pairs := *ppairs
	for {
		i += trimSpace(stream[i:])
		// 手动内联
		{
			i++
			n = strings.IndexByte(stream[i:], '"')
			if n >= 0 {
				n += i
				key = stream[i:n]
				i = n + 1
			}
		}
		n, nB = parseByte(stream[i:], ':')
		if nB != 1 {
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
		i += n
		pairs = append(pairs, pair{
			k: key,
		})
		n, iSlash = parseInterface(iSlash-i, stream[i:], &pairs[len(pairs)-1].v)
		iSlash += i
		i += n
		// m[string(key)] = *value
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && '}' == stream[i] {
				i++

				// map
				// m = make(map[string]interface{}, len(pairs))
				m = makeMapEface(len(pairs))

				for i := range pairs {
					m[*(*string)(unsafe.Pointer(&pairs[i].k))] = pairs[i].v
				}
				*ppairs = pairs[:0]
				pairPool.Put(ppairs)
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

func parseSliceInterface(idxSlash int, stream string) (s []interface{}, i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	var value interface{}
	s = poolSliceInterface.Get().([]interface{})
	for n, nB := 0, 0; ; {
		n, iSlash = parseInterface(iSlash-i, stream[i:], &value)
		iSlash += i
		i += n
		s = append(s, value)
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				if cap(s)-len(s) > 4 {
					sLeft := s[len(s):]
					poolSliceInterface.Put(sLeft)
					s = s[:len(s):len(s)]
				}
				return
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
}

//parseSlice 可以细化一下，每个类型来一个，速度可以加快
func parseSlice2(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream)
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	son := store.tag.ChildList[0]
	size := son.TypeSize
	tag := store.tag
	uint8s := tag.SPool.Get().(*[]uint8) // cpu %12; parseSlice, cpu 20%
	pHeader := (*SliceHeader)(store.obj)
	bases := (*[]uint8)(store.obj)
	SPoolN, BaseType := store.tag.SPoolN, store.tag.BaseType
	store.tag = son
	for n, nB := 0, 0; ; {
		if len(*uint8s)+size > cap(*uint8s) {
			l := cap(*uint8s) / size
			c := l * 2
			if c < int(SPoolN) {
				c = int(SPoolN)
			}
			v := reflect.MakeSlice(BaseType, 0, c)
			p := reflectValueToPointer(&v)
			news := (*[]uint8)(p)

			pH := (*SliceHeader)(p)
			pH.Cap = pH.Cap * size
			// copy(*news, *uint8s)
			// *uint8s = *news
			*uint8s = append((*news)[:0], *uint8s...)
		}

		l := len(*uint8s)
		*uint8s = (*uint8s)[:l+size]

		p := unsafe.Pointer(&(*uint8s)[l])
		store.obj = p
		n, iSlash = son.fUnm(iSlash-i, store, stream[i:])
		iSlash += i
		// if n == 0 {
		// 	pHeader.Len -= size
		// }
		i += n
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}

	*bases = (*uint8s)[:len(*uint8s):len(*uint8s)]
	if cap(*uint8s)-len(*uint8s) > 16*size {
		*uint8s = (*uint8s)[len(*uint8s):]
		tag.SPool.Put(uint8s)
	}
	// pH.Data = uintptr(pointerOffset(unsafe.Pointer(pHeader.Data), uintptr(pHeader.Len)))
	// pH.Cap = pHeader.Cap - pHeader.Len
	pHeader.Len = pHeader.Len / size
	pHeader.Cap = pHeader.Cap / size

	return
}

//parseSlice 可以细化一下，每个类型来一个，速度可以加快
func parseSlice(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream)
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	son := store.tag.ChildList[0]
	size := son.TypeSize

	// TODO : 从 store.pool 获取 pool
	// uint8s := store.tag.SPool.Get().(*[]uint8) // cpu %12; parseSlice, cpu 20%

	// uint8s := store.GetObjs(store.tag.idxSliceObjPool, store.tag.BaseType)
	uint8s := store.GetObjs(store.tag.sliceElemGoType)

	pHeader := (*SliceHeader)(store.obj)
	bases := (*[]uint8)(store.obj)
	// TODO
	// slicePool := son.slicePool.Get().(unsafe.Pointer)
	sliceElemGoType := store.tag.sliceElemGoType
	store.tag = son
	for n, nB := 0, 0; ; {
		uint8s = GrowObjs(uint8s, size, sliceElemGoType)
		store.obj = unsafe.Pointer(&uint8s[len(uint8s)-size])
		n, iSlash = son.fUnm(iSlash-i, store, stream[i:])
		iSlash += i
		i += n
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}

	*bases = (uint8s)[:len(uint8s):len(uint8s)]
	if cap(uint8s)-len(uint8s) > 4*size {
		uint8s = uint8s[len(uint8s):]
		// store.SetObjs(uint8s, store.tag.idxSliceObjPool)
		store.SetObjs(uint8s)
	}
	pHeader.Len = pHeader.Len / size
	pHeader.Cap = pHeader.Cap / size

	return
}

//parseNoscanSlice 解析没有 pointer 的 slice，分配内存是不需要标注指针
func parseNoscanSlice(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream)
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	son := store.tag.ChildList[0]
	size := son.TypeSize
	bytes := store.GetNoscan()
	for n, nB := 0, 0; ; {
		l := len(bytes)
		bytes = GrowBytes(bytes, size)
		p := unsafe.Pointer(&bytes[l])
		n, iSlash = son.fUnm(iSlash-i, PoolStore{
			obj:         p,
			tag:         son,
			pointerPool: store.pointerPool,
			slicePool:   store.slicePool,
			dynPool:     store.dynPool,
		}, stream[i:])
		iSlash += i
		i += n
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}

	store.SetNoscan(bytes[len(bytes):])
	l := len(bytes) / size
	*(*SliceHeader)(store.obj) = SliceHeader{
		Len:  l,
		Cap:  l,
		Data: unsafe.Pointer(&bytes[0]),
	}
	return
}

//parseNoscanSlice 解析没有 pointer 的 slice，分配内存是不需要标注指针
func parseIntSlice(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream)
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	ints := store.GetInts()
	for n, nB := 0, 0; ; {
		l := len(ints)
		ints = GrowInts(ints)
		p := unsafe.Pointer(&ints[l])
		{
			bs := stream[i:]
			for n = 0; n < len(bs); n++ {
				c := bs[n]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			num, err := strconv.ParseInt(bs[:n], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(bs[:n]))
				panic(err)
			}
			*(*int64)(p) = num
		}
		i += n
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}

	store.SetInts(ints[len(ints):])
	*(*[]int)(store.obj) = ints[:len(ints):len(ints)]
	return
}

//quadwords 4word: 64bit; d：doubleword，双字，32位; w：word，双字节，字，16位; b：byte，字节，8位
// tag 实际上已经可以提前知道了，这里无需再取一次，重复了
func parseSliceString1(idxSlash int, stream string, store PoolStore, SPoolN int, strsPool *sync.Pool) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	// TODO 使用 IndexByte 先计算slice 的长度，在分配内存
	// pstrs := strsPool.Get().(*[]string)
	// strs = strs[:0:cap(strs)]
	strs := (*[1 << 20]string)(unsafe.Pointer(strPool.GetN(4)))[:0:4] //make([]string, 0, 4)
	pstrs := (*[]string)(store.obj)
	*pstrs = strs
	for n, nB := 0, 0; ; {
		if len(*pstrs)+1 > cap(*pstrs) {
			c := len(*pstrs) * 2
			news := (*[1 << 20]string)(unsafe.Pointer(strPool.GetN(c)))[:0:c]
			*pstrs = append(news, *pstrs...)
		}
		*pstrs = (*pstrs)[:len(*pstrs)+1]
		// son := store.tag.ChildList[0]
		// n, iSlash = son.fUnm(iSlash-i, PoolStore{
		// 	obj:  unsafe.Pointer(&(*pstrs)[len(*pstrs)-1]),
		// 	tag:  son,
		// 	pool: store.pool,
		// }, stream[i:])
		// iSlash += i
		// i += n
		{
			// 全部内联
			i++
			n := strings.IndexByte(stream[i:], '"')
			n += i
			if iSlash > n {
				(*pstrs)[len(*pstrs)-1] = stream[i:n]
				i = n + 1
			} else {
				(*pstrs)[len(*pstrs)-1], n, iSlash = parseUnescapeStr(stream[i:], n-i, iSlash)
				iSlash += i
				i = i + n
			}
		}
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
	return
}
func parseSliceString(idxSlash int, stream string, store PoolStore) (i, iSlash int) {
	iSlash = idxSlash
	i = trimSpace(stream[i:])
	if stream[i] == ']' {
		i++
		pHeader := (*SliceHeader)(store.obj)
		pHeader.Data = store.obj
		return
	}
	strs := store.GetStrings()
	for n, nB := 0, 0; ; {
		strs = GrowStrings(strs, 1)
		{
			// 全部内联
			i++
			n := strings.IndexByte(stream[i:], '"')
			n += i
			if iSlash > n {
				(strs)[len(strs)-1] = stream[i:n]
				i = n + 1
			} else {
				(strs)[len(strs)-1], n, iSlash = parseUnescapeStr(stream[i:], n-i, iSlash)
				iSlash += i
				i = i + n
			}
		}
		n, nB = parseByte(stream[i:], ',')
		i += n
		if nB != 1 {
			if nB == 0 && ']' == stream[i] {
				i++
				break
			}
			panic(lxterrs.New(ErrStream(stream[i:])))
		}
	}
	store.SetStrings(strs[len(strs):])
	*(*[]string)(store.obj) = strs[:len(strs):len(strs)]
	return
}

// key 后面的单元: Num, str, bool, slice, obj, null
func parseInterface(idxSlash int, stream string, p *interface{}) (i, iSlash int) {
	iSlash = idxSlash
	// i = trimSpace(stream)
	switch stream[0] {
	default: // num
		var f float64
		f, i = float64UnmFuncs(stream)
		*p = f
	case '{': // obj
		var m map[string]interface{} // TODO： m 逃逸了
		m, i, iSlash = parseMapInterface(iSlash-1, stream[1:])
		iSlash++
		i++
		*p = m
	case '}':
	case '[': // slice
		var s []interface{}
		s, i, iSlash = parseSliceInterface(iSlash-1, stream[1:])
		iSlash++
		i++
		// *p = s
		ps := islicePool.Get()
		*ps = s
		pEface := (*GoEface)(unsafe.Pointer(p))
		pEface.Type = isliceEface.Type
		pEface.Value = unsafe.Pointer(ps)
	case ']':
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
		*p = true
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 5
		*p = false
	case '"':
		var raw string
		//
		raw, i, iSlash = parseStr(stream, iSlash)
		// *p = bytesString(raw)
		// return

		// pstr := strPool.Get() //
		pstr := BatchGet(strPool) //
		// bytesCopyToString(raw, pstr)
		*pstr = *(*string)(unsafe.Pointer(&raw))
		pEface := (*GoEface)(unsafe.Pointer(p))
		pEface.Type = strEface.Type
		pEface.Value = unsafe.Pointer(pstr)
	}
	return
}

func parseEmptyObjSlice(stream string, bLeft, bRight byte) (i int) {
	indexQuote := func(stream string, i int) int {
		for {
			iDQuote := strings.IndexByte(stream[i:], '"')
			if iDQuote < 0 {
				return math.MaxInt32
			}
			i += iDQuote // 指向 '"'
			if stream[i-1] != '\\' {
				return i
			}
			j := i - 2
			for ; stream[j] == '\\'; j-- {
			}
			if (i-j)%2 == 0 {
				i++
				continue
			}
			return i
		}
	}
	i++
	nBrace := 0                                     // " 和 {
	iBraceL := strings.IndexByte(stream[i:], bLeft) //通过 ’“‘ 的 idx 来确定'{' '}' 是否在字符串中
	iBraceR := strings.IndexByte(stream[i:], bRight)
	if iBraceL < 0 {
		iBraceL = math.MaxInt32 // 保证 +i 后不会溢出
	}
	if iBraceR < 0 {
		iBraceR = math.MaxInt32
	}
	iBraceL, iBraceR = iBraceL+i, iBraceR+i

	iDQuoteL := indexQuote(stream, i)
	iDQuoteR := indexQuote(stream, iDQuoteL+1)

	for {
		// 1. 以 iBraceR 为边界
		if iBraceR < iBraceL {
			if iDQuoteR < iBraceR {
				// ']'在右区间
				iDQuoteL = indexQuote(stream, iDQuoteR+1)
				iDQuoteR = indexQuote(stream, iDQuoteL+1)
				continue
			} else if iBraceR < iDQuoteL {
				// ']'在左区间
				if nBrace == 0 {
					i = iBraceR + 1
					return
				}
				nBrace--
				iBraceR++
				iBraceRNew := strings.IndexByte(stream[iBraceR:], bRight)
				if iBraceRNew < 0 {
					iBraceRNew = math.MaxInt32
				}
				iBraceR += iBraceRNew
				continue
			} else {
				// ']'在中间区间
				iBraceR = strings.IndexByte(stream[iDQuoteR:], bRight)
				if iBraceR < 0 {
					iBraceR = math.MaxInt32
				}
				iBraceR += iDQuoteR
				continue
			}
		} else {
			// iBraceL < iBraceR
			// 2. 以 iBraceR 为边界

			if iDQuoteR < iBraceL {
				// ']'在右区间
				iDQuoteL = indexQuote(stream, iDQuoteR+1)
				iDQuoteR = indexQuote(stream, iDQuoteL+1)
				continue
			} else if iBraceL < iDQuoteL {
				// ']'在左区间
				nBrace++
				iBraceL++
				iBraceLNew := strings.IndexByte(stream[iBraceL:], bLeft) //通过 ’“‘ 的 idx 来确定'{' '}' 是否在字符串中
				if iBraceLNew < 0 {
					iBraceLNew = math.MaxInt32 // 保证 +i 后不会溢出
				}
				iBraceL += iBraceLNew
				continue
			} else {
				// ']'在中间区间
				iBraceL = strings.IndexByte(stream[iDQuoteR:], bLeft)
				if iBraceL < 0 {
					iBraceL = math.MaxInt32
				}
				iBraceL += iDQuoteR
				continue
			}
		}
	}
	return
}

//TODO 通过 IndexByte 的方式快速跳过； 在下一层处理，这里 设为 nil
// 如果是 其他： 找 ','
// 如果是obj: 1. 找 ’}‘; 2. 找'{'； 3. 如果 2 比 1 小则循环 1 2
// 如果是 slice : 1. 找 ’]‘; 2. 找'['； 3. 如果 2 比 1 小则循环 1 2
// var iface interface{}
// n, iSlash = parseInterface(iSlash-i, stream[i:], &iface)
// iSlash += i
func parseEmpty(stream string) (i int) {
	switch stream[0] {
	default: // num
		for ; i < len(stream); i++ {
			c := stream[i]
			if c == ']' || c == '}' || c == ',' {
				break
			}
		}
	case '{': // obj
		n := parseEmptyObjSlice(stream[i:], '{', '}')
		i += n
	case '[': // slice
		n := parseEmptyObjSlice(stream[i:], '[', ']')
		i += n
	case ']', '}':
	case 'n':
		if stream[i+1] != 'u' || stream[i+2] != 'l' || stream[i+3] != 'l' {
			err := lxterrs.New("should be \"null\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
	case 't':
		if stream[i+1] != 'r' || stream[i+2] != 'u' || stream[i+3] != 'e' {
			err := lxterrs.New("should be \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 4
	case 'f':
		if stream[i+1] != 'a' || stream[i+2] != 'l' || stream[i+3] != 's' || stream[i+4] != 'e' {
			err := lxterrs.New("should be \"false\", not [%s]", ErrStream(stream))
			panic(err)
		}
		i = 5
	case '"':
		i++
		for {
			iDQuote := strings.IndexByte(stream[i:], '"')
			i += iDQuote // 指向 '"'

			// 处理转义字符串
			if stream[i-1] == '\\' {
				j := i - 2
				for ; stream[j] == '\\'; j-- {
				}
				if (i-j)%2 == 0 {
					i++
					continue
				}
			}
			i++
			return
		}
	}
	return
}

//解析 obj: {}, 或 []
func parseRoot(stream string, store PoolStore) (err error) {
	idxSlash := strings.IndexByte(stream[1:], '\\')
	if idxSlash < 0 {
		idxSlash = math.MaxInt
	}
	if stream[0] == '{' {
		parseObj(idxSlash, stream[1:], store)
		return
	}
	return
}

func parseStr(stream string, nextSlashIdx int) (raw string, i, nextSlashIdxOut int) {
	i = strings.IndexByte(stream[1:], '"')
	if i >= 0 && nextSlashIdx > i+1 {
		i++
		raw = stream[1:i]
		i++
		nextSlashIdxOut = nextSlashIdx
		return
	}
	i++
	return parseUnescapeStr(stream, i, nextSlashIdx)
}

func parseUnescapeStr(stream string, nextQuotesIdx, nextSlashIdxIn int) (raw string, i, nextSlashIdx int) {
	nextSlashIdx = nextSlashIdxIn
	if nextSlashIdx < 0 {
		nextSlashIdx = strings.IndexByte(stream[1:], '\\')
		if nextSlashIdx < 0 {
			nextSlashIdx = math.MaxInt
			i += nextQuotesIdx
			raw = stream[1:i]
			i++
			return
		}

		nextSlashIdx++
		// 处理 '\"'
		for {
			i += nextQuotesIdx // 指向 '"'
			if stream[i-1] == '\\' {
				j := i - 2
				for ; stream[j] == '\\'; j-- {
				}
				if (i-j)%2 == 0 {
					i++
					nextQuotesIdx = strings.IndexByte(stream[i:], '"')
					continue
				}
			}
			i++
			break
		}
	}
	if nextQuotesIdx < 0 {
		panic(string(stream[i:]))
	}
	if nextSlashIdx > nextQuotesIdx {
		i = nextQuotesIdx + 1
		raw = stream[:i]
		return
	}
	lastIdx := 0
	var bs []byte
	for {
		i = nextSlashIdx
		word, wordSize := unescapeStr(stream[i:])
		if len(bs) == 0 {
			bs = make([]byte, 0, nextQuotesIdx)
			bs = append(bs[:0], stream[1:i]...) //新建 string 避免修改员 stream
		} else if lastIdx < i {
			bs = append(bs, stream[lastIdx:i]...)
		}
		bs = append(bs, word...)
		i += wordSize
		lastIdx = i
		if word[0] == '"' {
			nextQuotesIdx = strings.IndexByte(stream[i:], '"')
			if nextQuotesIdx < 0 {
				panic(string(stream[i:]))
			}
			nextQuotesIdx += i
		}

		nextSlashIdx = strings.IndexByte(stream[i:], '\\')
		if nextSlashIdx < 0 {
			nextSlashIdx = math.MaxInt
			break
		}
		nextSlashIdx += i
		if nextSlashIdx > nextQuotesIdx {
			break
		}
	}
	bs = append(bs, stream[lastIdx:nextQuotesIdx]...)
	return bytesString(bs), nextQuotesIdx + 1, nextSlashIdx
}

// unescape unescapes a string
//“\\”、“\"”、“\/”、“\b”、“\f”、“\n”、“\r”、“\t”
// \u后面跟随4位16进制数字: "\uD83D\uDE02"
func unescapeStr(raw string) (word []byte, size int) {
	// i==0是 '\\', 所以从1开始
	switch raw[1] {
	case '\\':
		word, size = []byte{'\\'}, 2
	case '/':
		word, size = []byte{'/'}, 2
	case 'b':
		word, size = []byte{'\b'}, 2
	case 'f':
		word, size = []byte{'\f'}, 2
	case 'n':
		word, size = []byte{'\n'}, 2
	case 'r':
		word, size = []byte{'\r'}, 2
	case 't':
		word, size = []byte{'\t'}, 2
	case '"':
		word, size = []byte{'"'}, 2
	case 'u':
		//\uD83D
		if len(raw) < 6 {
			panic(errors.New("incorrect format: \\" + string(raw)))
		}
		last := raw[:6]
		r0 := unescapeToRune(last[2:])
		size, raw = 6, raw[6:]
		if utf16.IsSurrogate(r0) { // 如果utf-6还有后续(不完整)
			if len(raw) < 6 || raw[0] != '\\' || raw[1] != 'u' {
				l := 6
				if l > len(raw) {
					l = len(raw)
				}
				panic(errors.New("incorrect format: \\" + string(last) + string(raw[:l])))
			}
			r1 := unescapeToRune(raw[:6])
			// we expect it to be correct so just consume it
			r0 = utf16.DecodeRune(r0, r1)
			size += 6
		}
		// provide enough space to encode the largest utf8 possible
		word = make([]byte, 4)
		n := utf8.EncodeRune(word, r0)
		word = word[:n]
	default:
		panic(errors.New("incorrect format: " + ErrStream(raw)))
	}
	return
}

// runeit returns the rune from the the \uXXXX
func unescapeToRune(raw string) rune {
	n, err := strconv.ParseUint(string(raw), 16, 64)
	if err != nil {
		panic(errors.New("err:" + err.Error() + ",ncorrect format: " + string(raw)))
	}
	return rune(n)
}
