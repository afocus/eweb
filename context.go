package eweb

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Context struct {
	Writer      http.ResponseWriter
	Request     *http.Request
	Params      map[string]string
	Data        map[string]interface{}
	Ins         *EWeb
	ControlName string
}

// 获取url router 绑定的类似:xxx中xxx的具体值
func (ctx *Context) Param(key string) string {
	if val, has := ctx.Params[key]; has {
		return val
	}
	return ""
}

// ctx.Request.URL.Query().Get(key[0])的快捷方式
// key[1] 当key[0]不存在时的默认值
func (ctx *Context) Query(key ...string) string {
	val := ctx.Request.URL.Query().Get(key[0])
	if len(key) > 1 && len(val) == 0 {
		val = key[1]
	}
	return val
}
func (ctx *Context) QueryInt(key ...string) int {
	n, _ := strconv.Atoi(ctx.Query(key...))
	return n
}

//同上
func (ctx *Context) PostForm(key ...string) string {
	ctx.Request.ParseMultipartForm(32 << 20) //最大post数据的大小 32mb
	if values := ctx.Request.PostForm[key[0]]; len(values) > 0 {
		return values[0]
	}
	if ctx.Request.MultipartForm != nil && ctx.Request.MultipartForm.File != nil {
		if values := ctx.Request.MultipartForm.Value[key[0]]; len(values) > 0 {
			return values[0]
		}
	}
	if len(key) == 1 {
		return ""
	} else {
		return key[1]
	}
}

//设置共享数据 请勿在control结构体里定义属性并且修改它 如果要共享数据请使用ctx.set/get
func (ctx *Context) Set(key string, value interface{}) {
	ctx.Data[key] = value
}

//获取共享数据
func (ctx *Context) Get(key string) interface{} {
	if val, has := ctx.Data[key]; has {
		return val
	}
	return nil
}

//跳转 跳转后面的代码依然会继续执行 不想执行请return
func (ctx *Context) Redirect(path string, code int) {
	http.Redirect(ctx.Writer, ctx.Request, path, code)
}

//获取客户端ip
func (ctx *Context) ClientIP() string {
	clientIP := strings.TrimSpace(ctx.Request.Header.Get("X-Real-IP"))
	if len(clientIP) > 0 {
		return clientIP
	}
	clientIP = ctx.Request.Header.Get("X-Forwarded-For")
	clientIP = strings.TrimSpace(strings.Split(clientIP, ",")[0])
	if len(clientIP) > 0 {
		return clientIP
	}
	return strings.TrimSpace(ctx.Request.RemoteAddr)
}

//判断是否ajax请求
func (ctx *Context) IsAjax() bool {
	return ctx.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

//判断是否是上传
func (ctx *Context) IsUpload() bool {
	return ctx.Request.Header.Get("Content-Type") == "multipart/form-data"
}

func (ctx *Context) String(code int, format string, a ...interface{}) {
	err := ctx.Ins.render.Render(ctx.Writer, "", code, []byte(fmt.Sprintf(format, a...)))
	panic(err)
}

func (ctx *Context) Html(code int, tplname string, data interface{}) {
	err := ctx.Ins.render.Html(ctx.Writer, code, tplname, data, ctx.ControlName)
	panic(err)
}

func (ctx *Context) Json(code int, data interface{}) {
	err := ctx.Ins.render.Json(ctx.Writer, code, data)
	panic(err)
}

func (ctx *Context) Xml(code int, data interface{}) {
	err := ctx.Ins.render.Xml(ctx.Writer, code, data)
	panic(err)
}
