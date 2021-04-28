熔断器像是一个保险丝。当我们依赖的服务出现问题时，可以及时容错。一方面可以减少依赖服务对自身访问的依赖，防止出现雪崩效应；另一方面降低请求频率以方便上游尽快恢复服务。

## 熔断器模式

熔断器有三种状态，四种状态转移的情况：

三种状态：
- 熔断器关闭状态, 服务正常访问
- 熔断器开启状态，服务异常
- 熔断器半开状态，部分请求，验证是否可以访问


四种状态转移：
- 在熔断器关闭状态下，当失败后并满足一定条件后，将直接转移为熔断器开启状态。
- 在熔断器开启状态下，如果过了规定的时间，将进入半开启状态，验证目前服务是否可用。
- 在熔断器半开启状态下，如果出现失败，则再次进入关闭状态。
- 在熔断器半开启后，所有请求（有限额）都是成功的，则熔断器关闭。所有请求将正常访问。
![](https://blog.lpflpf.cn/passages/circuit-breaker/state_machine.png)
## gobreaker
```go
type CircuitBreaker struct {
  name          string
  maxRequests   uint32  // 最大请求数 （半开启状态会限流）
  interval      time.Duration   // 统计周期
  timeout       time.Duration   // 进入熔断后的超时时间
  readyToTrip   func(counts Counts) bool // 通过Counts 判断是否开启熔断。需要自定义
  onStateChange func(name string, from State, to State) // 状态修改时的钩子函数

  mutex      sync.Mutex // 互斥锁，下面数据的更新都需要加锁
  state      State  // 记录了当前的状态
  generation uint64 // 标记属于哪个周期
  counts     Counts // 计数器，统计了 成功、失败、连续成功、连续失败等，用于决策是否进入熔断
  expiry     time.Time // 进入下个周期的时间
}
```
熔断器的执行操作，主要包括三个阶段；①请求之前的判定；②服务的请求执行；③请求后的状态和计数的更新

```go
// 熔断器的调用
func (cb *CircuitBreaker) Execute(req func() (interface{}, error)) (interface{}, error) {

  // ①请求之前的判断
  generation, err := cb.beforeRequest()
  if err != nil {
    return nil, err
  }

  defer func() {
    e := recover()
    if e != nil {
      // ③ panic 的捕获
      cb.afterRequest(generation, false)
      panic(e)
    }
  }()

  // ② 请求和执行
  result, err := req()

  // ③ 更新计数
  cb.afterRequest(generation, err == nil)
  return result, err
}
```

总结
- 对于频繁请求一些远程或者第三方的不可靠的服务，存在失败的概率还是非常大的。使用熔断器的好处就是可以是我们自身的服务不被这些不可靠的服务拖垮，造成雪崩。
- 由于熔断器里面，不仅会维护不少的统计数据，还有互斥锁做资源隔离，成本也会不少。
- 在半开状态下，可能出现请求过多的情况。这是由于半开状态下，连续请求成功的数量未达到最大请求值。所以，熔断器对于请求时间过长（但是比较频繁）的服务可能会造成大量的 too many requests 错误