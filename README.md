# eweb
##一个golang的web框架

适合api服务 轻量&清晰&速度

##特点
1. 和大多数框架不一样 eweb使用结构体产生路由 更加面向对象 (貌似现在流行函数式编程:)) 自动实现分组功能
2. 没有中间件 你可以使用结构体的匿名继承 来实现一些中间件的效果
3. 简单速度快、路由采用树形结构 定位块 除了必须的一点正则外 没有其他任何负担
4. 结构清晰 没有满屏的r.Get("/",..) r.Post("/",..)..


##状态
1. 目前仅是开始阶段 正在完善 请不要使用
2. 计划加入分组功能

##概念
网址由控制器(control)+动作(action)组成

默认结构体的名字是控制器的名字 入口控制器的名字必须是Index(当路由是"/"的时候会自动调用Index注册的方法)

#example
##最简单的例子
例子看起来比别的框架复杂 但是结构却很清晰
```go
package main

import "github.com/afocus/eweb"

//index是入口的控制器
type Index struct{
	*eweb.Control
}
//返回路由信息
//GET/POST/DELETE/PUT/.../*   `*`代表匹配所有
func (this Index) GetRouter()[]eweb.ControlRouter{
	return []eweb.ControlRouter{
		{"GET","/",this.Index},
		{"*","/:name/say",this.Say},
	}
}

func (Index) Index(ctx *eweb.Context){
	ctx.String(200,"hello,eweb")
}

func (Index) Say(ctx *eweb.Context){
	//得到路由里 /:name/say 里面的name
	name:= ctx.Param("name")
	ctx.JSON(200,map[string]interface{}{
		"name":name,
		"code":0,
	})
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

##自定义404
```go
e:=eweb.New()
e.SetNotFound(func(ctx *eweb.Context){
	ctx.String(200,"自定义404")
})
//使用
e.NotFound(ctx)
```

##设置静态资源目录
```go
e.StatisDir["res"] = "/res"
```
##不想使用结构体的名字映射到网址 可以使用GetName方法
```go
type DefaultControl struct{
	*eweb.Control
}
func (DefaultControl) GetName() string{
	return "index"
}
```
##修改默认控制器名称
```go
e.DefaultControlName = "home"
```
##做一些前置 后后置操作
流程是 before(ctx) -> yourAction(ctx)->after(ctx)
```go
type UserController struct{
	*eweb.Control
}
//返回true继续 false终止
func (UserController) Before(ctx *eweb.Context) bool {
	//todo检查权限
	//ctx.Set 可以设置上线文需要的数据 并通过ctx.Get获取
	ctx.Set("time",time.Now().Unix())
	ctx.String("你没有权限访问")
	return false
}
//后置操作 就算前置返回false 依然会执行
func (UserController) After(ctx *eweb.Context){
	//todo计算处理时间
	t:=ctx.Get("time")
	ctx.String("用时%d",time.Now().Unix()-t)
}
```