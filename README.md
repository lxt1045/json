# json
Trying to implement the fastest JSON library for golang.

# 前言
本项目已存在 [blog](https://github.com/lxt1045/blog/main/sample/json/json) 仓库
下存在半年多了，一直没有精力整理。

当前还有一些特性没有实现，且许多边界条件还未覆盖。

边界条件：循环类型

# 性能表现
以纯 Go 语言实现，在性能上全面超越 SIMD 实现的 [sonic](https://github.com/bytedance/sonic)。
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
BenchmarkUnmarshalType/uint-10-lxt-12         	 2988464	       364.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/uint-10-sonic
BenchmarkUnmarshalType/uint-10-sonic-12       	 2242598	       513.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-lxt
BenchmarkUnmarshalType/Marshal-uint-10-lxt-12 	 4004668	       253.1 ns/op	     155 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-sonic
BenchmarkUnmarshalType/Marshal-uint-10-sonic-12         	 4579678	       262.9 ns/op	     247 B/op	       4 allocs/op
BenchmarkUnmarshalType/*uint-10-lxt
BenchmarkUnmarshalType/*uint-10-lxt-12                  	 3153646	       382.5 ns/op	      80 B/op	       0 allocs/op
BenchmarkUnmarshalType/*uint-10-sonic
BenchmarkUnmarshalType/*uint-10-sonic-12                	 1762716	       584.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-lxt
BenchmarkUnmarshalType/Marshal-*uint-10-lxt-12          	 4400548	       250.1 ns/op	     154 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-sonic
BenchmarkUnmarshalType/Marshal-*uint-10-sonic-12        	 3450585	       357.9 ns/op	     246 B/op	       4 allocs/op
BenchmarkUnmarshalType/int8-10-lxt
BenchmarkUnmarshalType/int8-10-lxt-12                   	 4916164	       288.4 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/int8-10-sonic
BenchmarkUnmarshalType/int8-10-sonic-12                 	 2101341	       511.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-lxt
BenchmarkUnmarshalType/Marshal-int8-10-lxt-12           	 3066370	       388.1 ns/op	     236 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-sonic
BenchmarkUnmarshalType/Marshal-int8-10-sonic-12         	 4181509	       259.1 ns/op	     214 B/op	       4 allocs/op
BenchmarkUnmarshalType/int-10-lxt
BenchmarkUnmarshalType/int-10-lxt-12                    	 3639271	       299.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/int-10-sonic
BenchmarkUnmarshalType/int-10-sonic-12                  	 2027686	       562.1 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-lxt
BenchmarkUnmarshalType/Marshal-int-10-lxt-12            	 4231722	       253.3 ns/op	     155 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-sonic
BenchmarkUnmarshalType/Marshal-int-10-sonic-12          	 3471198	       313.0 ns/op	     246 B/op	       4 allocs/op
BenchmarkUnmarshalType/bool-10-lxt
BenchmarkUnmarshalType/bool-10-lxt-12                   	 4855257	       243.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/bool-10-sonic
BenchmarkUnmarshalType/bool-10-sonic-12                 	 2286252	       470.3 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-lxt
BenchmarkUnmarshalType/Marshal-bool-10-lxt-12           	 8860080	       127.4 ns/op	     135 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-sonic
BenchmarkUnmarshalType/Marshal-bool-10-sonic-12         	 5114204	       212.0 ns/op	     231 B/op	       4 allocs/op
BenchmarkUnmarshalType/string-10-lxt
BenchmarkUnmarshalType/string-10-lxt-12                 	 3872169	       283.9 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/string-10-sonic
BenchmarkUnmarshalType/string-10-sonic-12               	 2089436	       528.5 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-lxt
BenchmarkUnmarshalType/Marshal-string-10-lxt-12         	 6741531	       166.1 ns/op	     207 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-sonic
BenchmarkUnmarshalType/Marshal-string-10-sonic-12       	 2640980	       416.5 ns/op	     300 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]int8-10-lxt
BenchmarkUnmarshalType/[]int8-10-lxt-12                 	 1652502	       635.0 ns/op	      40 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int8-10-sonic
BenchmarkUnmarshalType/[]int8-10-sonic-12               	 1318542	       785.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt-12         	 1000000	      1131 ns/op	     586 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic-12       	 2091970	       486.8 ns/op	     261 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]int-10-lxt
BenchmarkUnmarshalType/[]int-10-lxt-12                  	 1634740	       640.6 ns/op	     322 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int-10-sonic
BenchmarkUnmarshalType/[]int-10-sonic-12                	 1357246	       917.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-lxt
BenchmarkUnmarshalType/Marshal-[]int-10-lxt-12          	 2527028	       416.6 ns/op	     165 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-sonic
BenchmarkUnmarshalType/Marshal-[]int-10-sonic-12        	 2046219	       548.5 ns/op	     260 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]bool-10-lxt
BenchmarkUnmarshalType/[]bool-10-lxt-12                 	 1942707	       516.5 ns/op	      40 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]bool-10-sonic
BenchmarkUnmarshalType/[]bool-10-sonic-12               	 2187367	       542.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt-12         	 4511371	       241.1 ns/op	     258 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic-12       	 2676241	       376.1 ns/op	     358 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]string-10-lxt
BenchmarkUnmarshalType/[]string-10-lxt-12               	 1348515	       800.2 ns/op	     640 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]string-10-sonic
BenchmarkUnmarshalType/[]string-10-sonic-12             	 1000000	      1174 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-lxt
BenchmarkUnmarshalType/Marshal-[]string-10-lxt-12       	 3782775	       306.6 ns/op	     227 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-sonic
BenchmarkUnmarshalType/Marshal-[]string-10-sonic-12     	 1769052	       613.7 ns/op	     326 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-lxt
BenchmarkUnmarshalType/[]json_test.X-10-lxt-12          	  572046	      2428 ns/op	    1280 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-sonic
BenchmarkUnmarshalType/[]json_test.X-10-sonic-12        	  255798	      4336 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt-12  	 1000000	      1108 ns/op	     706 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic-12         	  644772	      1648 ns/op	     971 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-lxt
BenchmarkUnmarshalType/[]json_test.Y-10-lxt-12                   	  637653	      1818 ns/op	      80 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-sonic
BenchmarkUnmarshalType/[]json_test.Y-10-sonic-12                 	  399015	      2693 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt-12           	 1463352	       726.0 ns/op	     585 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic-12         	 1000000	      1135 ns/op	     845 B/op	       4 allocs/op
BenchmarkUnmarshalType/*int-10-lxt
BenchmarkUnmarshalType/*int-10-lxt-12                            	 4141521	       297.8 ns/op	      79 B/op	       0 allocs/op
BenchmarkUnmarshalType/*int-10-sonic
BenchmarkUnmarshalType/*int-10-sonic-12                          	 2139524	       506.6 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-lxt
BenchmarkUnmarshalType/Marshal-*int-10-lxt-12                    	 7335110	       154.4 ns/op	     113 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-sonic
BenchmarkUnmarshalType/Marshal-*int-10-sonic-12                  	 3312554	       314.3 ns/op	     211 B/op	       4 allocs/op
BenchmarkUnmarshalType/*bool-10-lxt
BenchmarkUnmarshalType/*bool-10-lxt-12                           	 5023176	       232.9 ns/op	      10 B/op	       0 allocs/op
BenchmarkUnmarshalType/*bool-10-sonic
BenchmarkUnmarshalType/*bool-10-sonic-12                         	 2453275	       425.7 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-lxt
BenchmarkUnmarshalType/Marshal-*bool-10-lxt-12                   	 9118153	       124.2 ns/op	     136 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-sonic
BenchmarkUnmarshalType/Marshal-*bool-10-sonic-12                 	 4318794	       241.6 ns/op	     231 B/op	       4 allocs/op
BenchmarkUnmarshalType/*string-10-lxt
BenchmarkUnmarshalType/*string-10-lxt-12                         	 2946948	       342.6 ns/op	     160 B/op	       0 allocs/op
BenchmarkUnmarshalType/*string-10-sonic
BenchmarkUnmarshalType/*string-10-sonic-12                       	 2044615	       577.8 ns/op	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-lxt
BenchmarkUnmarshalType/Marshal-*string-10-lxt-12                 	 7202666	       157.5 ns/op	     207 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-sonic
BenchmarkUnmarshalType/Marshal-*string-10-sonic-12               	 2640258	       395.4 ns/op	     295 B/op	       4 allocs/op
```
由测试结果可知，不同 struct 成员类型本 JSON 库和 [sonic](https://github.com/bytedance/sonic) 性能保持同一水平。
准确的说 sonic Unmarshal 性能更好一点， 此库的 Marshal 性能要好很多。

# 3. 持续优化
生命不息,折腾不止，作者将继续折腾。

# todo
当前存在的问题：
1. pointer、slice 的 cache 的 tag 的名字相同的的时候，会有冲突
2. slice 套 slice，pointer slice


# 交流学习
![扫码加微信好友](https://github.com/lxt1045/wechatbot/blob/main/resource/Wechat-lxt.png "微信")