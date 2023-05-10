# json
Trying to implement the fastest JSON library for golang.

# 前言
本项目已存在 [blog](https://github.com/lxt1045/blog/main/sample/json/json) 仓库
下存在半年多了，一直没有精力整理。

当前还有一些特性没有实现，且许多边界条件还未覆盖。


# 性能表现
以纯 Go 语言实现，在性能上全面超越 SIMD 实现的 [sonic](https://github.com/bytedance/sonic)。
## 1. sonic 的测试用例
### 1.1 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 small JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/bench_test.go#L350), 结果如下：
```sh
BenchmarkSmallBinding/decode-lxt
BenchmarkSmallBinding/decode-lxt-12         	  986773	      1019 ns/op	 358.32 MB/s	     484 B/op	       1 allocs/op
BenchmarkSmallBinding/decode-sonic
BenchmarkSmallBinding/decode-sonic-12       	  675369	      1636 ns/op	 223.08 MB/s	    1394 B/op	       7 allocs/op
BenchmarkSmallBinding/decode-parallel-lxt
BenchmarkSmallBinding/decode-parallel-lxt-12         	 4173538	       279.9 ns/op	1304.23 MB/s	     483 B/op	       1 allocs/op
BenchmarkSmallBinding/decode-parallel-sonic
BenchmarkSmallBinding/decode-parallel-sonic-12       	 3362064	       346.4 ns/op	1053.63 MB/s	    1398 B/op	       7 allocs/op
BenchmarkSmallBinding/encode-lxt
BenchmarkSmallBinding/encode-lxt-12                  	 1509781	       677.2 ns/op	 539.01 MB/s	     321 B/op	       0 allocs/op
BenchmarkSmallBinding/encode-sonic
BenchmarkSmallBinding/encode-sonic-12                	 1517035	       714.7 ns/op	 510.73 MB/s	     458 B/op	       4 allocs/op
BenchmarkSmallBinding/encode-parallel-lxt
BenchmarkSmallBinding/encode-parallel-lxt-12         	 8529007	       140.2 ns/op	2603.04 MB/s	     319 B/op	       0 allocs/op
BenchmarkSmallBinding/encode-parallel-sonic
BenchmarkSmallBinding/encode-parallel-sonic-12       	 8699536	       140.5 ns/op	2597.92 MB/s	     461 B/op	       4 allocs/op
```


### 1.2 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 medium JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/bench_test.go#L552), 结果如下：
```sh
BenchmarkMediumBinding/decode-lxt
BenchmarkMediumBinding/decode-lxt-12         	   88701	     13341 ns/op	 832.02 MB/s	    5757 B/op	      23 allocs/op
BenchmarkMediumBinding/decode-sonic
BenchmarkMediumBinding/decode-sonic-12       	   49826	     26159 ns/op	 424.33 MB/s	   24215 B/op	      34 allocs/op
BenchmarkMediumBinding/decode-parallel-lxt
BenchmarkMediumBinding/decode-parallel-lxt-12         	  312424	      3322 ns/op	3341.32 MB/s	    5795 B/op	      23 allocs/op
BenchmarkMediumBinding/decode-parallel-sonic
BenchmarkMediumBinding/decode-parallel-sonic-12       	  222912	      4962 ns/op	2237.22 MB/s	   24231 B/op	      34 allocs/op
BenchmarkMediumBinding/encode-lxt
BenchmarkMediumBinding/encode-lxt-12                  	  279264	      4051 ns/op	2739.95 MB/s	    8328 B/op	       0 allocs/op
BenchmarkMediumBinding/encode-sonic
BenchmarkMediumBinding/encode-sonic-12                	  246520	      4687 ns/op	2368.36 MB/s	    9585 B/op	       4 allocs/op
BenchmarkMediumBinding/encode-parallel-lxt
BenchmarkMediumBinding/encode-parallel-lxt-12         	  653260	      1600 ns/op	6935.93 MB/s	    8321 B/op	       0 allocs/op
BenchmarkMediumBinding/encode-parallel-sonic
BenchmarkMediumBinding/encode-parallel-sonic-12       	  931518	      1079 ns/op	10287.75 MB/s	    9764 B/op	       4 allocs/op
```

### 1.3 执行 [sonic](https://github.com/bytedance/sonic) 仓库下的 large JSON 数据，
[单测源码在这](https://github.com/lxt1045/json/blob/main/bench_test.go#L754), 结果如下：
```sh
BenchmarkLargeBinding/decode-lxt
BenchmarkLargeBinding/decode-lxt-12         	    1526	    772026 ns/op	 818.00 MB/s	  334450 B/op	    1469 allocs/op
BenchmarkLargeBinding/decode-sonic
BenchmarkLargeBinding/decode-sonic-12       	     985	   1299304 ns/op	 486.04 MB/s	  464453 B/op	    1682 allocs/op
BenchmarkLargeBinding/decode-parallel-lxt
BenchmarkLargeBinding/decode-parallel-lxt-12         	    7350	    174831 ns/op	3612.14 MB/s	  326045 B/op	    1469 allocs/op
BenchmarkLargeBinding/decode-parallel-sonic
BenchmarkLargeBinding/decode-parallel-sonic-12       	    6112	    189034 ns/op	3340.75 MB/s	  464345 B/op	    1682 allocs/op
BenchmarkLargeBinding/encode-lxt
BenchmarkLargeBinding/encode-lxt-12                  	    9457	    126598 ns/op	4988.34 MB/s	  262233 B/op	       0 allocs/op
BenchmarkLargeBinding/encode-sonic
BenchmarkLargeBinding/encode-sonic-12                	    7723	    151450 ns/op	4169.79 MB/s	  262362 B/op	       4 allocs/op
BenchmarkLargeBinding/encode-parallel-lxt
BenchmarkLargeBinding/encode-parallel-lxt-12         	   29104	     53317 ns/op	11844.60 MB/s	  262042 B/op	       0 allocs/op
BenchmarkLargeBinding/encode-parallel-sonic
BenchmarkLargeBinding/encode-parallel-sonic-12       	   32044	     44517 ns/op	14186.04 MB/s	  262337 B/op	       4 allocs/op
```

有以上结果可知，在性能上此 JSON 库已经超越 [sonic](https://github.com/bytedance/sonic) 。

## 2. 不同 struct 成员类型测试用例

[测试用例源码在这里](https://github.com/lxt1045/json/blob/main/bench_test.go#L217)

测试结果如下：
```sh
BenchmarkUnmarshalType/uint-10-lxt
BenchmarkUnmarshalType/uint-10-lxt-12         	 3349932	       361.3 ns/op	 473.35 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/uint-10-sonic
BenchmarkUnmarshalType/uint-10-sonic-12       	 2149837	       479.3 ns/op	 356.79 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-lxt
BenchmarkUnmarshalType/Marshal-uint-10-lxt-12 	 4923214	       242.5 ns/op	 705.22 MB/s	     155 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-uint-10-sonic
BenchmarkUnmarshalType/Marshal-uint-10-sonic-12         	 4302016	       263.9 ns/op	 648.04 MB/s	     246 B/op	       4 allocs/op
BenchmarkUnmarshalType/*uint-10-lxt
BenchmarkUnmarshalType/*uint-10-lxt-12                  	 2543614	       407.2 ns/op	 419.90 MB/s	      80 B/op	       0 allocs/op
BenchmarkUnmarshalType/*uint-10-sonic
BenchmarkUnmarshalType/*uint-10-sonic-12                	 2036256	       527.6 ns/op	 324.08 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-lxt
BenchmarkUnmarshalType/Marshal-*uint-10-lxt-12          	 4739866	       251.1 ns/op	 681.12 MB/s	     153 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*uint-10-sonic
BenchmarkUnmarshalType/Marshal-*uint-10-sonic-12        	 3007965	       333.0 ns/op	 513.45 MB/s	     247 B/op	       4 allocs/op
BenchmarkUnmarshalType/int8-10-lxt
BenchmarkUnmarshalType/int8-10-lxt-12                   	 4685821	       286.4 ns/op	 457.43 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/int8-10-sonic
BenchmarkUnmarshalType/int8-10-sonic-12                 	 2468440	       454.1 ns/op	 288.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-lxt
BenchmarkUnmarshalType/Marshal-int8-10-lxt-12           	 3337285	       352.8 ns/op	 371.36 MB/s	     236 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int8-10-sonic
BenchmarkUnmarshalType/Marshal-int8-10-sonic-12         	 4428140	       249.5 ns/op	 524.96 MB/s	     215 B/op	       4 allocs/op
BenchmarkUnmarshalType/int-10-lxt
BenchmarkUnmarshalType/int-10-lxt-12                    	 3962656	       297.2 ns/op	 575.30 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/int-10-sonic
BenchmarkUnmarshalType/int-10-sonic-12                  	 2211376	       464.2 ns/op	 368.39 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-lxt
BenchmarkUnmarshalType/Marshal-int-10-lxt-12            	 3302372	       316.7 ns/op	 539.87 MB/s	     154 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-int-10-sonic
BenchmarkUnmarshalType/Marshal-int-10-sonic-12          	 4071799	       274.1 ns/op	 623.93 MB/s	     248 B/op	       4 allocs/op
BenchmarkUnmarshalType/bool-10-lxt
BenchmarkUnmarshalType/bool-10-lxt-12                   	 4360530	       241.0 ns/op	 626.57 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/bool-10-sonic
BenchmarkUnmarshalType/bool-10-sonic-12                 	 2759778	       392.4 ns/op	 384.80 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-lxt
BenchmarkUnmarshalType/Marshal-bool-10-lxt-12           	 8728110	       136.4 ns/op	1107.24 MB/s	     136 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-bool-10-sonic
BenchmarkUnmarshalType/Marshal-bool-10-sonic-12         	 4859212	       245.5 ns/op	 615.00 MB/s	     232 B/op	       4 allocs/op
BenchmarkUnmarshalType/string-10-lxt
BenchmarkUnmarshalType/string-10-lxt-12                 	 4064012	       296.6 ns/op	 745.14 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/string-10-sonic
BenchmarkUnmarshalType/string-10-sonic-12               	 2531212	       529.0 ns/op	 417.75 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-lxt
BenchmarkUnmarshalType/Marshal-string-10-lxt-12         	 6624231	       186.6 ns/op	1184.29 MB/s	     207 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-string-10-sonic
BenchmarkUnmarshalType/Marshal-string-10-sonic-12       	 3352042	       374.0 ns/op	 590.86 MB/s	     297 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]int8-10-lxt
BenchmarkUnmarshalType/[]int8-10-lxt-12                 	 1803240	       588.8 ns/op	 307.40 MB/s	      40 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int8-10-sonic
BenchmarkUnmarshalType/[]int8-10-sonic-12               	 1492542	       714.9 ns/op	 253.18 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt
BenchmarkUnmarshalType/Marshal-[]int8-10-lxt-12         	 1000000	      1098 ns/op	 164.86 MB/s	     582 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic
BenchmarkUnmarshalType/Marshal-[]int8-10-sonic-12       	 2256643	       455.8 ns/op	 397.07 MB/s	     262 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]int-10-lxt
BenchmarkUnmarshalType/[]int-10-lxt-12                  	 1754726	       674.8 ns/op	 268.23 MB/s	     322 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]int-10-sonic
BenchmarkUnmarshalType/[]int-10-sonic-12                	 1000000	      1620 ns/op	 111.74 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-lxt
BenchmarkUnmarshalType/Marshal-[]int-10-lxt-12          	 3621668	       316.8 ns/op	 571.33 MB/s	     165 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]int-10-sonic
BenchmarkUnmarshalType/Marshal-[]int-10-sonic-12        	 2280108	       457.5 ns/op	 395.59 MB/s	     262 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]bool-10-lxt
BenchmarkUnmarshalType/[]bool-10-lxt-12                 	 2056238	       488.1 ns/op	 555.27 MB/s	      40 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]bool-10-sonic
BenchmarkUnmarshalType/[]bool-10-sonic-12               	 1916445	       528.0 ns/op	 513.21 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt
BenchmarkUnmarshalType/Marshal-[]bool-10-lxt-12         	 4599352	       241.4 ns/op	1122.62 MB/s	     255 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic
BenchmarkUnmarshalType/Marshal-[]bool-10-sonic-12       	 2877079	       377.6 ns/op	 717.73 MB/s	     358 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]string-10-lxt
BenchmarkUnmarshalType/[]string-10-lxt-12               	 1429495	       783.2 ns/op	 307.71 MB/s	     640 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]string-10-sonic
BenchmarkUnmarshalType/[]string-10-sonic-12             	 1000000	      1012 ns/op	 238.20 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-lxt
BenchmarkUnmarshalType/Marshal-[]string-10-lxt-12       	 3645444	       319.9 ns/op	 753.39 MB/s	     224 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]string-10-sonic
BenchmarkUnmarshalType/Marshal-[]string-10-sonic-12     	 1696880	       643.2 ns/op	 374.72 MB/s	     327 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-lxt
BenchmarkUnmarshalType/[]json_test.X-10-lxt-12          	  556695	      2451 ns/op	 343.19 MB/s	    1280 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]json_test.X-10-sonic
BenchmarkUnmarshalType/[]json_test.X-10-sonic-12        	  304773	      3432 ns/op	 245.04 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-lxt-12  	 1218081	       849.7 ns/op	 989.79 MB/s	     704 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.X-10-sonic-12         	  784118	      1634 ns/op	 514.54 MB/s	     970 B/op	       4 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-lxt
BenchmarkUnmarshalType/[]json_test.Y-10-lxt-12                   	  657778	      1706 ns/op	 422.75 MB/s	      80 B/op	       0 allocs/op
BenchmarkUnmarshalType/[]json_test.Y-10-sonic
BenchmarkUnmarshalType/[]json_test.Y-10-sonic-12                 	  398914	      2777 ns/op	 259.60 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-lxt-12           	 1330948	       904.4 ns/op	 797.23 MB/s	     587 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic
BenchmarkUnmarshalType/Marshal-[]json_test.Y-10-sonic-12         	  938292	      1077 ns/op	 669.50 MB/s	     845 B/op	       4 allocs/op
BenchmarkUnmarshalType/*int-10-lxt
BenchmarkUnmarshalType/*int-10-lxt-12                            	 4024768	       275.2 ns/op	 476.04 MB/s	      79 B/op	       0 allocs/op
BenchmarkUnmarshalType/*int-10-sonic
BenchmarkUnmarshalType/*int-10-sonic-12                          	 2532660	       502.6 ns/op	 260.64 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-lxt
BenchmarkUnmarshalType/Marshal-*int-10-lxt-12                    	 6856232	       169.5 ns/op	 773.02 MB/s	     113 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*int-10-sonic
BenchmarkUnmarshalType/Marshal-*int-10-sonic-12                  	 3795324	       322.5 ns/op	 406.14 MB/s	     213 B/op	       4 allocs/op
BenchmarkUnmarshalType/*bool-10-lxt
BenchmarkUnmarshalType/*bool-10-lxt-12                           	 4886638	       246.1 ns/op	 613.60 MB/s	      10 B/op	       0 allocs/op
BenchmarkUnmarshalType/*bool-10-sonic
BenchmarkUnmarshalType/*bool-10-sonic-12                         	 2474887	       431.2 ns/op	 350.15 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-lxt
BenchmarkUnmarshalType/Marshal-*bool-10-lxt-12                   	 9204621	       117.9 ns/op	1281.13 MB/s	     136 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*bool-10-sonic
BenchmarkUnmarshalType/Marshal-*bool-10-sonic-12                 	 4535430	       239.5 ns/op	 630.41 MB/s	     230 B/op	       4 allocs/op
BenchmarkUnmarshalType/*string-10-lxt
BenchmarkUnmarshalType/*string-10-lxt-12                         	 3473865	       310.4 ns/op	 712.09 MB/s	     159 B/op	       0 allocs/op
BenchmarkUnmarshalType/*string-10-sonic
BenchmarkUnmarshalType/*string-10-sonic-12                       	 2169774	       474.8 ns/op	 465.49 MB/s	       0 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-lxt
BenchmarkUnmarshalType/Marshal-*string-10-lxt-12                 	 7093699	       164.9 ns/op	1340.44 MB/s	     207 B/op	       0 allocs/op
BenchmarkUnmarshalType/Marshal-*string-10-sonic
BenchmarkUnmarshalType/Marshal-*string-10-sonic-12               	 2692192	       404.9 ns/op	 545.84 MB/s	     294 B/op	       4 allocs/op
```
由测试结果可知，针对不同 struct 成员类型，在性能上此 JSON 库你笨都比 [sonic](https://github.com/bytedance/sonic) 要好不少。

# 3. 持续优化
生命不息,折腾不止，作者将继续折腾。

# todo
当前存在的问题：
1. pointer、slice 的 cache 的 tag 的名字相同的的时候，会有冲突
2. slice 套 slice，pointer slice
3. 嵌套循环类型还未支持


# 交流学习
![扫码加微信好友](https://github.com/lxt1045/wechatbot/blob/main/resource/Wechat-lxt.png "微信")