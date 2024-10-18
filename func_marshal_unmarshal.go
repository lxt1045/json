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
	"encoding/base64"
	"strconv"
	"strings"
	"unsafe"

	lxterrs "github.com/lxt1045/errors"
)

type unmFunc = func(idxSlash int, store PoolStore, stream string) (i, iSlash int)
type mFunc = func(store Store, in []byte) (out []byte)

func pointerOffset(p unsafe.Pointer, offset uintptr) (pOut unsafe.Pointer) {
	return unsafe.Pointer(uintptr(p) + uintptr(offset))
}

func bytesSet(store PoolStore, bs string) (pBase unsafe.Pointer) {
	pBase = store.obj
	pbs := (*[]byte)(store.obj)
	// *pbs = make([]byte, len(bs)*2)
	// n, err := base64.StdEncoding.Decode(*pbs, stringBytes(bs))
	var err error
	*pbs, err = base64.StdEncoding.DecodeString(bs)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(bs))
		return
	}
	// *pbs = (*pbs)[:n]
	return
}
func bytesGet(store Store, in []byte) (out []byte) {
	pObj := store.obj
	bs := *(*[]byte)(pObj)
	l, need := len(in), base64.StdEncoding.EncodedLen(len(bs))
	if l+need > cap(in) {
		//没有足够空间
		in = append(in, make([]byte, need)...)
	}
	base64.StdEncoding.Encode(in[l:l+need], bs)
	out = in[:l+need]
	return
}

func boolMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			if stream[0] == 't' && stream[1] == 'r' && stream[2] == 'u' && stream[3] == 'e' {
				i = 4
				*(*bool)(store.obj) = true
			} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
				i = 5
				*(*bool)(store.obj) = false
			} else {
				err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(stream))
				panic(err)
			}
			return
		}
		// fM = boolGet
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			if *(*bool)(store.obj) {
				out = append(in, "true"...)
			} else {
				out = append(in, "false"...)
			}
			return
		}
		return
	}

	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		if stream[0] == 't' && stream[1] == 'r' && stream[2] == 'u' && stream[3] == 'e' {
			i = 4
			store.obj = store.Idx(*pidx)
			*(*bool)(store.obj) = true
		} else if stream[0] == 'f' && stream[1] == 'a' && stream[2] == 'l' && stream[3] == 's' && stream[4] == 'e' {
			i = 5
			store.obj = store.Idx(*pidx)
			*(*bool)(store.obj) = false
		} else if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			return
		} else {
			err := lxterrs.New("should be \"false\" or \"true\", not [%s]", ErrStream(stream))
			panic(err)
		}
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		pObj := *(*unsafe.Pointer)(store.obj)
		if pObj == nil {
			out = append(in, "null"...)
		} else if *(*bool)(pObj) {
			out = append(in, "true"...)
		} else {
			out = append(in, "false"...)
		}
		return
	}
	return
}

func float64UnmFuncs(stream string) (f float64, i int) {
	for ; i < len(stream); i++ {
		c := stream[i]
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			break
		}
	}
	f, err := strconv.ParseFloat(stream[:i], 64)
	if err != nil {
		err = lxterrs.Wrap(err, ErrStream(stream[:i]))
		panic(err)
	}
	return
}

func uint64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint64)(store.obj) = num
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint64)(store.obj)
			out = strconv.AppendUint(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint64)(store.obj) = num
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint64)(store.obj)
		out = strconv.AppendUint(in, num, 10)
		return
	}
	return
}

func int64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				return
			}
			num, i := ParseInt(stream)
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			*(*int64)(store.obj) = num
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int64)(store.obj)
			out = strconv.AppendInt(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		num, i := ParseInt(stream)
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		*(*int64)(store.obj) = num
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int64)(store.obj)
		out = strconv.AppendInt(in, num, 10)
		return
	}
	return
}

func uint32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint32)(store.obj) = uint32(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint32)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint32)(store.obj) = uint32(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint32)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			num, i := ParseInt(stream)
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			*(*int32)(store.obj) = int32(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int32)(store.obj)
			out = strconv.AppendInt(in, int64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		num, i := ParseInt(stream)
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		*(*int32)(store.obj) = int32(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int64)(store.obj)
		out = strconv.AppendInt(in, num, 10)
		return
	}
	return
}

func uint16MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint16)(store.obj) = uint16(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint16)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint16)(store.obj) = uint16(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint16)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int16MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			num, i := ParseInt(stream)
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			*(*int16)(store.obj) = int16(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int16)(store.obj)
			out = strconv.AppendInt(in, int64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		num, i := ParseInt(stream)
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		*(*int16)(store.obj) = int16(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int16)(store.obj)
		out = strconv.AppendInt(in, int64(num), 10)
		return
	}
	return
}

func uint8MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseUint(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*uint8)(store.obj) = uint8(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*uint8)(store.obj)
			out = strconv.AppendUint(in, uint64(num), 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		num, err := strconv.ParseUint(stream[:i], 10, 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*uint8)(store.obj) = uint8(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*uint8)(store.obj)
		out = strconv.AppendUint(in, uint64(num), 10)
		return
	}
	return
}

func int8MFuncs1(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			num, err := strconv.ParseInt(stream[:i], 10, 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*int8)(store.obj) = int8(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int64)(store.obj)
			out = strconv.AppendInt(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		num, i := ParseInt(stream)
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		*(*int8)(store.obj) = int8(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int8)(store.obj)
		out = strconv.AppendInt(in, int64(num), 10)
		return
	}
	return
}

// 数值：十进制数，不能有前导0，可以为负数，可以有小数部分。还可以用e或者E表示指数部分。
// 不能包含非数，如NaN。不区分整数与浮点数。JavaScript用双精度浮点数表示所有数值。
var toNums = func() (out [128]int8) {
	nums := map[int]int8{
		'0': 0, '1': 1, '2': 2, '3': 3, '4': 4,
		'5': 5, '6': 6, '7': 7, '8': 8, '9': 9,
	}
	for c := range out {
		if spaceTable[c] || c == ']' || c == '}' || c == ',' {
			out[c] = -3
			continue
		}
		if n, ok := nums[c]; ok {
			out[c] = n
			continue
		}
		out[c] = -4
	}
	out['.'] = -1
	out['e'] = -2
	out['E'] = -2
	out['-'] = -5
	out['+'] = -6
	return
}()

func ParseInt(stream string) (num int64, i int) {
	var e, float, nFloat int64 = 0, 0, 0
	sign, eFlag, floatFlag := false, false, false
	for i < len(stream) {
		c := stream[i]
		nextNum := toNums[c]
		if nextNum >= 0 {
			num = num*10 + int64(nextNum)
			i++
			continue
		}
		if nextNum == -3 {
			break // 退出条件
		}

		// 非主流逻辑分支后置，保持主要分支简单快速
		if i == 0 {
			if stream[0] == '-' {
				sign = true
				i++
				continue
			} else if stream[0] == '+' {
				i++
				continue
			}
		}
		if !eFlag && nextNum == -1 {
			eFlag = true // e 或 E
			for i++; i < len(stream); i++ {
				c := stream[i]
				nextNum = toNums[c]
				if nextNum >= 0 {
					e = e*10 + int64(nextNum)
					continue
				}
				break
			}
			if nextNum == -3 {
				E := int64(1)
				for j := int64(0); j < e; j++ {
					E *= 10
				}
				if num == 0 && E > 1 {
					num = 1
				}
				num *= E
				if float > 0 {
					f := float64(float)
					nFloat = nFloat - e
					if nFloat > 0 {
						for j := int64(0); j < nFloat; j++ {
							f /= 10
						}
					} else {
						nFloat = -nFloat
						for j := int64(0); j < nFloat; j++ {
							f *= 10
						}
					}
					num = num + int64(f)
				}
				break
			}
			continue
		}
		if !eFlag && !floatFlag && nextNum == -2 {
			floatFlag = true // .
			for i++; i < len(stream); i++ {
				c := stream[i]
				nextNum = toNums[c]
				if nextNum >= 0 {
					float = float*10 + int64(nextNum)
					nFloat++
					continue
				}
				break
			}
			continue
		}
		err := lxterrs.New("error ParseInt:%s", ErrStream(stream[:i]))
		panic(err)
	}
	if sign {
		num = num * -1
	}
	return
}

func int8MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			num, i := ParseInt(stream)
			*(*int8)(store.obj) = int8(num)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*int64)(store.obj)
			out = strconv.AppendInt(in, num, 10)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		num, i := ParseInt(stream)
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		*(*int8)(store.obj) = int8(num)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*int8)(store.obj)
		out = strconv.AppendInt(in, int64(num), 10)
		return
	}
	return
}

func float64MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			f, err := strconv.ParseFloat(stream[:i], 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*float64)(store.obj) = f
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*float64)(store.obj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		f, err := strconv.ParseFloat(stream[:i], 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*float64)(store.obj) = f
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*float64)(store.obj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	return
}
func float32MFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			iSlash = idxSlash
			for ; i < len(stream); i++ {
				c := stream[i]
				if spaceTable[c] || c == ']' || c == '}' || c == ',' {
					break
				}
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			f, err := strconv.ParseFloat(stream[:i], 64)
			if err != nil {
				err = lxterrs.Wrap(err, ErrStream(stream[:i]))
				panic(err)
			}
			*(*float32)(store.obj) = float32(f)
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			num := *(*float32)(store.obj)
			out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		iSlash = idxSlash
		for ; i < len(stream); i++ {
			c := stream[i]
			if spaceTable[c] || c == ']' || c == '}' || c == ',' {
				break
			}
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		f, err := strconv.ParseFloat(stream[:i], 64)
		if err != nil {
			err = lxterrs.Wrap(err, ErrStream(stream[:i]))
			panic(err)
		}
		*(*float32)(store.obj) = float32(f)
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		num := *(*float32)(store.obj)
		out = strconv.AppendFloat(in, float64(num), 'f', -1, 64)
		return
	}
	return
}

func structMFuncs(pidx, sonPidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.pointerPool = pointerOffset(store.pointerPool, *sonPidx) //这里有问题，这个 pool 导致 slicePool 的偏移
			store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
			n, iSlash := parseObj(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			out = marshalStruct(store, in)
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
		store.pointerPool = pointerOffset(store.pointerPool, *sonPidx) //这里有问题，这个 pool 导致 slicePool 的偏移
		store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseObj(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		out = marshalStruct(store, in)
		return
	}
	return
}

func sliceIntsMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseIntSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseIntSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}
func sliceNoscanMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func sliceNoscanMFuncs2(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseNoscanSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func sliceMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //
			store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
			n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
			iSlash++
			i += n + 1
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			son := store.tag.ChildList[0]
			out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
			return
		}
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = store.Idx(*pidx)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.slicePool = pointerOffset(store.slicePool, store.tag.idxSliceObjPool)
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx) // TODO 这个可以 pidx==nil 合并? 这时 *pidx==0？
		}
		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func bytesMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				store.obj = pointerOffset(store.obj, store.tag.Offset)
				pHeader := (*SliceHeader)(store.obj)
				pHeader.Data = store.obj
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset) //

			//TODO : 解析长度（,]}），hex.Decode
			bytesSet(store, stream[i:])
			return
		}
		fM = bytesGet
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		p := *(*unsafe.Pointer)(store.obj)
		if p == nil {
			store.obj = store.Idx(*pidx)
		}
		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		bytesGet(store, in)
		return
	}
	return
}

func slicePointerMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			store.obj = *(*unsafe.Pointer)(store.obj)
			if store.obj == nil {
				store.obj = store.Idx(*pidx)
			}
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			store.obj = store.Idx(*pidx)
		}

		n, iSlash := parseSlice(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		// pHeader := (*SliceHeader)(store.obj)
		// son := store.tag.ChildList[0]
		// out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func sliceStringsMFuncs() (fUnm unmFunc, fM mFunc) {
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pHeader := (*SliceHeader)(store.obj)
			pHeader.Data = store.obj
			return
		}
		store.obj = pointerOffset(store.obj, store.tag.Offset) //
		n, iSlash := parseSliceString(idxSlash-1, stream[1:], store)
		iSlash++
		i += n + 1
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		pHeader := (*SliceHeader)(store.obj)
		son := store.tag.ChildList[0]
		out = marshalSlice(in, Store{obj: pHeader.Data, tag: son}, pHeader.Len)
		return
	}
	return
}

func stringUnm(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
	if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
		i = 4
		iSlash = idxSlash
		return
	}
	store.obj = pointerOffset(store.obj, store.tag.Offset)
	pstr := (*string)(store.obj)
	{
		i = strings.IndexByte(stream[1:], '"')
		if idxSlash > i+1 {
			i++
			*pstr = stream[1:i]
			i++
			iSlash = idxSlash
		} else {
			i++
			*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
		}
	}
	return
}
func stringM(store Store, in []byte) (out []byte) {
	str := *(*string)(store.obj)
	return stringMm(str, in)
}
func stringMm(str string, in []byte) (out []byte) {
	out = append(in, '"')
	// strings.ReplaceAll(str, "\\", "\\\\")
	// nSlash := strings.Count(str, "\\")
	i := strings.IndexByte(str, '\\') // 只处理 " , \ 可以不处理
	if i == -1 {
	} else {
		bs := []byte{}
		for {
			bs = append(bs, str[:i]...)
			bs = append(bs, '\\', '\\')
			str = str[i+1:]
			i = strings.IndexByte(str, '\\')
			if i == -1 {
				bs = append(bs, str...)
				break
			}
		}
		if len(bs) > 0 {
			str = bytesString(bs)
		}
	}

	// nQuote := strings.Count(str, "\"")
	i = strings.IndexByte(str, '"') // 只处理 " , \ 可以不处理
	if i == -1 {
		out = append(out, str...) // TODO 需要转义： \ --> \\
	} else {
		for {
			out = append(out, str[:i]...)
			out = append(out, '\\', '"')
			str = str[i+1:]
			i = strings.IndexByte(str, '"')
			if i == -1 {
				out = append(out, str...)
				break
			}
		}
	}
	/*
		nSlash := strings.Count(str, "\\")
		nQuote := strings.Count(str, "\"")
		if nSlash == 0 {
			if nQuote == 0 {
				out = append(out, str...) // TODO 需要转义： \ --> \\
			} else {
				for {
					i := strings.IndexByte(str, '"')
					if i == -1 {
						out = append(out, str...)
						break
					}
					out = append(out, str[:i]...)
					out = append(out, '\\', '"')
					str = str[i+1:]
				}
			}
		} else if nQuote == 0 {
			for {
				i := strings.IndexByte(str, '\\')
				if i == -1 {
					out = append(out, str...)
					break
				}
				out = append(out, str[:i]...)
				out = append(out, '\\', '\\')
				str = str[i+1:]
			}
		} else {
			iSlash, iQuote := strings.IndexByte(str, '\\'), strings.IndexByte(str, '"')
			for iSlash == -1 && iQuote == -1 {
				if iSlash==
				if iSlash < iQuote {

				}

				i := strings.IndexByte(str, '\\')
				if i == -1 {
					out = append(out, str...)
					break
				}
				out = append(out, str[:i]...)
				out = append(out, '\\', '\\')
				str = str[i+1:]
			}
		}
	*/
	out = append(out, '"')
	return
}

func stringMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			pstr := (*string)(store.obj)
			{
				i = strings.IndexByte(stream[1:], '"')
				if idxSlash > i+1 {
					i++
					*pstr = stream[1:i]
					i++
					iSlash = idxSlash
				} else {
					i++
					*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
				}
			}
			return
		}
		fM = stringM
		return
	}
	fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
			i = 4
			iSlash = idxSlash
			return
		}
		pstr := (*string)(store.obj)
		{
			i = strings.IndexByte(stream[1:], '"')
			if idxSlash > i+1 {
				i++
				*pstr = stream[1:i]
				i++
				iSlash = idxSlash
			} else {
				i++
				*pstr, i, iSlash = parseUnescapeStr(stream, i, idxSlash)
			}
		}
		return
	}

	fM = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		return stringM(store, in)
	}
	return
}

func interfaceMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			iSlash = idxSlash
			n := trimSpace(stream[i:])
			i += n
			iv := (*interface{})(pointerOffset(store.obj, store.tag.Offset))
			n, iSlash = parseInterface(idxSlash-i, stream[i:], iv)
			idxSlash += i
			i += n
			// *iv = iface
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			iface := *(*interface{})(store.obj)
			out = marshalInterface(in, iface)
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
		iSlash = idxSlash
		n := trimSpace(stream[i:])
		i += n
		store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = store.Idx(*pidx)
		iv := (*interface{})(store.obj)
		n, iSlash = parseInterface(idxSlash-i, stream[i:], iv)
		idxSlash += i
		i += n
		// *iv = iface
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		iface := *(*interface{})(store.obj)
		out = marshalInterface(in, iface)
		return
	}
	return
}

func mapMFuncs(pidx *uintptr) (fUnm unmFunc, fM mFunc) {
	if pidx == nil {
		fUnm = func(idxSlash int, store PoolStore, stream string) (i, iSlash int) {
			if stream[0] == 'n' && stream[1] == 'u' && stream[2] == 'l' && stream[3] == 'l' {
				i = 4
				iSlash = idxSlash
				return
			}
			store.obj = pointerOffset(store.obj, store.tag.Offset)
			m, i, iSlash := parseMapInterface(idxSlash-1, stream[1:])
			iSlash++
			i++
			p := (*map[string]interface{})(store.obj)
			*p = m
			return
		}
		fM = func(store Store, in []byte) (out []byte) {
			// store.obj = pointerOffset(store.obj, store.tag.Offset)
			m := *(*map[string]interface{})(store.obj)
			out = marshalMapInterface(in, m)
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
		m, i, iSlash := parseMapInterface(idxSlash-1, stream[1:])
		iSlash++
		i++
		p := (*map[string]interface{})(store.obj)
		*p = m
		return
	}
	fM = func(store Store, in []byte) (out []byte) {
		// store.obj = pointerOffset(store.obj, store.tag.Offset)
		store.obj = *(*unsafe.Pointer)(store.obj)
		if store.obj == nil {
			out = append(in, "null"...)
			return
		}
		m := *(*map[string]interface{})(store.obj)
		out = marshalMapInterface(in, m)
		return
	}
	return
}
