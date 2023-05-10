# 目标是性能最好的 Go JSON 库


# 1. 常用的性能判断方法
由于笔者能力有限，这里并不会晚上的介绍所有优化方法，仅根据自己经验简单介绍 Go 下的优化工具的一种使用方式，其他使用方式需要读者自己探索。
## 1.1 pprof 工具使用：
参考文档:\
[golang 性能优化分析工具 pprof (上) - 基础使用介绍](https://www.cnblogs.com/jiujuan/p/14588185.html)\
[《Go 语言编程之旅》6.1 Go 大杀器之性能剖析 PProf（上）](https://golang2.eddycjy.com/posts/ch6/01-pprof-1/)

我们知道 Go 的 pprof 工具有三种使用方式：
runtime/pprof：采集程序（非 Server）的指定区块的运行数据进行分析。\
net/http/pprof：基于 HTTP Server 运行，并且可以采集运行时数据进行分析。\
go test：通过运行测试用例，并指定所需标识来进行采集。

这里仅介绍 go test 方式，分为这几个步骤：\
1. 这种方式需要先创建一个测试函数，最好是 Benchmark 类型。
```go
func Benchmark_Sample(b *testing.B) {
    for i := 0; i < b.N; i++ {
        f()
    }
}
```
2. 执行测试，生成 cpu.prof 测试文档
```sh
go test -benchmem -run=^$ -bench ^Benchmark_Sample$ github.com/lxt1045/json -count=1 -v -cpuprofile cpu.prof
```
3. 执行编译命令生成二进制
```sh
go test -benchmem -run=^$ -bench ^Benchmark_Sample$ github.com/lxt1045/json -c -o test.bin 
```
4. 使用 go tool 命令解析 cpu.prof 测试文档
```sh
go tool pprof ./test.bin cpu.prof
```
5. 使用以下命令查看:\
5.1
```sh
web # 查看 graph 图
```
5.2
```sh
top n # 查看占用排行
```
输出例子：
```
Showing nodes accounting for 9270ms, 67.91% of 13650ms total
Dropped 172 nodes (cum <= 68.25ms)
Showing top 10 nodes out of 116
      flat  flat%   sum%        cum   cum%
    2490ms 18.24% 18.24%     2490ms 18.24%  runtime.madvise
    1880ms 13.77% 32.01%     1880ms 13.77%  runtime.pthread_cond_signal
     970ms  7.11% 39.12%     1230ms  9.01%  [test.bin]
     920ms  6.74% 45.86%      940ms  6.89%  runtime.pthread_cond_wait
     720ms  5.27% 51.14%     2570ms 18.83%  github.com/lxt1045/json.parseObj
     640ms  4.69% 55.82%      640ms  4.69%  github.com/bytedance/sonic/internal/native/avx2.__native_entry__
     550ms  4.03% 59.85%      740ms  5.42%  github.com/lxt1045/json.(*tireTree).Get
     530ms  3.88% 63.74%      710ms  5.20%  runtime.scanobject
     300ms  2.20% 65.93%      300ms  2.20%  runtime.memmove
     270ms  1.98% 67.91%      270ms  1.98%  runtime.kevent
```
5.3
```sh
list func_name # 查看函数内每行代码开销
```
输出例子：
```sh
Total: 13.65s
ROUTINE ======================== github.com/lxt1045/json.(*tireTree).Get in /Users/bytedance/go/src/github.com/lxt1045/json/tire_tree.go
     550ms      740ms (flat, cum)  5.42% of Total
         .          .    282:           return nil
         .          .    283:   }
         .          .    284:
         .          .    285:   return nil
         .          .    286:}
      30ms       30ms    287:func (root *tireTree) Get(key string) *TagInfo {
      10ms       10ms    288:   status := &root.tree[0]
         .          .    289:   // for _, c := range []byte(key) {
      20ms       20ms    290:   for i := 0; i < len(key); i++ {
      10ms       10ms    291:           c := key[i]
         .          .    292:           k := c & 0x7f
     160ms      160ms    293:           next := status[k]
         .          .    294:           if next.next >= 0 {
         .          .    295:                   i += int(next.skip)
      10ms       10ms    296:                   status = &root.tree[next.next]
         .          .    297:                   continue
         .          .    298:           }
      10ms       10ms    299:           if next.idx >= 0 {
      40ms       40ms    300:                   tag := root.tags[next.idx]
     250ms      440ms    301:                   if len(key) > len(tag.TagName) && key[len(tag.TagName)] == '"' && tag.TagName == key[:len(tag.TagName)] {
      10ms       10ms    302:                           return tag
         .          .    303:                   }
         .          .    304:           }
         .          .    305:           return nil
         .          .    306:   }
         .          .    307:
```
5.4 
```sh
go tool pprof -http=:8080 cpu.prof # 通过浏览器查看测试结果
```
执行后，通过浏览器打开 http://localhost:8080/ 链接就可以查看了。

# 2. json 库的特点和针对应的优化方案
3. lxt1045/json 的优化方案
3.1 
3.2 

我们知道 golang 的反射性能

采用了哪些优化方案
##  1. 反射
我们知道golang 的反射性能是比较差的，只要是因为反射的时候需要生成一个逃逸的对象。

但是，因为生成新对象的时候，需要配合 GC 做内存标注，所以必须使用

## 缓存

RCU

针对指针 通过 reflect.struct 动态生成 struct

## 字符串搜索

## 