## jaeger组件介绍：

jaeger-client：jaeger 的客户端，实现了opentracing协议；

jaeger-agent：jaeger client的一个代理程序，client将收集到的调用链数据发给agent，然后由agent发给collector；

jaeger-collector：负责接收jaeger client或者jaeger agent上报上来的调用链数据，然后做一些校验，比如时间范围是否合法等，最终会经过内部的处理存储到后端存储；

jaeger-query：专门负责调用链查询的一个服务，有自己独立的UI；

jaeger-ingester：中文名称“摄食者”，可用从kafka读取数据然后写到jaeger的后端存储，比如Cassandra和Elasticsearch；

spark-job：基于spark的运算任务，可以计算服务的依赖关系，调用次数等；

## docker安装 

其中jaeger-collector和jaeger-query是必须的，其余的都是可选的，我们没有采用agent上报的方式，而是让客户端直接通过endpoint上报到collector。

### es 安装：

`sudo docker pull elasticsearch:7.4.2`

创建目录:

```
mkdir -p /mydata/elasticsearch/config/
mkdir -p /mydata/elasticsearch/data/
echo "http.host: 0.0.0.0">>/mydata/elasticsearch/config/elasticsearch.yml
```

拉起实例并启动

```sh
sudo docker run --name elasticsearch -p 9200:9200 -p 9300:9300 \
-e ES_JAVA_OPS="-Xms256m -Xmx256m" \
-v /mydata/elasticsearch/config/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml \
-v /mydata/elasticsearch/data:/usr/share/elasticsearch/data \
-v /mydata/elasticsearch/plugins:/usr/share/elasticsearch/plugins \
-d elasticsearch:7.4.2
```

注意:
chmod -R 777 /mydata/elasticsearch
要有访问权限

参数说明:

-p 9200:9200 将容器的9200端口映射到主机的9200端口;

--name elasticsearch 给当前启动的容器取名叫 elasticsearch

-v /mydata/elasticsearch/data:/usr/share/elasticsearch/data 将数据文件夹挂载到主机;

-v /mydata/elasticsearch/config/elasticsearch.yml:/usr/share/elasticsearch/config/elasticsearch.yml 将配置文件挂载到主机;

-d 以后台方式运行(daemon)

-e ES_JAVA_OPS="-Xms256m -Xmx256m" 测试时限定内存小一点

启动elasticsearch容器

`docker start elasticsearch`

验证：

`curl 'http://localhost:9200'`

### jaeger 安装
```sh
docker run -d --rm -p 14268:14268 -p 14269:14269 -e SPAN_STORAGE_TYPE=elasticsearch -e ES_SERVER_URLS=http://172.27.122.2:9200 jaegertracing/jaeger-collector:1.14
```

```sh
docker run -d --rm -p 16686:16686 -p 16687:16687 -e SPAN_STORAGE_TYPE=elasticsearch -e ES_SERVER_URLS=http://172.27.122.2:9200 jaegertracing/jaeger-query:1.14
```

验证：

`http://172.27.122.2:16686/`