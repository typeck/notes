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

