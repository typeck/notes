# CAP

分布式系统的最大难点，就是各个节点的状态如何同步。CAP 定理是这方面的基本定理，也是理解分布式系统的起点。

三个指标：
- Consistency
- Availability
- Partition tolerance

## Partition tolerance
大多数分布式系统都分布在多个子网络。每个子网络就叫做一个区（partition）。分区容错的意思是，区间通信可能失败。比如，一台服务器放在中国，另一台服务器放在美国，这就是两个区，它们之间可能无法通信。

一般来说，分区容错无法避免，因此可以认为 CAP 的 P 总是成立。CAP 定理告诉我们，剩下的 C 和 A 无法同时做到。

## Consistency
Consistency 中文叫做"一致性"。意思是，写操作之后的读操作，必须返回该值。

## Abiailability
Availability 中文叫做"可用性"，意思是只要收到用户的请求，服务器就必须给出回应。

如果保证 G2 的一致性，那么 G1 必须在写操作时，锁定 G2 的读操作和写操作。只有数据同步后，才能重新开放读写。锁定期间，G2 不能读写，没有可用性不。

如果保证 G2 的可用性，那么势必不能锁定 G2，所以一致性不成立。

综上所述，G2 无法同时做到一致性和可用性。系统设计时只能选择一个目标。如果追求一致性，那么无法保证所有节点的可用性；如果追求所有节点的可用性，那就没法做到一致性。

# 分布式锁

在很多场景中，我们为了保证数据的最终一致性，需要很多的技术方案来支持，比如分布式事务、分布式锁等。有的时候，我们需要保证一个方法在同一时间内只能被同一个线程执行。

## 数据库实现

```sql
DROP TABLE IF EXISTS `method_lock`;
CREATE TABLE `method_lock` (
  `id` int(11) unsigned NOT NULL AUTO_INCREMENT COMMENT '主键',
  `method_name` varchar(64) NOT NULL COMMENT '锁定的方法名',
  `desc` varchar(255) NOT NULL COMMENT '备注信息',
  `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uidx_method_name` (`method_name`) USING BTREE
) ENGINE=InnoDB AUTO_INCREMENT=3 DEFAULT CHARSET=utf8 COMMENT='锁定中的方法';
```
想要执行某个方法，就使用这个方法名向表中插入数据：
```sql
INSERT INTO method_lock (method_name, desc) VALUES ('methodName', '测试的methodName');
```

成功插入则获取锁，执行完成后删除对应的行数据释放锁：
```sql
delete from method_lock where method_name ='methodName';
```
缺点：
- 可用性 （双机部署、数据同步、主备切换；）
- 不具备可重入的特性
- 没有锁失效机制
- 不具备阻塞锁特性

# redis实现
[redis设计与实现](../redis/redis设计与实现.md)

（1）获取锁的时候，使用setnx加锁，并使用expire命令为锁添加一个超时时间，超过该时间则自动释放锁，锁的value值为一个随机生成的UUID，通过此在释放锁的时候进行判断。

（2）获取锁的时候还设置一个获取的超时时间，若超过这个时间则放弃获取锁。

（3）释放锁的时候，通过UUID判断是不是该锁，若是该锁，则执行delete进行锁释放。

## etcd实现

```go
func Test_TryLock(t *testing.T) {
	cli, err := clientv3.New(clientv3.Config{Endpoints: []string{"http://172.27.122.2:23790"}})
	if err != nil {
		fmt.Println(err)
	}
	defer cli.Close()

	// create two separate sessions for lock competition
	s1, err := concurrency.NewSession(cli)
	if err != nil {
		fmt.Println(err)
	}
	defer s1.Close()
	m1 := concurrency.NewMutex(s1, "/my-lock")

	s2, err := concurrency.NewSession(cli)
	if err != nil {
		fmt.Println(err)
	}
	defer s2.Close()
	m2 := concurrency.NewMutex(s2, "/my-lock")

	// acquire lock for s1
	if err = m1.Lock(context.TODO()); err != nil {
		fmt.Println(err)
	}
	fmt.Println("acquired lock for s1")

	if err = m2.TryLock(context.TODO()); err == nil {
		fmt.Println("should not acquire lock")
	}
	if err == concurrency.ErrLocked {
		fmt.Println("cannot acquire lock for s2, as already locked in another session")
	}

	if err = m1.Unlock(context.TODO()); err != nil {
		fmt.Println(err)
	}
	fmt.Println("released lock for s1")
	if err = m2.TryLock(context.TODO()); err != nil {
		fmt.Println(err)
	}
	fmt.Println("acquired lock for s2")
}

```