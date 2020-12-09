## slice 底层实现
**Array**

在 Go 中，与 C 数组变量隐式作为指针使用不同，Go 数组是值类型，赋值和函数传参操作都会复制整个数组数据。

**切片的数据结构**

切片本身并不是动态数组或者数组指针。它内部实现的数据结构通过指针引用底层数组，设定相关属性将数据读写操作限定在指定的区域内。切片本身是一个只读对象，其工作机制类似数组指针的一种封装。

切片（slice）是对数组一个连续片段的引用，所以切片是一个引用类型（因此更类似于 C/C++ 中的数组类型，或者 Python 中的 list 类型）。这个片段可以是整个数组，或者是由起始和终止索引标识的一些项的子集。需要注意的是，终止索引标识的项不包括在切片内。切片提供了一个与指向数组的动态窗口。

slice的数据结构：
```go
type slice struct {
    array unsafe.Pointer
    len   int
    cap   int
}
```
Pointer 是指向一个数组的指针，len 代表当前切片的长度，cap 是当前切片的容量。cap 总是大于等于 len 的。

手动构造一个slice：
```go
var ptr unsafe.Pointer
var s1 = struct {
    addr uintptr
    len int
    cap int
}{ptr, length, length}
s := *(*[]byte)(unsafe.Pointer(&s1))
```
make 函数允许在运行期动态指定数组长度，绕开了数组类型必须使用编译期常量的限制。

Go 中切片扩容的策略是这样的：

如果切片的容量小于 1024 个元素，于是扩容的时候就翻倍增加容量。上面那个例子也验证了这一情况，总容量从原来的4个翻倍到现在的8个。

一旦元素个数超过 1024 个元素，那么增长因子就变成 1.25 ，即每次增加原来容量的四分之一。

## channel

<-chan 和 chan定义的区别：
```go
type Ticker struct {
	//c can only receive data
    C  <-chan time.Time 
	C1 chan time.Time
}
```

chan 只读不写或者只写不读会panic。

在写的同时如果读的不及时，可以在写时加入default丢弃数据，而保证不阻塞。
```go
func test4() {
	var ch = make(chan int)
	go func() {
		time.Sleep(time.Second)
		select {
		case ch <- 1:
			fmt.Println("send")
		default:
			fmt.Println("send default")
		}
	}()
	//time.Sleep(time.Second)
	select {
	case t := <-ch:
		fmt.Println("receive:", t)
	default:
		fmt.Println("receive default")
	}
	time.Sleep(time.Second)
}
```

## golang内存模型
**进程**：进程是系统进行资源分配的基本单位，有独立的内存空间。

**线程**：线程是cpu调度和分派的最小单位，有独立的内存空间。

**协程**：协程是一种用户态的轻量级线程，协程的调度完全由用户控制，协程间的切换只需要保存任务的上下文，没有内核开销。

**虚拟内存大小（vsz)**：是进程可以访问的所有内存，包括换出的内存、分配但未使用的内存和共享库中的内存。

**驻留集大小（rss）**：是进程在实际内存中的内存页数乘以内存页大小，这里不包括换出的内存页。

go语言不使用malloc来获取内存，而是通过操作系统申请（mmap），基于[TCMalloc](http://goog-perftools.sourceforge.net/doc/tcmalloc.html)实现内存的分配和释放。

### TCMalloc(Thread-Caching Malloc)

高效的多线程内存管理，用于替代系统的内存分配函数。

**小内存管理：**

对于256KB以内的小对象分配，TCMalloc按大小划分了85个类别,称为Size Class，每个size class都对应一个大小，比如8字节，16字节，32字节。应用程序申请内存时，TCMalloc会首先将所申请的内存大小向上取整到size class的大小，比如1~8字节之间的内存申请都会分配8字节，9~16字节之间都会分配16字节，以此类推。因此这里会产生内部碎片。TCMalloc将这里的内部碎片控制在12.5%以内。

对于每个线程，TCMalloc都为其保存了一份单独的缓存，称之为ThreadCache。每个ThreadCache中对于每个size class都有一个单独的FreeList，缓存了n个还未被应用程序使用的空闲对象。

![img](img/v2-39c1586740e79d9dcfa1fc1a42148b68_720w.jpg)

小对象的分配直接从ThreadCache的FreeList中返回一个空闲对象，相应的，小对象的回收也是将其重新放回ThreadCache中对应的FreeList中。

一旦线程本地缓存耗尽空间，内存对象就会从中心数据结构移动到线程本地缓存。
![img](./img/v2-86ed188a599609fb09469992dec7bc20_720w.jpg)

与ThreadCache类似，CentralCache中对于每个size class也都有一个单独的链表来缓存空闲对象，称之为CentralFreeList，供各线程的ThreadCache从中取用空闲对象。

当CentralCache中的空闲对象不够用时，CentralCache会向PageHeap申请一块内存（可能来自PageHeap的缓存，也可能向系统申请新的内存），并将其拆分成一系列空闲对象，添加到对应size class的CentralFreeList中。

PageHeap内部根据内存块（span）的大小采取了两种不同的缓存策略。128个page以内的span，每个大小都用一个链表来缓存，超过128个page的span，存储于一个有序set（std::set）
![img](img/v2-a2dc32a2f4f4016bdb73467394295e89_720w.jpg)

应用程序调用free()或delete一个小对象时，仅仅是将其插入到ThreadCache中其size class对应的FreeList中而已，不需要加锁，因此速度也是非常快的。

只有当满足一定的条件时，ThreadCache中的空闲对象才会重新放回CentralCache中，以供其他线程取用。同样的，当满足一定条件时，CentralCache中的空闲对象也会还给PageHeap，PageHeap再还给系统。

小对象分配流程大致如下：

* 将要分配的内存大小映射到对应的size class。
* 查看ThreadCache中该size class对应的FreeList。
* 如果FreeList非空，则移除FreeList的第一个空闲对象并将其返回，分配结束。
* 如果FreeList是空的：
* 从CentralCache中size class对应的CentralFreeList获取一堆空闲对象。
    * 如果CentralFreeList也是空的，则：
    * 向PageHeap申请一个span。
    * 拆分成size class对应大小的空闲对象，放入CentralFreeList中。
* 将这堆对象放置到ThreadCache中size class对应的FreeList中（第一个对象除外）。
* 返回从CentralCache获取的第一个对象，分配结束。

**中对象管理**

超过256KB但不超过1MB（128个page）的内存分配被认为是中对象分配，采取了与小对象不同的分配策略。

首先，TCMalloc会将应用程序所要申请的内存大小向上取整到整数个page（因此，这里会产生1B~8KB的内部碎片）。之后的操作表面上非常简单，向PageHeap申请一个指定page数量的span并返回其起始地址即可：
```c++
Span* span = Static::pageheap()->New(num_pages);
result = (PREDICT_FALSE(span == NULL) ? NULL : SpanToMallocResult(span));
return result;
```
PageHeap提供了一层缓存，因此PageHeap::New()并非每次都向系统申请内存，也可能直接从缓存中分配。

对128个page以内的span和超过128个page的span，PageHeap采取的缓存策略不一样。为了描述方便，以下将128个page以内的span称为小span，大于128个page的span称为大span。

假设要分配一块内存，其大小经过向上取整之后对应k个page，因此需要从PageHeap取一个大小为k
个page的span，过程如下：

* 从k个page的span链表开始，到128个page的span链表，按顺序找到第一个非空链表。
* 取出这个非空链表中的一个span，假设有n个page，将这个span拆分成两个span：
* 一个span大小为k个page，作为分配结果返回。
* 另一个span大小为n - k个page，重新插入到n - k个page的span链表中。
* 如果找不到非空链表，则将这次分配看做是大对象分配，分配过程详见下文。

**大对象管理**

超过1MB（128个page）的内存分配被认为是大对象分配，与中对象分配类似，也是先将所要分配的内存大小向上取整到整数个page，假设是k个page，然后向PageHeap申请一个k个page大小的span。

大对象分配用到的span都是超过128个page的span，其缓存方式不是链表，而是一个按span大小排序的有序set（std::set），以便按大小进行搜索。

假设要分配一块超过1MB的内存，其大小经过向上取整之后对应k个page（k>128），或者是要分配一块1MB以内的内存，但无法由中对象分配逻辑来满足，此时k <= 128。不管哪种情况，都是要从PageHeap的span set中取一个大小为k个page的span，其过程如下：

* 搜索set，找到不小于k个page的最小的span（best-fit），假设该span有n个page。
* 将这个span拆分为两个span：
* 一个span大小为k个page，作为结果返回。
* 另一个span大小为n - k个page，如果n - k > 128，则将其插入到大span的set中，否则，将其插入到对应的小span链表中。
* 如果找不到合适的span，则使用sbrk或mmap向系统申请新的内存以生成新的span，并重新执行中对象或大对象的分配算法。
![img](img/v2-7f6862001e54df832ba538e0a0dad3ba_720w.jpg)

### go内存管理

大对象专门分配，小对象以8的倍数为单位，分成66种规格。
```go
// class  bytes/obj  bytes/span  objects  tail waste  max waste
//     1          8        8192     1024           0     87.50%
//     2         16        8192      512           0     43.75%
//     3         32        8192      256           0     46.88%
//     4         48        8192      170          32     31.52%
//     5         64        8192      128           
// 		...
//    65      28672       57344        2           0      4.91%
//    66      32768       32768        1           0     12.50%
```
![img](img/2020-09-09-105534.png)

顶层堆管理部件（heap）每次向操作系统申请一大块内存（最少1MB），还负责管理未使用的大块内存（span），为大对象直接分配空间。

central从堆提取大块内存，没个central只负责一种规格，不同规格的请求会被分配到不同的中间部件，减小锁粒度。
在central向heap请求时，会按size class静态表设置大小进行分割，在回收时尝试和相邻未使用的span合并，以形成更大可切分空间。




参考：

[1] [TCMalloc解密](https://wallenwang.com/2018/11/tcmalloc/)

[2] [Go 语言的内存管理](https://www.zhihu.com/people/tang-xu-83-53)

## go context

Go 语言中的每一个请求的都是通过一个单独的 Goroutine 进行处理的，HTTP/RPC 请求的处理器往往都会启动新的 Goroutine 访问数据库和 RPC 服务，我们可能会创建多个 Goroutine 来处理一次请求，而 Context 的主要作用就是在不同的 Goroutine 之间同步请求特定的数据、取消信号以及处理请求的截止日期。

当最上层的 Goroutine 因为某些原因执行失败时，下两层的 Goroutine 由于没有接收到这个信号所以会继续工作；但是当我们正确地使用 Context 时，就可以在下层及时停掉无用的工作减少额外资源的消耗.

这其实就是 Golang 中上下文的最大作用，在不同 Goroutine 之间对信号进行同步避免对计算资源的浪费，与此同时 Context 还能携带以请求为作用域的键值对信息。

**Context使用原则**

* context.Background 只应用在最高等级，作为所有派生 context 的根。
* context 取消是建议性的，这些函数可能需要一些时间来清理和退出。
* 不要把Context放在结构体中，要以参数的方式传递。
* 以Context作为参数的函数方法，应该把Context作为第一个参数，放在第一位。
* 给一个函数方法传递Context的时候，不要传递nil，如果不知道传递什么，就使用context.TODO
* Context的Value相关方法应该传递必须的数据，不要什么数据都使用这个传递。context.Value 应该很少使用，它不应该被用来传递可选参数。这使得 API 隐式的并且可以引起错误。取而代之的是，这些值应该作为参数传递。
* Context是线程安全的，可以放心的在多个goroutine中传递。同一个Context可以传给使用其的多个goroutine，且Context可被多个goroutine同时安全访问。
* Context 结构没有取消方法，因为只有派生 context 的函数才应该取消 context。

## 延迟语句

defer的语义表明，它会在函数返回、产生恐慌或者`runtime.Goexit`时使用，直觉上看， defer 应该由编译器直接将需要的函数调用插入到该调用的地方，似乎是一个编译期特性， 不应该存在运行时性能问题，非常类似于 C++ 的 RAII(Resource Acquisition is Initialization) 范式（当离开资源的作用域时， 自动执行析构函数）。但实际情况是，由于 defer 并没有与其依赖资源挂钩，也允许在条件、循环语句中出现， 从而不再是一个作用域相关的概念，这就是使得 defer 的语义变得相对复杂。 在一些复杂情况下，无法在编译期决定存在多少个 defer 调用。

对于延迟语句而言，其中间表示会产生三种不同的延迟形式， 第一种是最一般情况下的在堆上分配的延迟语句，第二种是允许在栈上分配的延迟语句， 最后一种则是 开放编码式(Open-coded) 的延迟语句。
```go
// src/cmd/compile/internal/gc/ssa.go
func (s *state) stmt(n *Node) {
	...
	switch n.Op {
	case ODEFER:
		// 开放编码式 defer
		if s.hasOpenDefers {
			s.openDeferRecord(n.Left)
		} else {
			// 堆上分配的 defer
			d := callDefer
			if n.Esc == EscNever {
				// 栈上分配的 defer
				d = callDeferStack
			}
			s.call(n.Left, d)
		}
	case ...
	}
	...
}
```
* 在堆上分配的原因是 defer 语句出现 在了循环语句里，或者无法执行更高阶的编译器优化导致的。如果一个与 defer 出现在循环语句中， 则可执行的次数可能无法在编译期决定；如果一个调用中 defer 由于数量过多等原因， 不能被编译器进行开放编码，则也会在堆上分配 defer。在堆上分配的 defer 需要最多的运行时支持， 因而产生的运行时开销也最大。
  
  一个函数中的延迟语句会被保存为一个 _defer 记录的链表，附着在一个 Goroutine 上。 _defer 记录的具体结构也非常简单，主要包含了参与调用的参数大小、 当前 defer 语句所在函数的 PC 和 SP 寄存器、被 defer 的函数的入口地址以及串联 多个 defer 的 link 链表，该链表指向下一个需要执行的 defer。
  ```golang
	// src/runtime/panic.go
	type _defer struct {
		siz       int32
		heap      bool
		sp        uintptr
		pc        uintptr
		fn        *funcval
		link      *_defer
		...
	}
	// src/runtime/runtime2.go
	type g struct {
		...
		_defer *_defer
		...
	}
  ```

* defer 还可以直接在栈上进行分配，在栈上分配 defer 的好处在于函数返回后 _defer 便已得到释放， 不再需要考虑内存分配时产生的性能开销，只需要适当的维护 _defer 的链表即可。
* 一定情况下能够让 defer 进化为一个仅编译期特性，即在函数末尾直接对延迟函数进行调用， 做到几乎不需要额外的开销。
  * 没有禁用编译器优化，即没有设置 -gcflags "-N"
  * 存在 defer 调用
  * 函数内 defer 的数量不超过 8 个、且返回语句与延迟语句个数的乘积不超过 15
  * 没有与 defer 发生在循环语句中
  
  延迟比特

  如果一个 defer 发生在一个条件语句中，而这个条件必须等到运行时才能确定：
  ```golang
  if rand.Intn(100) < 42 {
	defer fmt.Println("meaning-of-life")
  }
  ```
  如何才能使用最小的成本，让插入到函数末尾的延迟语句，在条件成立时候被正确执行呢？ 这便需要一种机制，能够记录存在延迟语句的条件分支是否被执行， 这种机制在 Go 中利用了延迟比特（defer bit）。
  
  使用延迟比特的核心思想可以用下面的伪代码来概括。 在创建延迟调用的阶段，首先通过延迟比特的特定位置记录哪些带条件的 defer 被触发。 这个延迟比特是一个长度为 8 位的二进制码（也是硬件架构里最小、最通用的情况）， 以每一位是否被设置为 1，来判断延迟语句是否在运行时被设置，如果设置，则发生调用。 否则则不调用：
  
  在退出位置，再重新根据被标记的延迟比特，反向推导哪些位置的 defer 需要被触发，从而 执行延迟调用：
  ```golang
  defer f1(a1)
  if cond {
	defer f2(a2)
  }

	//伪代码
  deferBits = 0           // 初始值 00000000
  deferBits |= 1 << 0     // 遇到第一个 defer，设置为 00000001
  _f1 = f1
  _a1 = a1
  if cond {
	// 如果第二个 defer 被设置，则设置为 00000011，否则依然为 00000001
	deferBits |= 1 << 1
	_f2 = f2
	_a2 = a2
  }


  exit:
  // 按顺序倒序检查延迟比特。如果第二个 defer 被设置，则
  //   00000011 & 00000010 == 00000010，即延迟比特不为零，应该调用 f2。
  // 如果第二个 defer 没有被设置，则 
  //   00000001 & 00000010 == 00000000，即延迟比特为零，不应该调用 f2。
  if deferBits & 1 << 1 != 0 { // 00000011 & 00000010 != 0
	deferBits &^= 1<<1       // 00000001
	_f2(_a2)
  }
  // 同理，由于 00000001 & 00000001 == 00000001，因此延迟比特不为零，应该调用 f1
  if deferBits && 1 << 0 != 0 {
	deferBits &^= 1<<0
	_f1(_a1)
  }
  ```

## 同步

**自旋锁**

自旋锁与互斥锁类似，自旋锁不会引起调用者的睡眠，如果自旋锁已经被其他执行单元保持，调用者就一直循环在那里看是否自旋锁的保持者已经释放了锁。其作用是为了解决某项资源的互斥使用。因为自旋锁不会引起调用者睡眠，所以自旋锁的效率远 高于互斥锁。

自旋锁一直占用CPU，他在未获得锁的情况下，一直运行。

在用自旋锁时有可能造成死锁，当递归调用时有可能造成死锁，调用有些其他函数也可能造成死锁。

自旋锁适用于锁使用者保持锁时间比较短的情况下。

**互斥锁**

互斥锁属于sleep-waiting类型的锁。例如在一个双核的机器上有两个线程(线程A和线程B)，它们分别运行在Core0和 Core1上。假设线程A想要通过pthread_mutex_lock操作去得到一个临界区的锁，而此时这个锁正被线程B所持有，那么线程A就会被阻塞 (blocking)，Core0 会在此时进行上下文切换(Context Switch)将线程A置于等待队列中，此时Core0就可以运行其他的任务(例如另一个线程C)而不必进行忙等待。

而自旋锁则不然，它属于busy-waiting类型的锁，如果线程A是使用pthread_spin_lock操作去请求锁，那么线程A就会一直在 Core0上进行忙等待并不停的进行锁请求，直到得到这个锁为止。

互斥锁：线程会从sleep（加锁）——>running（解锁），过程中有上下文的切换，cpu的抢占，信号的发送等开销。

自旋锁：线程一直是running(加锁——>解锁)，死循环检测锁的标志位，机制不复杂。


**条件变量**

https://zhuanlan.zhihu.com/p/75220465

```golang
type Mutex struct {
	state int32  // 表示 mutex 锁当前的状态
	//goroutine会在这个值上设置各种标志位，用于获取这个锁的goroutine通信，（cas)
	sema  uint32 // 信号量，用于唤醒 goroutine
	//通过runtime_Semrelease唤醒一个使用Semacquire阻塞在这个信号量的goroutine
}
```
```go
const (
	mutexLocked = 1 << iota // 互斥锁已锁住
	mutexWoken
	mutexStarving
	mutexWaiterShift = iota
	starvationThresholdNs = 1e6
)
```
Mutex 可能处于两种不同的模式：正常模式和饥饿模式。

在正常模式中，等待者按照 FIFO 的顺序排队获取锁，但是一个被唤醒的等待者有时候并不能获取 mutex， 它还需要和新到来的 goroutine 们竞争 mutex 的使用权。 新到来的 goroutine 存在一个优势，它们已经在 CPU 上运行且它们数量很多， 因此一个被唤醒的等待者有很大的概率获取不到锁，在这种情况下它处在等待队列的前面。 如果一个 goroutine 等待 mutex 释放的时间超过 1ms，它就会将 mutex 切换到饥饿模式

在饥饿模式中，mutex 的所有权直接从解锁的 goroutine 递交到等待队列中排在最前方的 goroutine。 新到达的 goroutine 们不要尝试去获取 mutex，即使它看起来是在解锁状态，也不要试图自旋， 而是排到等待队列的尾部。

如果一个等待者获得 mutex 的所有权，并且看到以下两种情况中的任一种：

1) 它是等待队列中的最后一个，
2) 它等待的时间少于 1ms，它便将 mutex 切换回正常操作模式

正常模式下的性能会更好，因为一个 goroutine 能在即使有很多阻塞的等待者时多次连续的获得一个 mutex，饥饿模式的重要性则在于避免了病态情况下的尾部延迟。

**条件变量**

sync.Cond 的内部结构包含一个锁（Locker）、通知列表（notifyList）以及一个复制检查器 copyChecker。

```go
type Locker interface {
	Lock()
	Unlock()
}
type Cond struct {
	L Locker

	notify  notifyList
	checker copyChecker
}
func NewCond(l Locker) *Cond {
	return &Cond{L: l}
}
```
copychecker
```go
// copyChecker 保存指向自身的指针来检测对象的复制行为。
type copyChecker uintptr

func (c *copyChecker) check() {
	if uintptr(*c) != uintptr(unsafe.Pointer(c)) &&
		!atomic.CompareAndSwapUintptr((*uintptr)(c), 0, uintptr(unsafe.Pointer(c))) &&
		uintptr(*c) != uintptr(unsafe.Pointer(c)) {
		panic("sync.Cond is copied")
	}
}

type cond struct {
	c copyChecker
}

func test15() {
	var c cond
	c.c = copyChecker(unsafe.Pointer(&c))
	var c1 = c
	c1.c.check()
}
```
Wait / Signal / Broadcast

Wait/Signal/Broadcast 都是由通知列表来实现的，撇开 copyChecker， Wait 无非就是向 notifyList 注册一个通知，而后阻塞到被通知， Signal 则负责通知一个在 notifyList 注册过的 waiter 发出通知， Broadcast 更是直接粗暴的向所有人都发出通知。

核心是notifyList，本质上是一个队列
```go
// notifyList 基于 ticket 实现通知列表
type notifyList struct {
	// wait 为下一个 waiter 的 ticket 编号
	// 在没有 lock 的情况下原子自增
	wait uint32

	// notify 是下一个被通知的 waiter 的 ticket 编号
	// 它可以在没有 lock 的情况下进行读取，但只有在持有 lock 的情况下才能进行写
	//
	// wait 和 notify 会产生 wrap around，只要它们 "unwrapped"
	// 的差别小于 2^31，这种情况可以被正确处理。对于 wrap around 的情况而言，
	// 我们需要超过 2^31+ 个 goroutine 阻塞在相同的 condvar 上，这是不可能的。
	//
	notify uint32
	// waiter 列表.
	lock mutex
	head *sudog
	tail *sudog
}
```
如果l.wait == l.notify 则表示 上次通知后没有新的 waiter， 可以直接返回。

**同步组**

sync.WaitGroup 可以达到并发 Goroutine 的执行屏障的效果，等待多个 Goroutine 执行完毕。

```go
// WaitGroup 用于等待一组 Goroutine 执行完毕。
// 主 Goroutine 调用 Add 来设置需要等待的 Goroutine 的数量
// 然后每个 Goroutine 运行并调用 Done 来确认已经执行网完毕
// 同时，Wait 可以用于阻塞并等待所有 Goroutine 完成。
//
// WaitGroup 在第一次使用后不能被复制
type WaitGroup struct {
	// 64 位值: 高 32 位用于计数，低 32 位用于等待计数
	// 64 位的原子操作要求 64 位对齐，但 32 位编译器无法保证这个要求
	// 因此分配 12 字节然后将他们对齐，其中 8 字节作为状态，其他 4 字节用于存储原语
	state1 [3]uint32
}
```

在32位平台上对64的uint操作不可能是原子的，要分两个指令取数，比如在读取一个字长度的时候，另外一个字的数据很有可能已经发生改变了(在32位操作系统上,字长是4，而uint64长度为8)。所以在实际计数的时候，其实sync.WaitGroup也就使用了4个字节来进行。

```go
// state 返回 wg.state1 中存储的状态和原语字段
func (wg *WaitGroup) state() (statep *uint64, semap *uint32) {
	if uintptr(unsafe.Pointer(&wg.state1))%8 == 0 {
		return (*uint64)(unsafe.Pointer(&wg.state1)), &wg.state1[2]
	}
	return (*uint64)(unsafe.Pointer(&wg.state1[1])), &wg.state1[0]
}
```
如果起始地址不能被8整除，那么偏移一位 &WaitGroup.State1[1]一定能被8整除。

todo: (32位机器一定可以整除，64位一定不可以么？)

**补充：内存对齐**
1) 结构体变量的起始地址能够被其最宽的成员大小整除
2) 结构体每个成员相对于起始地址的偏移能够被其自身大小整除，如果不能则在前一个成员后面补充字节
3) 结构体总体大小能够被最宽的成员的大小整除，如不能则在后面补充字节


Add 将 statep 的值作为两段来处理，前 32 位处理为计数器，后 32 位处理为等待器。

- 在初始阶段，等待器为 0 ，计数器随着 Add 正数的调用而增加。
- 如果 Add 使用错误导致计数器为负，则会立即 panic
- 由于并发的效果，计数器和等待器的值是分开操作的，因此可能出现计数器已经为零（说明当前 Add 的操作为负，即 Done），但等待器为正的情况，依次调用存储原语释放产生的阻塞（本质上为加 1 操作）

整个流程：
```go
wg := sync.WaitGroup{}
wg.Add(1)
go func() { wg.Done() }()
wg.Wait()
```
在 wg 创建之初，计数器、等待器、存储原语的值均初始化为零值。不妨假设调用 wg.Add(1)，则计数器加 1 等待器、存储原语保持不变，均为 0。

wg.Done() 和 wg.Wait() 的调用顺序可能成两种情况：

情况 1：先调用 wg.Done() 再调用 wg.Wait()。

这时候 wg.Done() 使计数器减 1 ，这时计数器、等待器、存储原语均为 0，由于等待器为 0 则 runtime_Semrelease 不会被调用。 于是当 wg.Wait() 开始调用时，读取到计数器已经为 0，循环退出，wg.Wait() 调用完毕。

情况 2：先调用 wg.Wait() 再调用 wg.Done()。

这时候 wg.Wait() 开始调用时，读取到计数器为 1，则为等待器加 1，并调用 runtime_Semacquire 开始阻塞在存储原语为 0 的状态。

在阻塞的过程中，Goroutine 被调度器调度，开始执行 wg.Done()，于是计数器清零，但由于等待器为 1 大于零。 这时将等待器也清零，并调用与等待器技术相同次数（此处为 1 次）的 runtime_Semrelease，这导致存储原语的值变为 1，计数器和等待器均为零。 这时，runtime_Semacquire 在存储原语大于零后被唤醒，这时检查计数器和等待器是否为零（如果不为零则说明 Add 与 Wait 产生并发调用，直接 panic），这时他们为 0，因此进入下一个循环，当再次读取计数器时，发现计数器已经清理，于是退出 wg.Wait() 调用，结束阻塞。

## 并发模型

**MPG模型**

- G: Goroutine，即我们在 Go 程序中使用 go 关键字创建的执行体；
- M: Machine，或 worker thread，即传统意义上进程的线程；
- P: Processor，即一种人为抽象的、用于执行 Go 代码被要求局部资源。只有当 M 与一个 P 关联后才能执行 Go 代码。除非 M 发生阻塞或在进行系统调用时间过长时，没有与之关联的 P。

运行时调度器的任务是给不同的工作线程 (worker thread) 分发可供运行的（ready-to-run）Goroutine。

调度器的设计本质是：即既要保持足够的运行工作线程来利用有效硬件并发资源，又要暂止过多的工作线程来节约 CPU 能耗。

 Go 的调度器实现中设计了工作线程的自旋（spinning）状态：
1) 如果一个工作线程的本地队列、全局运行队列或网络轮询器中均没有可调度的任务，则该线程成为自旋线程；
2) 满足该条件、被复始的线程也被称为自旋线程，对于这种线程，运行时不做任何事情。

自旋线程在进行暂止之前，会尝试从任务队列中寻找任务。当发现任务时，则会切换成非自旋状态， 开始执行 Goroutine。而找到不到任务时，则进行暂止。

当一个 Goroutine 准备就绪时，会首先检查自旋线程的数量，而不是去复始一个新的线程。

如果最后一个自旋线程发现工作并且停止自旋时，则复始一个新的自旋线程。 这个方法消除了不合理的线程复始峰值，且同时保证最终的最大 CPU 并行度利用率。

```
  如果存在空闲的 P，且存在暂止的 M，并就绪 G
          +------+
          v      |
执行 --> 自旋 --> 暂止
 ^        |
 +--------+
  如果发现工作
```
就绪一个 G 的通用流程为：

- 提交一个 G 到 per-P 的本地工作队列
- 执行 StoreLoad 风格的写屏障
- 检查 sched.nmspinning 数量
  
而从自旋到非自旋转换的一般流程为：
- 减少 nmspinning 的数量
- 执行 StoreLoad 风格的写屏障
- 在所有 per-P 本地任务队列检查新的工作

P 只是处理器的抽象，而非处理器本身，它存在的意义在于实现工作窃取（work stealing）算法。 简单来说，每个 P 持有一个 G 的本地队列。

在没有 P 的情况下，所有的 G 只能放在一个全局的队列中。 当 M 执行完 G 而没有 G 可执行时，必须将队列锁住从而取值。

当引入了 P 之后，P 持有 G 的本地队列，而持有 P 的 M 执行完 G 后在 P 本地队列中没有 发现其他 G 可以执行时，虽然仍然会先检查全局队列、网络，但这时增加了一个从其他 P 的 队列偷取（steal）一个 G 来执行的过程。优先级为本地 > 全局 > 网络 > 偷取。

调度器，所有 Goroutine 被调度的核心，存放了调度器持有的全局资源，访问这些资源需要持有锁：

- 管理了能够将 G 和 M 进行绑定的 M 队列
- 管理了空闲的 P 链表（队列）
- 管理了 G 的全局队列
- 管理了可被复用的 G 的全局缓存
- 管理了 defer 池

## 内存分配
Go 的内存分配器主要包含以下几个核心组件：

- heapArena: 保留整个虚拟地址空间
- mheap：分配的堆，在页大小为 8KB 的粒度上进行管理
- mspan：是 mheap 上管理的一连串的页
- mcentral：收集了给定大小等级的所有 span
- mcache：为 per-P 的缓存。

其中页是向操作系统申请内存的最小单位，目前设计为 8KB。

**heapArena**

Go 堆被视为由多个 arena 组成，每个 arena 在 64 位机器上为 64MB，且起始地址与 arena 的大小对齐， 所有的 arena 覆盖了整个 Go 堆的地址空间。

**mspan**

然而管理 arena 如此粒度的内存并不符合实践，相反，所有的堆对象都通过 span 按照预先设定好的 大小等级分别分配，小于 32KB 的小对象则分配在固定大小等级的 span 上，否则直接从 mheap 上进行分配。

mspan 是相同大小等级的 span 的双向链表的一个节点，每个节点还记录了自己的起始地址、 指向的 span 中页的数量。

**mcache**
是一个 per-P 的缓存，它是一个包含不同大小等级的 span 链表的数组，其中 mcache.alloc 的每一个数组元素 都是某一个特定大小的 mspan 的链表头指针。

当 mcache 中 span 的数量不够使用时，会向 mcentral 的 nonempty 列表中获得新的 span。

**mcentral**

当 mcentral 中 nonempty 列表中也没有可分配的 span 时，则会向 mheap 提出请求，从而获得 新的 span，并进而交给 mcache。

Go 程序的执行是基于 goroutine 的，goroutine 和传统意义上的程序一样，也有栈和堆的概念。只不过 Go 的运行时帮我们屏蔽掉了这两个概念，只在运行时内部区分并分别对应：goroutine 执行栈以及 Go 堆。

**小对象分配**

当对一个小对象（<32KB）分配内存时，会将该对象所需的内存大小调整到某个能够容纳该对象的大小等级（size class）， 并查看 mcache 中对应等级的 mspan，通过扫描 mspan 的 freeindex 来确定是否能够进行分配。

当没有可分配的 mspan 时，会从 mcentral 中获取一个所需大小空间的新的 mspan，从 mcentral 中分配会对其进行加锁， 但一次性获取整个 span 的过程均摊了对 mcentral 加锁的成本。

如果 mcentral 的 mspan 也为空时，则它也会发生增长，从而从 mheap 中获取一连串的页，作为一个新的 mspan 进行提供。 而如果 mheap 仍然为空，或者没有足够大的对象来进行分配时，则会从操作系统中分配一组新的页（至少 1MB）， 从而均摊与操作系统沟通的成本。

**微对象分配**

对于过小的微对象（<16B），它们的分配过程与小对象的分配过程基本类似，但是是直接存储在 mcache 上，并由其以 16B 的块大小直接进行管理和释放。

**大对象分配**

大对象分配非常粗暴，不与 mcache 和 mcentral 沟通，直接绕过并通过 mheap 进行分配。

![img](img/mem-struct.png)

heap 最中间的灰色区域 arena 覆盖了 Go 程序的整个虚拟内存， 每个 arena 包括一段 bitmap 和一段指向连续 span 的指针； 每个 span 由一串连续的页组成；每个 arena 的起始位置通过 arenaHint 进行记录。

分配的顺序从右向左，代价也就越来越大。 小对象和微对象优先从白色区域 per-P 的 mcache 分配 span，这个过程不需要加锁（白色）； 若失败则会从 mheap 持有的 mcentral 加锁获得新的 span，这个过程需要加锁，但只是局部（灰色）； 若仍失败则会从右侧的 free 或 scav 进行分配，这个过程需要对整个 heap 进行加锁，代价最大（黑色）。

