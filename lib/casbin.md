## Casbin 是什么？
Casbin 可以：

1) 支持自定义请求的格式，默认的请求格式为{subject, object, action}。
2) 具有访问控制模型model和策略policy两个核心概念。
3) 支持RBAC中的多层角色继承，不止主体可以有角色，资源也可以具有角色。
4) 支持内置的超级用户 例如：root或administrator。超级用户可以执行任何操作而无需显式的权限声明。
5) 支持多种内置的操作符，如 keyMatch，方便对路径式的资源进行管理，如 /foo/bar 可以映射到 /foo*


Casbin 不能：

1) 身份认证 authentication（即验证用户的用户名、密码），casbin只负责访问控制。应该有其他专门的组件负责身份认证，然后由casbin进行访问控制，二者是相互配合的关系。
2) 管理用户列表或角色列表。 Casbin 认为由项目自身来管理用户、角色列表更为合适， 用户通常有他们的密码，但是 Casbin 的设计思想并不是把它作为一个存储密码的容器。 而是存储RBAC方案中用户和角色之间的映射关系。

## 工作原理

在 Casbin 中, 访问控制模型被抽象为基于 PERM (Policy, Effect, Request, Matcher) 的一个文件。 因此，切换或升级项目的授权机制与修改配置一样简单。 您可以通过组合可用的模型来定制您自己的访问控制模型。 例如，您可以在一个model中获得RBAC角色和ABAC属性，并共享一组policy规则。

```
# Request definition
[request_definition]
r = sub, obj, act

# Policy definition
[policy_definition]
p = sub, obj, act

# Policy effect
[policy_effect]
e = some(where (p.eft == allow))

# Matchers
[matchers]
m = r.sub == p.sub && r.obj == p.obj && r.act == p.act
```
ACL model的示例policy如下:
```
p, alice, data1, read
p, bob, data2, write
```
这表示：

- alice可以读取data1
- bob可以编写data2

[示例](https://casbin.org/docs/zh-CN/supported-models)

代码示例：
```go
package main

import (
    “fmt”

    “github.com/casbin/casbin”
)

func main() {
    //通过策略文件和模型配置穿件一个casbin访问控制实例
    e := casbin.NewEnforcer(“./perm.conf”, “./policy.csv”)

    //定义各种sub，obj和act的数组
    subs := []string{“bob”, “zeta”}
    objs := []string{“data1”, “data2”}
    acts := []string{“read”, “write”}

    //遍历组合sub，obj，act并打印出对应策略匹配结果。
    for _, sub := range subs {
        for _, obj := range objs {
            for _, act := range acts {
                fmt.Println(sub, “/“, obj, “/“, act, “=“, e.Enforce(sub, obj, act))
            }
        }
    }

}
```
[管理API](https://casbin.org/docs/zh-CN/management-api)