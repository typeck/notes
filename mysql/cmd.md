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

## 命令行sql
```sh
mysql -h test.trdplace.ads.sg1.mysql -u root -pBfnByprkhUeu57lr -D sands -e 'show tables' > sands_tables.txt
```
## 创建db
```
CREATE DATABASE 数据库名;
```