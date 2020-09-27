## 键值设计

### key
以业务名(或数据库名)为前缀(防止key冲突)，用冒号分隔，比如业务名:表名:id

`ugc:video:1`

保证语义的前提下，控制key的长度，当key较多时，内存占用也不容忽视，例如：

`user:{uid}:friends:messages:{mid}`

简化为

`u:{uid}:fr:m:{mid}`

不要包含特殊字符. 反例：包含空格、换行、单双引号以及其他转义字符
### value

防止网卡流量、慢查询，string类型控制在10KB以内，hash、list、set、zset元素个数不要超过5000。反例：一个包含200万个元素的list。非字符串的bigkey，不要使用del删除，使用hscan、sscan、zscan方式渐进式删除，同时要注意防止bigkey过期时间自动删除问题(例如一个200万的zset设置1小时过期，会触发del操作，造成阻塞，而且该操作不会不出现在慢查询中(latency可查))，查找方法和删除方法.

控制key的生命周期. 
redis不是垃圾桶，建议使用expire设置过期时间(条件允许可以打散过期时间，防止集中过期)，不过期的数据重点关注idletime。

## 命令使用

1、O(N)命令关注N的数量

例如hgetall、lrange、smembers、zrange、sinter等并非不能使用，但是需要明确N的值。有遍历的需求可以使用hscan、sscan、zscan代替。

2、使用批量操作提高效率

原生命令：例如mget、mset。

非原生命令：可以使用pipeline提高效率。

但要注意控制一次批量操作的元素个数(例如500以内，实际也和元素字节数有关)。

注意两者不同：

原生是原子操作，pipeline是非原子操作。

pipeline可以打包不同的命令，原生做不到

pipeline需要客户端和服务端同时支持。

## 客户端使用

1、避免多个应用使用一个Redis实例

不相干的业务拆分，公共数据做服务化。

2、使用连接池

可以有效控制连接，同时提高效率。

3、熔断功能

高并发下建议客户端添加熔断功能(例如netflix hystrix)

4、淘汰策略

根据自身业务类型，选好maxmemory-policy(最大内存淘汰策略)，设置好过期时间。默认策略是volatile-lru，即超过最大内存后，在过期键中使用lru算法进行key的剔除，保证不过期数据不被删除，但是可能会出现OOM问题。

其他策略如下：

allkeys-lru：根据LRU算法删除键，不管数据有没有设置超时属性，直到腾出足够空间为止。

allkeys-random：随机删除所有键，直到腾出足够空间为止。

volatile-random:随机删除过期键，直到腾出足够空间为止。

volatile-ttl：根据键值对象的ttl属性，删除最近将要过期数据。如果没有，回退到noeviction策略。

noeviction：不会剔除任何数据，拒绝所有写入操作并返回客户端错误信息"(error) OOM command not allowed when used memory"，此时Redis只响应读操作。

参考

[阿里官方Redis开发规范](https://zhuanlan.zhihu.com/p/149370312)