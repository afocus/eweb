# eweb

使用结构体产生路由
没有中间件 使用结构体集成 很方便实现中间件的功能 也更加清晰和面向对象



#注意 目前代码处于开发状态 请不要使用生产环境


#example
##最简单的例子
默认使用结构体的名字当做路由控制器的名字
```go
package main

import "github.com/afocus/eweb"

//index是入口的控制器
type Index struct{
	*e.Control
}
//返回接口下面方法的路由
func (this *Index) GetRouter()[]e.ControlRouter{
	return []e.ControlRouter{
		{"GET","/",this.Index},
		{"GET","/:name/say",this.Say},
	}
}

func (this *Index) Index(ctx *e.Context){
	ctx.String("hello,eweb")
}

func (this *Index) Say(ctx *e.Context){
	//得到路由里 /:name/say 里面的name
	name:=ctx.Param("name")
	ctx.String(name+" say: hello")
}

func main(){
	e:=eweb.New()
	//注册控制器
	e.Register(new(Index))
	e.Run(":8080")
}

```


##注册多个控制器
```go
func main(){
	e:=eweb.New()
	//注册控制器
	e.Register(new(Index),new(Home),new(User))
	e.Run(":8080")
}
```

##控制器名称太长 想重新定义
```go
type UserController struct{
	*e.Control
}
func (*UserController) GetName() string{
	return "u"
}
```

##做一些前置 后后置操作
```go
type UserController struct{
	*e.Control
}
//返回true继续 false终止
func (*UserController) Before(ctx *e.Context) bool {
	//todo检查权限
	ctx.String("你没有权限访问")
	ctx.Set("time",time.Now().Unix())
	return false
}
//后置操作 就算前置返回false 依然会执行
func (*UserController) After(ctx *e.Context){
	//todo计算处理时间
	t:=ctx.Get("time")
	ctx.String("用时%d",time.Now().Unix()-t)
}
```