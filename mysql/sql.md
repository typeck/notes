## 索引
**索引类型**：
* UNIQUE(唯一索引)：不可以出现相同的值，可以有NULL值
* INDEX(普通索引)：允许出现相同的索引内容
* PROMARY KEY(主键索引)：不允许出现相同的值
* fulltext index(全文索引)：可以针对值中的某个单词，但效率确实不敢恭维
* 组合索引：实质上是将多个字段建到一个索引里，列值的组合必须唯一


**创建索引**
```sql
ALTER TABLE 表名 ADD 索引类型 （unique,primary key,fulltext,index）[索引名]（字段名）
```
CREATE INDEX可用于对表增加普通索引或UNIQUE索引，可用于建表时创建索引。**索引名不可选。**
```sql
CREATE INDEX index_name ON table_name(username(length)); 
```

**删除索引**

删除索引可以使用ALTER TABLE或DROP INDEX语句来实现。DROP INDEX可以在ALTER TABLE内部作为一条语句处理
```sql
drop index index_name on table_name ;

alter table table_name drop index index_name ;

alter table table_name drop primary key ;
```
**查看索引**
```sql
SHOW INDEX FROM <表名> [ FROM <数据库名>]
```

## 分区

**创建分区**

```sql
ALTER TABLE third_party_offer_info PARTITION BY RANGE ( dsp_id ) (
	PARTITION d1
	VALUES
		less than ( 10001 ),
		PARTITION d3
	VALUES
		less than ( 10003 ),
		PARTITION d4
	VALUES
		less than ( 10004 ),
		PARTITION d5
	VALUES
		less than ( 10005 ),
		PARTITION d6
	VALUES
		less than ( 10006 ),
		PARTITION d7
	VALUES
		less than ( 10007 ),
		PARTITION d8
	VALUES
		less than ( 10008 ),
		PARTITION d9
	VALUES
	less than ( 10009 ) 
	);

```
限制：

1) 分区键必须包含在表的所有主键、唯一键中。
   
   解决：修改主键，使分区键成为主键的一部分 `alter table tmp_qw_test drop primary key, add primary key(id,gmt_create)`
2) MYSQL只能在使用分区函数的列本身进行比较时才能过滤分区，而不能根据表达式的值去过滤分区，即使这个表达式就是分区函数也不行。

3) 最大分区数： 不使用NDB存储引擎的给定表的最大可能分区数为8192（包括子分区）。如果当分区数很大，但是未达到8192时提示 Got error … from storage engine: Out of resources when opening file,可以通过增加open_files_limit系统变量的值来解决问题，当然同时打开文件的数量也可能由操作系统限制。

4) 不支持查询缓存： 分区表不支持查询缓存，对于涉及分区表的查询，它自动禁用。 查询缓存无法启用此类查询。

5) 分区的innodb表不支持外键。

6) 服务器SQL_mode影响分区表的同步复制。 主机和从机上的不同SQL_mode可能会导致sql语句; 这可能导致分区之间的数据分配给定主从位置不同，甚至可能导致插入主机上成功的分区表在从库上失败。 为了获得最佳效果，您应该始终在主机和从机上使用相同的服务器SQL模式。

7) ALTER TABLE … ORDER BY： 对分区表运行的ALTER TABLE … ORDER BY列语句只会导致每个分区中的行排序。

8) 全文索引。 分区表不支持全文索引，即使是使用InnoDB或MyISAM存储引擎的分区表。
9) 分区表无法使用外键约束。
10) Spatial columns： 具有空间数据类型（如POINT或GEOMETRY）的列不能在分区表中使用。
11) 临时表： 临时表不能分区。
12) subpartition问题： subpartition必须使用HASH或KEY分区。 只有RANGE和LIST分区可能被分区; HASH和KEY分区不能被子分区。
13) 分区表不支持mysqlcheck，myisamchk和myisampack。 

## 基本命令

**修改字段**
```sql
ALTER TABLE user10 MODIFY card CHAR(10) AFTER test;
```