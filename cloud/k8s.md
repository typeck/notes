# 基本概念
**Container**（容器）是一种便携式、轻量级的操作系统级虚拟化技术。它使用 namespace 隔离不同的软件运行环境，并通过镜像自包含软件的运行环境，从而使得容器可以很方便的在任何地方运行。

**pod**，Kubernetes 使用 Pod 来管理容器，每个 Pod 可以包含一个或多个紧密关联的容器。Pod 是一组紧密关联的容器集合，它们共享 PID、IPC、Network 和 UTS namespace，是 Kubernetes 调度的基本单位。Pod 内的多个容器共享网络和文件系统，可以通过进程间通信和文件共享这种简单高效的方式组合完成服务。

**Node** 是 Pod 真正运行的主机，可以是物理机，也可以是虚拟机。为了管理 Pod，每个 Node 节点上至少要运行 container runtime（比如 docker 或者 rkt）、kubelet 和 kube-proxy 服务。

**Namespace** 是对一组资源和对象的抽象集合，比如可以用来将系统内部的对象划分为不同的项目组或用户组。常见的 pods, services, replication controllers 和 deployments 等都是属于某一个 namespace 的（默认是 default），而 node, persistentVolumes 等则不属于任何 namespace。

**Service** 是应用服务的抽象，通过 labels 为应用提供负载均衡和服务发现。匹配 labels 的 Pod IP 和端口列表组成 endpoints，由 kube-proxy 负责将服务 IP 负载均衡到这些 endpoints 上。

每个 Service 都会自动分配一个 cluster IP（仅在集群内部可访问的虚拟地址）和 DNS 名，其他容器可以通过该地址或 DNS 来访问服务，而不需要了解后端容器的运行。

**Label** 是识别 Kubernetes 对象的标签，以 key/value 的方式附加到对象上（key 最长不能超过 63 字节，value 可以为空，也可以是不超过 253 字节的字符串）。

Label 不提供唯一性，并且实际上经常是很多对象（如 Pods）都使用相同的 label 来标志具体的应用。

## k8s 安装后启动报错`The connection to the server localhost:8080 was refused - did you specify th`
解决方法：
```shell
 kubeadm init --kubernetes-version=v1.19.3 --pod-network-cidr 10.244.0.0/16  --ignore-preflight-errors=NumCPU

# To start using your cluster, you need to run the following as a regular user:

  mkdir -p $HOME/.kube
  sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config
  sudo chown $(id -u):$(id -g) $HOME/.kube/config
```
创建容器
```
kubectl run --image=nginx:alpine nginx-app --port=80
```
- kubectl get - 类似于 docker ps，查询资源列表
- kubectl describe - 类似于 docker inspect，获取资源的详细信息
- kubectl logs - 类似于 docker logs，获取容器的日志
- kubectl exec - 类似于 docker exec，在容器内执行一个命令

部署pod, 默认部署在 default namespace, `kubectl apply -f nginx-deployment.yaml`
查看部署,`kubectl  get  pod`
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
  labels:
    app: nginx
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:latest
        ports:
        - containerPort: 80
```
部署 service, kubectl create -f nginx-service.yaml
```yaml
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
  labels:
    app: nginx
spec:
  ports:
  - port: 88
    targetPort: 80
  selector:
    app: nginx
  type: NodePort
```
查看service `kubectl get svc`

**集群** 一个 Kubernetes 集群由分布式存储 etcd、控制节点 controller 以及服务节点 Node 组成。

- 控制节点主要负责整个集群的管理，比如容器的调度、维护资源的状态、自动扩展以及滚动更新等
- 服务节点是真正运行容器的主机，负责管理镜像和容器以及 cluster 内的服务发现和负载均衡
- etcd 集群保存了整个集群的状态

**核心组件**
- etcd 保存了整个集群的状态；
- kube-apiserver 提供了资源操作的唯一入口，并提供认证、授权、访问控制、API 注册和发现等机制；
- kube-controller-manager 负责维护集群的状态，比如故障检测、自动扩展、滚动更新等；
- kube-scheduler 负责资源的调度，按照预定的调度策略将 Pod 调度到相应的机器上；
- kubelet 负责维持容器的生命周期，同时也负责 Volume（CVI）和网络（CNI）的管理；
- Container runtime 负责镜像管理以及 Pod 和容器的真正运行（CRI），默认的容器运行时为 Docker；
- kube-proxy 负责为 Service 提供 cluster 内部的服务发现和负载均衡；

![img](img/components.png)

## helm 安装consul
Helm 是 Kubernetes 的包管理器。包管理器类似于我们在 Ubuntu 中使用的apt、Centos中使用的yum 或者Python中的 pip 一样，能快速查找、下载和安装软件包。Helm 由客户端组件 helm 和服务端组件 Tiller 组成, 能够将一组K8S资源打包统一管理, 是查找、共享和使用为Kubernetes构建的软件的最佳方式。

config.yaml 
```yaml
# Choose an optional name for the datacenter
global:
  datacenter: minidc

# Enable the Consul Web UI via a NodePort
ui:
  service:
    type: NodePort

# Enable Connect for secure communication between nodes
connectInject:
  enabled: true
# Enable CRD controller
controller:
  enabled: true

client:
  enabled: true

# Use only one Consul server for local development
server:
  replicas: 1
  bootstrapExpect: 1
  disruptionBudget:
    enabled: true
    maxUnavailable: 0
```
执行：
```sh
helm install -f kind_config.yaml consul hashicorp/consul
```
出现的问题：
```
root@ip-172-27-122-2 ~/k/consul# kubectl get pod
NAME                                                              READY   STATUS              RESTARTS   AGE
consul-consul-connect-injector-webhook-deployment-779c55bfnw876   0/1     ContainerCreating   0          5s
consul-consul-controller-7b58885d98-htjnw                         0/1     ContainerCreating   0          5s
consul-consul-ltqpl                                               0/1     Pending             0          5s
consul-consul-server-0                                            0/1     Pending             0          5s
consul-consul-webhook-cert-manager-7d59b9f4f5-dglpm               0/1     ContainerCreating   0          5s
```
pod 处于pending状态，

```sh
root@ip-172-27-122-2 ~/k/consul# kubectl describe pod consul-consul-server-0
Name:           consul-consul-server-0
Namespace:      default
Priority:       0
Node:           <none>
Labels:         app=consul
                chart=consul-helm
                component=server
                controller-revision-hash=consul-consul-server-6974c44994
                hasDNS=true
                release=consul
                statefulset.kubernetes.io/pod-name=consul-consul-server-0
Annotations:    consul.hashicorp.com/config-checksum: 8dacaf419ab1cdebc2823c153b9cd411e489ed7cb75a9738c84173f04b8cd85a
                consul.hashicorp.com/connect-inject: false
Status:         Pending
IP:
IPs:            <none>
Controlled By:  StatefulSet/consul-consul-server
Containers:
  consul:
    Image:       hashicorp/consul:1.9.1
    Ports:       8500/TCP, 8301/TCP, 8301/UDP, 8302/TCP, 8300/TCP, 8600/TCP, 8600/UDP
    Host Ports:  0/TCP, 0/TCP, 0/UDP, 0/TCP, 0/TCP, 0/TCP, 0/UDP
    Command:
      /bin/sh
      -ec
      CONSUL_FULLNAME="consul-consul"

      exec /bin/consul agent \
        -advertise="${ADVERTISE_IP}" \
        -bind=0.0.0.0 \
        -bootstrap-expect=1 \
        -client=0.0.0.0 \
        -config-dir=/consul/config \
        -datacenter=minidc \
        -data-dir=/consul/data \
        -domain=consul \
        -hcl="connect { enabled = true }" \
        -ui \
        -retry-join="${CONSUL_FULLNAME}-server-0.${CONSUL_FULLNAME}-server.${NAMESPACE}.svc:8301" \
        -serf-lan-port=8301 \
        -server

    Limits:
      cpu:     100m
      memory:  100Mi
    Requests:
      cpu:      100m
      memory:   100Mi
    Readiness:  exec [/bin/sh -ec curl http://127.0.0.1:8500/v1/status/leader \
2>/dev/null | grep -E '".+"'
] delay=5s timeout=5s period=3s #success=1 #failure=2
    Environment:
      ADVERTISE_IP:   (v1:status.podIP)
      POD_IP:         (v1:status.podIP)
      NAMESPACE:     default (v1:metadata.namespace)
    Mounts:
      /consul/config from config (rw)
      /consul/data from data-default (rw)
      /var/run/secrets/kubernetes.io/serviceaccount from consul-consul-server-token-9chvd (ro)
Conditions:
  Type           Status
  PodScheduled   False
Volumes:
  data-default:
    Type:       PersistentVolumeClaim (a reference to a PersistentVolumeClaim in the same namespace)
    ClaimName:  data-default-consul-consul-server-0
    ReadOnly:   false
  config:
    Type:      ConfigMap (a volume populated by a ConfigMap)
    Name:      consul-consul-server-config
    Optional:  false
  consul-consul-server-token-9chvd:
    Type:        Secret (a volume populated by a Secret)
    SecretName:  consul-consul-server-token-9chvd
    Optional:    false
QoS Class:       Guaranteed
Node-Selectors:  <none>
Tolerations:     node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                 node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type     Reason            Age                 From               Message
  ----     ------            ----                ----               -------
  Warning  FailedScheduling  48s (x13 over 13m)  default-scheduler  0/1 nodes are available: 1 node(s) didn't find available persistent volumes to bind.
  Warning  FailedScheduling  37s (x2 over 37s)   default-scheduler  0/1 nodes are available: 1 node(s) had volume node affinity conflict.
  ```
  原因是自动创建的pvc无法通过storgeclass创建合适的pv。需要手动创建pvc、pv。

  参考 [Helm Chart Configuration](https://www.consul.io/docs/k8s/helm.html#v-server-storageclass)

**pv**
PV全称叫做Persistent Volume，持久化存储卷。它是用来描述或者说用来定义一个存储卷的，这个通常都是有运维或者数据存储工程师来定义。

注意：PV必须先与POD创建，而且只能是网络存储不能属于任何Node，虽然它支持HostPath类型但由于你不知道POD会被调度到哪个Node上，所以你要定义HostPath类型的PV就要保证所有节点都要有HostPath中指定的路径。

**如下values必须和node name相同才能正确调度。**

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: consul-server-0
  labels:
    type: local
    app: consul
spec:
  claimRef:
    namespace: default
    name: data-default-consul-consul-server-0
  capacity:
    storage: 2Gi
  accessModes:
    - ReadWriteOnce
  persistentVolumeReclaimPolicy: Retain
  storageClassName: local-ssd # <<<<<<<<< This must match the storage class.
  local:
    path: /var/data # <<<<<<<<<<<<<< wherever you want the data
  nodeAffinity:
    required:
      nodeSelectorTerms:
        - matchExpressions:
            - key: kubernetes.io/hostname
              operator: In
              values:
                - ip-172-27-122-2.ap-southeast-1.compute.internal # <<<<< Your hostnames
```
**PVC**

是用来描述希望使用什么样的或者说是满足什么条件的存储，它的全称是Persistent Volume Claim，也就是持久化存储声明。开发人员使用这个来描述该容器需要一个什么存储。

PVC就相当于是容器和PV之间的一个接口，使用人员只需要和PVC打交道即可。

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: data-default-consul-consul-server-0
  labels:
    app: consul
spec:
  accessModes:
    - ReadWriteOnce
      #volumeMode: Filesystem
  resources:
    requests:
      storage: 2Gi
  storageClassName: local-ssd
  selector:
    matchLabels:
      app: consul
```
**storgeclass** 

PV是运维人员来创建的，开发操作PVC，可是大规模集群中可能会有很多PV，如果这些PV都需要运维手动来处理这也是一件很繁琐的事情，所以就有了动态供给概念，也就是Dynamic Provisioning。而我们上面的创建的PV都是静态供给方式，也就是Static Provisioning。而动态供给的关键就是StorageClass，它的作用就是创建PV模板。

创建StorageClass里面需要定义PV属性比如存储类型、大小等；另外创建这种PV需要用到存储插件。最终效果是，用户提交PVC，里面指定存储类型，如果符合我们定义的StorageClass，则会为其自动创建PV并进行绑定。
```yaml
kind: StorageClass
apiVersion: storage.k8s.io/v1
metadata:
  name: local-ssd
  provisioner: kubernetes.io/host-path
volumeBindingMode: WaitForFirstConsumer
```

**本地持久化存储**

本地持久化存储（Local Persistent Volume）就是把数据存储在POD运行的宿主机上，我们知道宿主机有hostPath和emptyDir，由于这两种的特定不适用于本地持久化存储。那么本地持久化存储必须能保证POD被调度到具有本地持久化存储的节点上。

为什么需要这种类型的存储呢？有时候你的应用对磁盘IO有很高的要求，网络存储性能肯定不如本地的高，尤其是本地使用了SSD这种磁盘。

但这里有个问题，通常我们先创建PV，然后创建PVC，这时候如果两者匹配那么系统会自动进行绑定，哪怕是动态PV创建，也是先调度POD到任意一个节点，然后根据PVC来进行创建PV然后进行绑定最后挂载到POD中，可是本地持久化存储有一个问题就是这种PV必须要先准备好，而且不一定集群所有节点都有这种PV，如果POD随意调度肯定不行，如何保证POD一定会被调度到有PV的节点上呢？这时候就需要在PV中声明节点亲和，且POD被调度的时候还要考虑卷的分布情况。

[从外部访问 Kubernetes 集群中应用的几种方式](https://www.cnblogs.com/hongdada/p/11328082.html)