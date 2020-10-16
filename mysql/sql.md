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
