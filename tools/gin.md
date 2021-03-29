# 概述 (基于gin1.6.3)
- engine：
  web server的基础支持，服务的入口，根级数据结构
- RouterGroup：
  用于gin，rest路由绑定和匹配的基础，源于radix-tree数据结构的支持
- HandlerFunc：
  逻辑处理器和中间件实现的函数签名
- Context：
  封装了请求和响应的操作，上下文

# demo
```go
type barForm struct {
    Foo string  `form:"foo" binding:"required"`
    Bar int     `form:"bar" binding:"required"`
}

func (fooHdl FooHdl) Bar(c *gin.Context) {
    var bform = new(barForm)
    if err := c.ShouldBind(bform); err != nil {
        // true: parse form error
        return
    }

    // handle biz logic and generate response structure
    // c (gin.Context) methods could called to support process-controling

    c.JSON(http.StatusOK, resp)
    // c.String() alse repsonse to client
}

// mountRouters .
func mountRouters(engi *gin.Engine) {
    // use middlewares
    engi.Use(gin.Logger())
    engi.Use(gin.Recovery())

    // mount routers
    group := engi.Group("/v1")
    {
        fooHdl := demohtp.New()
        group.GET("/foo", fooHdl.Bar)
        group.GET("/echo", fooHdl.Echo)
        // subGroup := group.Group("/subg")
        // subGroup.GET("/hdl1", fooHdl.SubGroupHdl1) // 最终路由："targetURI = /v1/subg/hdl1"
    }
}

func main() {
    engi := gin.New()

    mountRouters(engi)

    if err := engi.Run(":8080"); err != nil {
        log.Fatalf("engi exit with err=%v", err)
    }
}
```
# engine 
Engine是根入口，它把RouterGroup结构图体嵌入自身，以获得了Group，GET，POST等路由管理方法。
初始化时，routerGroup.basePath为"/"，树的根节点
```go
func New() *Engine {
	debugPrintWARNINGNew()
	engine := &Engine{
		RouterGroup: RouterGroup{
			Handlers: nil,
			basePath: "/",
			root:     true,
		},
		FuncMap:                template.FuncMap{},
		RedirectTrailingSlash:  true,
		RedirectFixedPath:      false,
		HandleMethodNotAllowed: false,
		ForwardedByClientIP:    true,
		AppEngine:              defaultAppEngine,
		UseRawPath:             false,
		RemoveExtraSlash:       false,
		UnescapePathValues:     true,
		MaxMultipartMemory:     defaultMultipartMemory,
		trees:                  make(methodTrees, 0, 9),
		delims:                 render.Delims{Left: "{{", Right: "}}"},
		secureJsonPrefix:       "while(1);",
	}
	engine.RouterGroup.engine = engine
	engine.pool.New = func() interface{} {
		return engine.allocateContext()
	}
	return engine
}
```
engine实现了ServeHTTP(ResponseWriter, *Request)函数签名，可以注册到原生http包。
使用sync.pool复用gin.Context

```go
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
}
```
```go

```
# RouterGroup & MethodTree
RouterGroup通过Group函数来衍生下一级别的RouterGroup，会拷贝父级RouterGroup的中间件，重新计算basePath。
```go
func (group *RouterGroup) Group(relativePath string, handlers ...HandlerFunc) *RouterGroup {
	return &RouterGroup{
		Handlers: group.combineHandlers(handlers),
		basePath: group.calculateAbsolutePath(relativePath),
		engine:   group.engine,
	}
}
```
routerGroup的POST、GET、PUT都会调用同一个函数：handel
调用链路：engine.GET -> routergroup.GET -> routergroup.handle -> engine.addRoute -> methodTree.addRoute -> node(radix-tree's node).insertChild
```go
func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
    //这里是核心，通过addRoute将路由注册到engine中
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}
```
```go
func (engine *Engine) addRoute(method, path string, handlers HandlersChain) {
    // engine.trees : []methodTree，每个方法都有一个单独的树
	root := engine.trees.get(method)
	if root == nil {
		root = new(node)
		root.fullPath = "/"
		engine.trees = append(engine.trees, methodTree{method: method, root: root})
	}
	root.addRoute(path, handlers)
}
```
# 路由匹配，radix-tree
对于基数树的每个节点，如果该节点是唯一的子树的话，就和父节点合并。
![](./img/radix_tree.png)

```
GET("/", func1)
GET("/search/", func2)
GET("/support/", func3)
GET("/blog/", func4)
GET("/blog/:post/", func5)
GET("/about-us/", func6)
GET("/about-us/team/", func7)
GET("/contact/", func8)
```
如上对应的树为：
```
Priority   Path             Handle
9          \                *<1>
3          ├s               nil
2          |├earch\         *<2>
1          |└upport\        *<3>
2          ├blog\           *<4>
1          |    └:post      nil
1          |         └\     *<5>
2          ├about-us\       *<6>
1          |        └team\  *<7>
1          └contact\        *<8>
```
:post是真实的post name的一个占位符（就是一个参数）。这里体现了radix tree相较于hash-map的一个优点，树结构允许我们的路径中存在动态的部分（参数）,因为我们匹配的是路由的模式而不是hash值

每一层的节点按照priority排序，priority是节点的子节点（儿子节点，孙子节点等）注册的handler的数量，这样做有两个好处：

1) 被最多路径包含的节点会被最先评估。这样可以让尽量多的路由快速被定位。
2) 有点像成本补偿。最长的路径可以被最先评估，补偿体现在最长的路径需要花费更长的时间来定位，如果最长路径的节点能被优先评估（即每次拿子节点都命中），那么所花时间不一定比短路径的路由长。

节点数据结构：
```go
type node struct {
    // 节点路径，比如上面的s，earch，和upport
    path      string
    // 节点是否是参数节点，比如上面的:post
    wildChild bool
    // 节点类型，包括static, root, param, catchAll
    // static: 静态节点，比如上面的s，earch等节点
    // root: 树的根节点
    // catchAll: 有*匹配的节点
    // param: 参数节点
    nType     nodeType
    // 路径上最大参数个数
    maxParams uint8
    // 和children字段对应, 保存的是分裂的分支的第一个字符
    // 例如search和support, 那么s节点的indices对应的"eu"
    // 代表有两个分支, 分支的首字母分别是e和u
    indices   string
    // 儿子节点
    children  []*node
    // 处理函数
    handlers  HandlersChain
    // 优先级，子节点注册的handler数量
    priority  uint32
}
```
# context
Context是上下文传递的核心，它包括了请求处理，响应处理，表单解析等重要工作。

```go
// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
    writermem responseWriter // 实现了http.ResponseWriter 和 gin.ResponseWriter
    Request   *http.Request  // http.Request, 暴露给handler
    // gin.ResponseWriter 包含了：
    // http.ResponseWriter，http.Hijacker，http.Flusher，http.CloseNotifier和额外方法
    // 暴露给handler，是writermm的复制
    Writer    ResponseWriter 

    Params   Params        // 路径参数 
    handlers HandlersChain // 调用链
    index    int8          // 当前handler的索引
    fullPath string

    engine *Engine // 对engine的引用
    Keys map[string]interface{} // c.GET / c.SET 的支持，常用于session传递。

    // Errors is a list of errors attached to all the handlers/middlewares who used this context.
    Errors errorMsgs

    // Accepted defines a list of manually accepted formats for content negotiation.
    Accepted []string

    // query 参数缓存
    queryCache url.Values
    // 表单参数缓存, 跟queryCache作用类似
    formCache url.Values
}
```
handelHTTPRequest 为所有流量的入口
```go
func (engine *Engine) handleHTTPRequest(c *Context) {
	httpMethod := c.Request.Method
	rPath := c.Request.URL.Path
	unescape := false
	if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
		rPath = c.Request.URL.RawPath
		unescape = engine.UnescapePathValues
	}

	if engine.RemoveExtraSlash {
		rPath = cleanPath(rPath)
	}

	// Find root of the tree for the given HTTP method
	t := engine.trees
	for i, tl := 0, len(t); i < tl; i++ {
		if t[i].method != httpMethod {
			continue
		}
		root := t[i].root
		// Find route in tree ， 路径查找
		value := root.getValue(rPath, c.Params, unescape)
		if value.handlers != nil {
			c.handlers = value.handlers
			c.Params = value.params
			c.fullPath = value.fullPath
            //开始顺序执行查找到路径的handlers
			c.Next()
			c.writermem.WriteHeaderNow()
			return
		}
		if httpMethod != "CONNECT" && rPath != "/" {
			if value.tsr && engine.RedirectTrailingSlash {
				redirectTrailingSlash(c)
				return
			}
			if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
				return
			}
		}
		break
	}

	if engine.HandleMethodNotAllowed {
		for _, tree := range engine.trees {
			if tree.method == httpMethod {
				continue
			}
			if value := tree.root.getValue(rPath, nil, unescape); value.handlers != nil {
				c.handlers = engine.allNoMethod
				serveError(c, http.StatusMethodNotAllowed, default405Body)
				return
			}
		}
	}
	c.handlers = engine.allNoRoute
	serveError(c, http.StatusNotFound, default404Body)
}
```
c.Next() 嵌套执行handler，包括中间件等
//源码注释：Next() should be used only inside middleware.
```go
func (c *Context) Next() {
	c.index++
	for c.index < int8(len(c.handlers)) {
		c.handlers[c.index](c)
		c.index++
	}
}
```
# 参数解析
<table>
<thead>
<tr>
<th>序号</th>
<th>参数类型</th>
<th>解释</th>
<th>Context支持</th>
</tr>
</thead>
<tbody>
<tr>
<td>1</td>
<td>path param</td>
<td>在URI中将参数作为路径的一部分</td>
<td>c.Param(&ldquo;key&rdquo;) string</td>
</tr>
<tr>
<td>2</td>
<td>query param</td>
<td>在URI中以&quot;?&ldquo;开始的，&ldquo;key=value&quot;形式的部分</td>
<td>c.Query(&ldquo;key&rdquo;) string</td>
</tr>
<tr>
<td>3</td>
<td>body [form; json; xml等等]</td>
<td>根据请求头Content-Type判定或指定</td>
<td>c.Bind类似函数</td>
</tr>
</tbody>
</table>

gin内置了众多的参数binder
```go
var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	ProtoBuf      = protobufBinding{}
	MsgPack       = msgpackBinding{}
	YAML          = yamlBinding{}
	Uri           = uriBinding{}
	Header        = headerBinding{}
)

func (c *Context) BindJSON(obj interface{}) error {
	return c.MustBindWith(obj, binding.JSON)
}
```
# 自定义参数验证
```go
package vali

import (
    "strconv"
    "strings"

    "github.com/gin-gonic/gin/binding"
    "github.com/go-playground/validator/v10"
)

type Register struct {
    Mobile   uint   `json:"mobile" binding:"required,checkMobile"`
    Password string `json:"password" binding:"required,gte=6"`
}

var v *validator.Validate
var trans ut.Translator

func InitVali() {
    v, ok := binding.Validator.Engine().(*validator.Validate)
    if ok {
        // 自定义验证方法
        v.RegisterValidation("checkMobile", checkMobile)
    }
}

func checkMobile(fl validator.FieldLevel) bool {
    mobile := strconv.Itoa(int(fl.Field().Uint()))
    re := `^1[3456789]\d{9}$`
    r := regexp.MustCompile(re)
    return r.MatchString(mobile)
}
```
# 响应
```go
// Render writes the response headers and calls render.Render to render data.
func (c *Context) Render(code int, r render.Render) {
	c.Status(code)

	if !bodyAllowedForStatus(code) {
		r.WriteContentType(c.Writer)
		c.Writer.WriteHeaderNow()
		return
	}

	if err := r.Render(c.Writer); err != nil {
		panic(err)
	}
}
```
内置render
```go
var (
	_ Render     = JSON{}
	_ Render     = IndentedJSON{}
	_ Render     = SecureJSON{}
	_ Render     = JsonpJSON{}
	_ Render     = XML{}
	_ Render     = String{}
	_ Render     = Redirect{}
	_ Render     = Data{}
	_ Render     = HTML{}
	_ HTMLRender = HTMLDebug{}
	_ HTMLRender = HTMLProduction{}
	_ Render     = YAML{}
	_ Render     = Reader{}
	_ Render     = AsciiJSON{}
	_ Render     = ProtoBuf{}
)
```


