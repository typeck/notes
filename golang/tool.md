# Pprof

**性能分析基础数据的获取有三种方式：**
1) runtime/pprof 包
2) net/http/pprof 包
3) go test 时添加收集参数

```go
import(_ "net/http/pprof")
```
访问 http://127.0.0.1:8080/debug/pprof/  即可实时查看性能数据，业面如下：
```
/debug/pprof/
 
Types of profiles available:
Count Profile
2 allocs       # 所有过去内存分析采样
0 block        # 导致同步原语阻塞的堆栈跟踪
0 cmdline      # 程序启动参数
4 goroutine    # 所有当前goroutine堆栈跟踪
2 heap         # 活动对象内存分配采样
0 mutex        # 互斥锁跟踪
0 profile      # 生成cpuprofile文件 生成文件使用go tool pprof工具分析
8 threadcreate # 创建系统线程的堆栈跟踪
0 trace        # 对当前程序执行的跟踪 生成文件使用go tool trace工具分析
full goroutine stack dump  # 显示所有goroutine的堆栈
```
对生成的profile文件，可使用go tool pprof profilename来分析，例如：
`go tool pprof http://localhost:8080/debug/pprof/heap`

`wget 'http://localhost:6060/debug/pprof/goroutine?debug=1' -O /tmp/goroutine.log`

**web可视化方式**

提前安装 Graphviz 用于画图
下载地址：https://graphviz.gitlab.io/download/

go tool pprof -http=:8000 http://localhost:8080/debug/pprof/heap    查看内存使用

go tool pprof -http=:8000 http://localhost:8080/debug/pprof/profile 查看cpu占用

访问地址：http://localhost:8000/ui/


参考：
[Golang性能分析工具PProf的使用/Go GC监控](https://pdf.us/2019/02/18/2772.html)