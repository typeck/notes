hadoop 实现了MapReduce计算模型，可以将计算任务分割成多个处理单元，然后分散到一群硬件机器上，从而降低成本并提供水平扩展性。

Hive提供了HQL，来查询存储在hadoop集群中的数据。

Hive不是一个完整的数据库。Hive最适用于数据仓库应用，使用改应用程序进行相关的静态数据分析，不需要快速给出结果，而且数据本身不会频繁变化。