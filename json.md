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
list func_name # 查看函数内每行代码开销; 注意 '(' 、')'、'*' 需要转义
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

## 附
1. net/http/pprof：基于 HTTP Server 运行，并且可以采集运行时数据进行分析。

1.1 在main package 中加入以下代码：
```go
//main.go
import (
	"net/http"
	_ "net/http/pprof"
)
func main() {
	go func() {
		runtime.SetBlockProfileRate(1)     // 开启对阻塞操作的跟踪，block
		runtime.SetMutexProfileFraction(1) // 开启对锁调用的跟踪，mutex

		err := http.ListenAndServe(":6060", nil)
		stdlog.Fatal(err)
	}()
}
```
就可以通过 http://127.0.0.1:6060/debug/pprof/ url访问相关信息：
```sh
/debug/pprof/
Set debug=1 as a query parameter to export in legacy text format


Types of profiles available:
Count	Profile
3	allocs
2	block
0	cmdline
10	goroutine
3	heap
0	mutex
0	profile
10	threadcreate
0	trace
full goroutine stack dump
Profile Descriptions:

allocs: A sampling of all past memory allocations
block: Stack traces that led to blocking on synchronization primitives
cmdline: The command line invocation of the current program
goroutine: Stack traces of all current goroutines. Use debug=2 as a query parameter to export in the same format as an unrecovered panic.
heap: A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.
mutex: Stack traces of holders of contended mutexes
profile: CPU profile. You can specify the duration in the seconds GET parameter. After you get the profile file, use the go tool pprof command to investigate the profile.
threadcreate: Stack traces that led to the creation of new OS threads
trace: A trace of execution of the current program. You can specify the duration in the seconds GET parameter. After you get the trace file, use the go tool trace command to investigate the trace.
```

通过以下命令可以获取相关的pprof文件：
```sh
# CPU
curl -o cpu.out http://127.0.0.1:6060/debug/pprof/profile?seconds=20
# memery
curl -o cpu.out http://127.0.0.1:6060/debug/pprof/allocs?seconds=30
# 
curl -o cpu.out http://127.0.0.1:6060/debug/pprof/mutex?seconds=15
# 
curl -o cpu.out http://127.0.0.1:6060/debug/pprof/block?seconds=15
```

其中 net/http/pprof 使用 runtime/pprof 包来进行封装，并在 http 端口上暴露出来。runtime/pprof 可以用来产生 dump 文件，再使用 Go Tool PProf 来分析这运行日志。

如果应用使用了自定义的 Mux，则需要手动注册一些路由规则：
```go
r.HandleFunc("/debug/pprof/", pprof.Index)
r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
r.HandleFunc("/debug/pprof/profile", pprof.Profile)
r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
r.HandleFunc("/debug/pprof/trace", pprof.Trace)
```
```go
func RouteRegister(rg *gin.RouterGroup, prefixOptions ...string) {
   prefix := getPrefix(prefixOptions...)

   prefixRouter := rg.Group(prefix)
   {
      prefixRouter.GET("/", pprofHandler(pprof.Index))
      prefixRouter.GET("/cmdline", pprofHandler(pprof.Cmdline))
      prefixRouter.GET("/profile", pprofHandler(pprof.Profile))
      prefixRouter.POST("/symbol", pprofHandler(pprof.Symbol))
      prefixRouter.GET("/symbol", pprofHandler(pprof.Symbol))
      prefixRouter.GET("/trace", pprofHandler(pprof.Trace))
      prefixRouter.GET("/allocs", pprofHandler(pprof.Handler("allocs").ServeHTTP))
      prefixRouter.GET("/block", pprofHandler(pprof.Handler("block").ServeHTTP))
      prefixRouter.GET("/goroutine", pprofHandler(pprof.Handler("goroutine").ServeHTTP))
      prefixRouter.GET("/heap", pprofHandler(pprof.Handler("heap").ServeHTTP))
      prefixRouter.GET("/mutex", pprofHandler(pprof.Handler("mutex").ServeHTTP))
      prefixRouter.GET("/threadcreate", pprofHandler(pprof.Handler("threadcreate").ServeHTTP))
   }
}
```

其它的数据的分析和CPU、Memory基本一致。下面列一下所有的数据类型：

http://localhost:6060/debug/pprof/ ：获取概况信息，即图一的信息
http://localhost:6060/debug/pprof/allocs : 分析内存分配
http://localhost:6060/debug/pprof/block : 分析堆栈跟踪导致阻塞的同步原语
http://localhost:6060/debug/pprof/cmdline : 分析命令行调用的程序，web下调用报错
http://localhost:6060/debug/pprof/goroutine : 分析当前 goroutine 的堆栈信息
http://localhost:6060/debug/pprof/heap : 分析当前活动对象内存分配
http://localhost:6060/debug/pprof/mutex : 分析堆栈跟踪竞争状态互斥锁的持有者
http://localhost:6060/debug/pprof/profile : 分析一定持续时间内CPU的使用情况
http://localhost:6060/debug/pprof/threadcreate : 分析堆栈跟踪系统新线程的创建
http://localhost:6060/debug/pprof/trace : 分析追踪当前程序的执行状况

