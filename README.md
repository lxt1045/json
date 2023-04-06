# json
Trying to implement the fastest JSON library for golang.

# 前言
本项目已存在 [blog](https://github.com/lxt1045/blog/main/sample/json/json) 仓库
下存在半年多了，一直没有精力整理。

当前还有一些特性没有实现，且许多边界条件还未覆盖。

边界条件：循环类型

# 性能表现
以纯 Go 语言实现，在性能略胜于 SIMD 实现的 [sonic](https://github.com/bytedance/sonic)。
## 1. sonic 的测试用例
### 1.1 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 small JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/struct_bench_test.go#L684), 结果如下：
```sh
BenchmarkUnmarshalStruct1x_small
BenchmarkUnmarshalStruct1x_small/lxt-st
BenchmarkUnmarshalStruct1x_small/lxt-st-12               1000000              1088 ns/op             484 B/op          1 allocs/op
BenchmarkUnmarshalStruct1x_small/sonic-st
BenchmarkUnmarshalStruct1x_small/sonic-st-12             1000000              1549 ns/op            1393 B/op          7 allocs/op
BenchmarkUnmarshalStruct1x_small/std-st
BenchmarkUnmarshalStruct1x_small/std-st-12               1000000              7272 ns/op            1056 B/op         36 allocs/op
BenchmarkUnmarshalStruct1x_small/sonic.marshal-st
BenchmarkUnmarshalStruct1x_small/sonic.marshal-st-12             1000000               575.6 ns/op           459 B/op          4 allocs/op
BenchmarkUnmarshalStruct1x_small/lxt.marshal-st
BenchmarkUnmarshalStruct1x_small/lxt.marshal-st-12               1000000               591.3 ns/op           320 B/op          0 allocs/op
BenchmarkUnmarshalStruct1x_small/std.marshal-st
BenchmarkUnmarshalStruct1x_small/std.marshal-st-12               1000000              1562 ns/op             384 B/op          1 allocs/op
```


### 1.2 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 medium JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/struct_bench_test.go#L803), 结果如下：
```sh
BenchmarkUnmarshalStruct1x_medium/lxt-st
BenchmarkUnmarshalStruct1x_medium/lxt-st-12               100000             17860 ns/op            5758 B/op         23 allocs/op
BenchmarkUnmarshalStruct1x_medium/sonic-st
BenchmarkUnmarshalStruct1x_medium/sonic-st-12             100000             27510 ns/op           24212 B/op         34 allocs/op
BenchmarkUnmarshalStruct1x_medium/std-st
BenchmarkUnmarshalStruct1x_medium/std-st-12               100000             79680 ns/op             616 B/op          7 allocs/op
BenchmarkUnmarshalStruct1x_medium/sonic.marshal-st
BenchmarkUnmarshalStruct1x_medium/sonic.marshal-st-12             100000              4701 ns/op            9621 B/op          4 allocs/op
BenchmarkUnmarshalStruct1x_medium/lxt.marshal-st
BenchmarkUnmarshalStruct1x_medium/lxt.marshal-st-12               100000              4250 ns/op            8337 B/op          0 allocs/op
BenchmarkUnmarshalStruct1x_medium/std.marshal-st
BenchmarkUnmarshalStruct1x_medium/std.marshal-st-12               100000             15973 ns/op            9473 B/op          1 allocs/op
```

### 1.3 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 large JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/struct_bench_test.go#L921), 结果如下：
```sh

BenchmarkUnmarshalStruct1x_large/lxt-st
BenchmarkUnmarshalStruct1x_large/lxt-st-12                 10000            886464 ns/op          329947 B/op       1469 allocs/op
BenchmarkUnmarshalStruct1x_large/sonic-st
BenchmarkUnmarshalStruct1x_large/sonic-st-12               10000           1333001 ns/op          464365 B/op       1682 allocs/op
BenchmarkUnmarshalStruct1x_large/std-st
BenchmarkUnmarshalStruct1x_large/std-st-12                 10000           5544251 ns/op          601856 B/op       5848 allocs/op
BenchmarkUnmarshalStruct1x_large/sonic.marshal-st
BenchmarkUnmarshalStruct1x_large/sonic.marshal-st-12               10000            244829 ns/op          262327 B/op          4 allocs/op
BenchmarkUnmarshalStruct1x_large/lxt.marshal-st
BenchmarkUnmarshalStruct1x_large/lxt.marshal-st-12                 10000            127117 ns/op          262150 B/op          0 allocs/op
BenchmarkUnmarshalStruct1x_large/std.marshal-st
BenchmarkUnmarshalStruct1x_large/std.marshal-st-12                 10000            734368 ns/op          337952 B/op       1527 allocs/op
```

有以上结果可知，本 JSON 库和 [sonic](https://github.com/bytedance/sonic) 性能处于一个水平，略有胜出。

## 2. 不同 struct 成员类型测试用例

[测试用例源码在这里](https://github.com/lxt1045/json/blob/main/bench_test.go#L186)

测试结果如下：
```sh
BenchmarkUnmarshalType/uint-10-lxt
BenchmarkUnmarshalType/uint-10-lxt-12            1000000               507.5 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/uint-10-sonic
BenchmarkUnmarshalType/uint-10-sonic-12          1000000               494.3 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-lxt
BenchmarkUnmarshalType/Marshal-uint-10-lxt-12    1000000               258.8 ns/op           153 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-sonic
BenchmarkUnmarshalType/Marshal-uint-10-sonic-12                  1000000               264.6 ns/op           247 B/op          4 allocs/op
BenchmarkUnmarshalType/*uint-10-lxt
BenchmarkUnmarshalType/*uint-10-lxt-12                           1000000               487.3 ns/op            79 B/op          0 allocs/op
BenchmarkUnmarshalType/*uint-10-sonic
BenchmarkUnmarshalType/*uint-10-sonic-12                         1000000               498.7 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-lxt
BenchmarkUnmarshalType/Marshal-*uint-10-lxt-12                   1000000               253.4 ns/op           155 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-sonic
BenchmarkUnmarshalType/Marshal-*uint-10-sonic-12                 1000000               300.2 ns/op           249 B/op          4 allocs/op
BenchmarkUnmarshalType/int8-10-lxt
BenchmarkUnmarshalType/int8-10-lxt-12                            1000000               307.8 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/int8-10-sonic
BenchmarkUnmarshalType/int8-10-sonic-12                          1000000               414.5 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-lxt
BenchmarkUnmarshalType/Marshal-int8-10-lxt-12                    1000000               326.5 ns/op           218 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-sonic
BenchmarkUnmarshalType/Marshal-int8-10-sonic-12                  1000000               252.7 ns/op           214 B/op          4 allocs/op
BenchmarkUnmarshalType/int-10-lxt
BenchmarkUnmarshalType/int-10-lxt-12                             1000000               400.2 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/int-10-sonic
BenchmarkUnmarshalType/int-10-sonic-12                           1000000               463.0 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-lxt
BenchmarkUnmarshalType/Marshal-int-10-lxt-12                     1000000               236.8 ns/op           152 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-sonic
BenchmarkUnmarshalType/Marshal-int-10-sonic-12                   1000000               258.4 ns/op           247 B/op          4 allocs/op
BenchmarkUnmarshalType/bool-10-lxt
BenchmarkUnmarshalType/bool-10-lxt-12                            1000000               261.6 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/bool-10-sonic
BenchmarkUnmarshalType/bool-10-sonic-12                          1000000               349.9 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-lxt
BenchmarkUnmarshalType/Marshal-bool-10-lxt-12                    1000000               117.2 ns/op           135 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-sonic
BenchmarkUnmarshalType/Marshal-bool-10-sonic-12                  1000000               207.2 ns/op           231 B/op          4 allocs/op
BenchmarkUnmarshalType/string-10-lxt
BenchmarkUnmarshalType/string-10-lxt-12                          1000000               313.4 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/string-10-sonic
BenchmarkUnmarshalType/string-10-sonic-12                        1000000               440.0 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-lxt
BenchmarkUnmarshalType/Marshal-string-10-lxt-12                  1000000               150.6 ns/op           205 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-sonic
BenchmarkUnmarshalType/Marshal-string-10-sonic-12                1000000               326.2 ns/op           298 B/op          4 allocs/op
BenchmarkUnmarshalType/[]int8-10-lxt
BenchmarkUnmarshalType/[]int8-10-lxt-12                          1000000               673.1 ns/op            40 B/op          0 allocs/op
BenchmarkUnmarshalType/[]int8-10-sonic
BenchmarkUnmarshalType/[]int8-10-sonic-12                        1000000               702.2 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt-12                  1000000               937.7 ns/op           585 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic-12                1000000               414.2 ns/op           261 B/op          4 allocs/op
BenchmarkUnmarshalType/[]int-10-lxt
BenchmarkUnmarshalType/[]int-10-lxt-12                           1000000               660.6 ns/op           322 B/op          0 allocs/op
BenchmarkUnmarshalType/[]int-10-sonic
BenchmarkUnmarshalType/[]int-10-sonic-12                         1000000               692.2 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-lxt
BenchmarkUnmarshalType/Marshal-[]int-10-lxt-12                   1000000               311.1 ns/op           163 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-sonic
BenchmarkUnmarshalType/Marshal-[]int-10-sonic-12                 1000000               451.1 ns/op           264 B/op          4 allocs/op
BenchmarkUnmarshalType/[]bool-10-lxt
BenchmarkUnmarshalType/[]bool-10-lxt-12                          1000000               584.5 ns/op            40 B/op          0 allocs/op
BenchmarkUnmarshalType/[]bool-10-sonic
BenchmarkUnmarshalType/[]bool-10-sonic-12                        1000000               513.2 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt-12                  1000000               250.4 ns/op           254 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic-12                1000000               325.5 ns/op           359 B/op          4 allocs/op
BenchmarkUnmarshalType/[]string-10-lxt
BenchmarkUnmarshalType/[]string-10-lxt-12                        1000000               833.0 ns/op           640 B/op          0 allocs/op
BenchmarkUnmarshalType/[]string-10-sonic
BenchmarkUnmarshalType/[]string-10-sonic-12                      1000000               800.2 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-lxt
BenchmarkUnmarshalType/Marshal-[]string-10-lxt-12                1000000               312.6 ns/op           228 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-sonic
BenchmarkUnmarshalType/Marshal-[]string-10-sonic-12              1000000               544.5 ns/op           327 B/op          4 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-lxt
BenchmarkUnmarshalType/[]json_test.X-10-lxt-12                   1000000              2358 ns/op            1279 B/op          0 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-sonic
BenchmarkUnmarshalType/[]json_test.X-10-sonic-12                 1000000              3094 ns/op               0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt-12           1000000               833.6 ns/op           706 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic-12                 1000000              1308 ns/op             968 B/op          4 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-lxt
BenchmarkUnmarshalType/[]json_test.Y-10-lxt-12                           1000000              2426 ns/op              80 B/op          0 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-sonic
BenchmarkUnmarshalType/[]json_test.Y-10-sonic-12                         1000000              2958 ns/op               0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt-12                   1000000               655.1 ns/op           584 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic-12                 1000000               899.4 ns/op           843 B/op          4 allocs/op
BenchmarkUnmarshalType/*int-10-lxt
BenchmarkUnmarshalType/*int-10-lxt-12                                    1000000               348.0 ns/op            79 B/op          0 allocs/op
BenchmarkUnmarshalType/*int-10-sonic
BenchmarkUnmarshalType/*int-10-sonic-12                                  1000000               422.3 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-lxt
BenchmarkUnmarshalType/Marshal-*int-10-lxt-12                            1000000               150.2 ns/op           114 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-sonic
BenchmarkUnmarshalType/Marshal-*int-10-sonic-12                          1000000               294.2 ns/op           213 B/op          4 allocs/op
BenchmarkUnmarshalType/*bool-10-lxt
BenchmarkUnmarshalType/*bool-10-lxt-12                                   1000000               320.9 ns/op             9 B/op          0 allocs/op
BenchmarkUnmarshalType/*bool-10-sonic
BenchmarkUnmarshalType/*bool-10-sonic-12                                 1000000               360.7 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-lxt
BenchmarkUnmarshalType/Marshal-*bool-10-lxt-12                           1000000               126.0 ns/op           134 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-sonic
BenchmarkUnmarshalType/Marshal-*bool-10-sonic-12                         1000000               243.1 ns/op           231 B/op          4 allocs/op
BenchmarkUnmarshalType/*string-10-lxt
BenchmarkUnmarshalType/*string-10-lxt-12                                 1000000               357.9 ns/op           159 B/op          0 allocs/op
BenchmarkUnmarshalType/*string-10-sonic
BenchmarkUnmarshalType/*string-10-sonic-12                               1000000               451.8 ns/op             0 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-lxt
BenchmarkUnmarshalType/Marshal-*string-10-lxt-12                         1000000               180.6 ns/op           209 B/op          0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-sonic
BenchmarkUnmarshalType/Marshal-*string-10-sonic-12                       1000000               348.9 ns/op           294 B/op          4 allocs/op
```
由测试结果可知，不同 struct 成员类型本 JSON 库和 [sonic](https://github.com/bytedance/sonic) 性能保持同一水平。
准确的说 sonic Unmarshal 性能更好一点， 此库的 Marshal 性能要好很多。

# 3. 继续优化

由 cpu profile 文件 [pprof001.svg](https://raw.githubusercontent.com/lxt1045/json/main/pprof001.svg) 可知，当前 map 访问 CPU 占比已高达 17.53%，如果不修改架构，剩余的优化空间已经比较小了。

不过 "生命不息,折腾不止"，作者将继续折腾。

# todo
当前存在的问题：
1. pointer、slice 的 cache 的 tag 的名字相同的的时候，会有冲突
2. slice 套 slice，pointer slice