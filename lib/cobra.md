Cobra是一个库，其提供简单的接口来创建强大现代的CLI接口，类似于git或者go工具。同时，它也是一个应用，用来生成应用框架，从而开发以Cobra为基础的应用。Docker和Kubernetes源码中使用了Cobra。

Cobra有三个基本概念commands,arguments和flags。其中commands代表行为，arguments代表数值，flags代表对行为的改变。

基本模型
```
APPNAME COMMAND ARG --FLAG
```
