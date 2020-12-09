## 当前运行的所有事务
```sql
select * from information_schema.innodb_trx\G
```
## 当前出现的锁
```sql
select * from information_schema.innodb_locks\G
```
## 锁等待的对应关系 
```sql
select * from information_schema.innodb_lock_waits\G
```
