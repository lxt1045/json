# json
Trying to implement the fastest JSON library for golang.

# 前言
本项目已存在 [blog](https://github.com/lxt1045/blog/tree/main/sample/json/json) 仓库
下存在半年多了，一直没有精力整理。

当前还有一些特性没有实现，且许多边界条件还未覆盖。

# 性能表现
总的来说，此库以纯 Go 语言实现，在性能追平了以汇编和 SIMD 实现的 [sonic](https://github.com/bytedance/sonic)，甚至略胜一筹。

## 1. sonic 的测试用例
### 1.1 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 small JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/tree/main/struct_bench_test.go#L684), 结果如下：
```sh
BenchmarkUnmarshalStruct1x_small
BenchmarkUnmarshalStruct1x_small/lxt-st
BenchmarkUnmarshalStruct1x_small/lxt-st-12             658035   1699 ns/op  559 B/op  1 allocs/op
BenchmarkUnmarshalStruct1x_small/sonic-st
BenchmarkUnmarshalStruct1x_small/sonic-st-12           685155   1730 ns/op   1394 B/op  7 allocs/op
BenchmarkUnmarshalStruct1x_small/std-st
BenchmarkUnmarshalStruct1x_small/std-st-12             127114   8229 ns/op   1056 B/op   36 allocs/op
BenchmarkUnmarshalStruct1x_small/sonic.marshal-st
BenchmarkUnmarshalStruct1x_small/sonic.marshal-st-12   1771588  678.0 ns/op  456 B/op  4 allocs/op
BenchmarkUnmarshalStruct1x_small/lxt.marshal-st
BenchmarkUnmarshalStruct1x_small/lxt.marshal-st-12     1733680  673.5 ns/op  369 B/op  0 allocs/op
BenchmarkUnmarshalStruct1x_small/std.marshal-st
BenchmarkUnmarshalStruct1x_small/std.marshal-st-12     624548   1703 ns/op  384 B/op  1 allocs/op
```


### 1.2 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 medium JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/tree/main/struct_bench_test.go#L803), 结果如下：
```sh
BenchmarkUnmarshalStruct1x_medium/lxt-st
BenchmarkUnmarshalStruct1x_medium/lxt-st-12            47306  21882 ns/op   4551 B/op   23 allocs/op
BenchmarkUnmarshalStruct1x_medium/sonic-st
BenchmarkUnmarshalStruct1x_medium/sonic-st-12          35698  33193 ns/op  24222 B/op   34 allocs/op
BenchmarkUnmarshalStruct1x_medium/std-st
BenchmarkUnmarshalStruct1x_medium/std-st-12            13432   128645 ns/op  616 B/op  7 allocs/op
BenchmarkUnmarshalStruct1x_medium/sonic.marshal-st
BenchmarkUnmarshalStruct1x_medium/sonic.marshal-st-12  198567   5416 ns/op   9677 B/op  4 allocs/op
BenchmarkUnmarshalStruct1x_medium/lxt.marshal-st
BenchmarkUnmarshalStruct1x_medium/lxt.marshal-st-12    263822   5140 ns/op   8971 B/op  0 allocs/op
BenchmarkUnmarshalStruct1x_medium/std.marshal-st
BenchmarkUnmarshalStruct1x_medium/std.marshal-st-12     60115  18047 ns/op   9473 B/op  1 allocs/op
```

### 1.3 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 large JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/tree/main/struct_bench_test.go#L921), 结果如下：
```sh

BenchmarkUnmarshalStruct1x_large/lxt-st
BenchmarkUnmarshalStruct1x_large/lxt-st-12             1464   787451 ns/op   317953 B/op   1470 allocs/op
BenchmarkUnmarshalStruct1x_large/sonic-st
BenchmarkUnmarshalStruct1x_large/sonic-st-12           1060  1097331 ns/op   464386 B/op   1682 allocs/op
BenchmarkUnmarshalStruct1x_large/std-st
BenchmarkUnmarshalStruct1x_large/std-st-12             237  5181966 ns/op   601856 B/op   5848 allocs/op
BenchmarkUnmarshalStruct1x_large/sonic.marshal-st
BenchmarkUnmarshalStruct1x_large/sonic.marshal-st-12  7488   160644 ns/op   263155 B/op  4 allocs/op
BenchmarkUnmarshalStruct1x_large/lxt.marshal-st
BenchmarkUnmarshalStruct1x_large/lxt.marshal-st-12    7185   142817 ns/op   262279 B/op  0 allocs/op
BenchmarkUnmarshalStruct1x_large/std.marshal-st
BenchmarkUnmarshalStruct1x_large/std.marshal-st-12    1753   697935 ns/op   338285 B/op   1527 allocs/op
```

有以上结果可知，本 JSON 库和 [sonic](https://github.com/bytedance/sonic) 性能处于一个水平，略有胜出。

## 2. 不同 struct 成员类型测试用例

[测试用例源码在这里](https://github.com/lxt1045/json/blob/tree/main/bench_test.go#L186)
```go
func BenchmarkUnmarshalType(b *testing.B) {
 type X struct {
  A string
  B string
 }
 type Y struct {
  A bool
  B bool
 }
 all := []struct {
  V   interface{}
  JsonV string
 }{
  {uint(0), `888888`},  // 0
  {(*uint)(nil), `888888`},   // 1
  {int8(0), `88`},  // 2
  {int(0), `888888`},   // 3
  {true, `true`},   // 4
  {"", `"asdfghjkl"`},  // 5
  {[]int8{}, `[1,2,3]`},  // 6
  {[]int{}, `[1,2,3]`},   // 7
  {[]bool{}, `[true,true,true]`}, // 8
  {[]string{}, `["1","2","3"]`},  // 9
  {[]X{}, `[{"A":"aaaa","B":"bbbb"},{"A":"aaaa","B":"bbbb"},{"A":"aaaa","B":"bbbb"}]`},
  {[]Y{}, `[{"A":true,"B":true},{"A":true,"B":true},{"A":true,"B":true}]`},
  {(*int)(nil), `88`},   // 11
  {(*bool)(nil), `true`},  // 12
  {(*string)(nil), `"asdfghjkl"`}, // 13
 }
 N := 10
 idxs := []int{}
 // idxs = []int{8, 9, 10, 11}
 if len(idxs) > 0 {
  get := all[:0]
  for _, i := range idxs {
   get = append(get, all[i])
  }
  all = get
 }
 for _, obj := range all {
  builder := lxt.NewTypeBuilder()
  buf := bytes.NewBufferString("{")
  fieldType := reflect.TypeOf(obj.V)
  for i := 0; i < N; i++ {
   if i != 0 {
  buf.WriteByte(',')
   }
   key := fmt.Sprintf("Field_%d", i)
   builder.AddField(key, fieldType)
   buf.WriteString(fmt.Sprintf(`"%s":%v`, key, obj.JsonV))
  }
  buf.WriteByte('}')
  bs := buf.Bytes()
  str := string(bs)

  typ := builder.Build()
  value := reflect.New(typ).Interface()

  runtime.GC()
  b.Run(fmt.Sprintf("%s-%d-lxt", fieldType, N), func(b *testing.B) {
   for i := 0; i < b.N; i++ {
  err := lxt.UnmarshalString(str, value)
  if err != nil {
   b.Fatal(err)
  }
   }
  })
  runtime.GC()
  b.Run(fmt.Sprintf("%s-%d-sonic", fieldType, N), func(b *testing.B) {
   for i := 0; i < b.N; i++ {
  err := sonic.UnmarshalString(str, value)
  if err != nil {
   b.Fatal(err)
  }
   }
  })
  
  runtime.GC()
  var err error
  b.Run(fmt.Sprintf("Marshal-%s-%d-lxt", fieldType, N), func(b *testing.B) {
   for i := 0; i < b.N; i++ {
  bs, err = lxt.Marshal(value)
  if err != nil {
   b.Fatal(err)
  }
   }
  })
  runtime.GC()
  b.Run(fmt.Sprintf("Marshal-%s-%d-sonic", fieldType, N), func(b *testing.B) {
   for i := 0; i < b.N; i++ {
  bs, err = sonic.Marshal(value)
  if err != nil {
   b.Fatal(err)
  }
   }
  })
 }
}
```

测试结果如下：
```sh

BenchmarkUnmarshalType/uint-10-lxt
BenchmarkUnmarshalType/uint-10-lxt-12     2145248  545.7 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/uint-10-sonic
BenchmarkUnmarshalType/uint-10-sonic-12   2321377  497.0 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-lxt
BenchmarkUnmarshalType/Marshal-uint-10-lxt-12     4514630  251.8 ns/op  176 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-sonic
BenchmarkUnmarshalType/Marshal-uint-10-sonic-12   3929481  293.5 ns/op  249 B/op  4 allocs/op
BenchmarkUnmarshalType/*uint-10-lxt
BenchmarkUnmarshalType/*uint-10-lxt-12      1803104  603.5 ns/op   80 B/op  0 allocs/op
BenchmarkUnmarshalType/*uint-10-sonic
BenchmarkUnmarshalType/*uint-10-sonic-12    2040158  536.3 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-lxt
BenchmarkUnmarshalType/Marshal-*uint-10-lxt-12    4806735  246.2 ns/op  174 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-sonic
BenchmarkUnmarshalType/Marshal-*uint-10-sonic-12  3650094  302.2 ns/op  245 B/op  4 allocs/op
BenchmarkUnmarshalType/int8-10-lxt
BenchmarkUnmarshalType/int8-10-lxt-12       2390233  482.3 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/int8-10-sonic
BenchmarkUnmarshalType/int8-10-sonic-12     2479201  434.7 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-lxt
BenchmarkUnmarshalType/Marshal-int8-10-lxt-12     3148892  371.7 ns/op  255 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-sonic
BenchmarkUnmarshalType/Marshal-int8-10-sonic-12   4257580  276.5 ns/op  216 B/op  4 allocs/op
BenchmarkUnmarshalType/int-10-lxt
BenchmarkUnmarshalType/int-10-lxt-12      1775660  592.9 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/int-10-sonic
BenchmarkUnmarshalType/int-10-sonic-12    2389962  462.8 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-lxt
BenchmarkUnmarshalType/Marshal-int-10-lxt-12    4953722  247.4 ns/op  174 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-sonic
BenchmarkUnmarshalType/Marshal-int-10-sonic-12  4396863  270.4 ns/op  248 B/op  4 allocs/op
BenchmarkUnmarshalType/bool-10-lxt
BenchmarkUnmarshalType/bool-10-lxt-12       4092759  312.3 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/bool-10-sonic
BenchmarkUnmarshalType/bool-10-sonic-12     3231327  375.9 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-lxt
BenchmarkUnmarshalType/Marshal-bool-10-lxt-12     8891221  137.4 ns/op  157 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-sonic
BenchmarkUnmarshalType/Marshal-bool-10-sonic-12   5494783  315.2 ns/op  231 B/op  4 allocs/op
BenchmarkUnmarshalType/string-10-lxt
BenchmarkUnmarshalType/string-10-lxt-12     2655135  411.4 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/string-10-sonic
BenchmarkUnmarshalType/string-10-sonic-12   2535213  481.9 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-lxt
BenchmarkUnmarshalType/Marshal-string-10-lxt-12     6676614  164.2 ns/op  228 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-sonic
BenchmarkUnmarshalType/Marshal-string-10-sonic-12   3184688  365.0 ns/op  297 B/op  4 allocs/op
BenchmarkUnmarshalType/[]int8-10-lxt
BenchmarkUnmarshalType/[]int8-10-lxt-12     952368   1315 ns/op   59 B/op  0 allocs/op
BenchmarkUnmarshalType/[]int8-10-sonic
BenchmarkUnmarshalType/[]int8-10-sonic-12   1641888  732.7 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt-12    1000000   1205 ns/op  622 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic-12   2295127  448.8 ns/op  265 B/op  4 allocs/op
BenchmarkUnmarshalType/[]int-10-lxt
BenchmarkUnmarshalType/[]int-10-lxt-12      897546   1276 ns/op  241 B/op  0 allocs/op
BenchmarkUnmarshalType/[]int-10-sonic
BenchmarkUnmarshalType/[]int-10-sonic-12    1492687  721.1 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-lxt
BenchmarkUnmarshalType/Marshal-[]int-10-lxt-12    3161004  326.0 ns/op  185 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-sonic
BenchmarkUnmarshalType/Marshal-[]int-10-sonic-12  2737060  432.6 ns/op  262 B/op  4 allocs/op
BenchmarkUnmarshalType/[]bool-10-lxt
BenchmarkUnmarshalType/[]bool-10-lxt-12     1000000   1056 ns/op   60 B/op  0 allocs/op
BenchmarkUnmarshalType/[]bool-10-sonic
BenchmarkUnmarshalType/[]bool-10-sonic-12   2110680  538.5 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt-12     4232835  245.0 ns/op  277 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic-12   2706502  426.9 ns/op  359 B/op  4 allocs/op
BenchmarkUnmarshalType/[]string-10-lxt
BenchmarkUnmarshalType/[]string-10-lxt-12     1203486  954.1 ns/op  480 B/op  0 allocs/op
BenchmarkUnmarshalType/[]string-10-sonic
BenchmarkUnmarshalType/[]string-10-sonic-12   1294734  887.6 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-lxt
BenchmarkUnmarshalType/Marshal-[]string-10-lxt-12     3569140  392.5 ns/op  245 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-sonic
BenchmarkUnmarshalType/Marshal-[]string-10-sonic-12   1871164  654.7 ns/op  328 B/op  4 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-lxt
BenchmarkUnmarshalType/[]json_test.X-10-lxt-12     319884   3538 ns/op  966 B/op  0 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-sonic
BenchmarkUnmarshalType/[]json_test.X-10-sonic-12   263923   4407 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt-12    1242693  949.6 ns/op  848 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic-12  630661   1674 ns/op  976 B/op  4 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-lxt
BenchmarkUnmarshalType/[]json_test.Y-10-lxt-12      364486   2772 ns/op   1139 B/op  0 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-sonic
BenchmarkUnmarshalType/[]json_test.Y-10-sonic-12    382804   2740 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt-12     1680015  668.6 ns/op  725 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic-12   1239610  897.5 ns/op  843 B/op  4 allocs/op
BenchmarkUnmarshalType/*int-10-lxt
BenchmarkUnmarshalType/*int-10-lxt-12      2422323  570.3 ns/op   79 B/op  0 allocs/op
BenchmarkUnmarshalType/*int-10-sonic
BenchmarkUnmarshalType/*int-10-sonic-12    2481496  505.3 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-lxt
BenchmarkUnmarshalType/Marshal-*int-10-lxt-12      6687646  165.8 ns/op  137 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-sonic
BenchmarkUnmarshalType/Marshal-*int-10-sonic-12    4006986  305.5 ns/op  215 B/op  4 allocs/op
BenchmarkUnmarshalType/*bool-10-lxt
BenchmarkUnmarshalType/*bool-10-lxt-12       3354952  384.6 ns/op   10 B/op  0 allocs/op
BenchmarkUnmarshalType/*bool-10-sonic
BenchmarkUnmarshalType/*bool-10-sonic-12     3091408  400.2 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-lxt
BenchmarkUnmarshalType/Marshal-*bool-10-lxt-12       9628600  125.2 ns/op  157 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-sonic
BenchmarkUnmarshalType/Marshal-*bool-10-sonic-12     4840723  263.6 ns/op  230 B/op  4 allocs/op
BenchmarkUnmarshalType/*string-10-lxt
BenchmarkUnmarshalType/*string-10-lxt-12       2298200  471.9 ns/op  159 B/op  0 allocs/op
BenchmarkUnmarshalType/*string-10-sonic
BenchmarkUnmarshalType/*string-10-sonic-12     2440693  498.4 ns/op  0 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-lxt
BenchmarkUnmarshalType/Marshal-*string-10-lxt-12     6849554  171.3 ns/op  227 B/op  0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-sonic
BenchmarkUnmarshalType/Marshal-*string-10-sonic-12   3336834  355.1 ns/op  294 B/op  4 allocs/op
```
由测试结果可知，不同 struct 成员类型本 JSON 库和 [sonic](https://github.com/bytedance/sonic) 性能保持同一水平。
准确的说 sonic Unmarshal 性能更好一点， 此库的 Marshal 性能要好很多。

# 3. 继续优化

由 cpu profile 文件 [pprof001.svg](https://github.com/lxt1045/json/blob/tree/main/pprof001.svg) 可知，当前 map 访问 CPU 占比已高达 17.53%，如果不修改架构，剩余的优化空间已经比较小了。

不过 "生命不息,折腾不止"，作者将继续折腾。

# todo
当前存在的问题：
1. pointer、slice 的 cache 的 tag 的名字相同的的时候，会有冲突
2. slice 套 slice，pointer slice