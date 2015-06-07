package eweb

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	Data    map[string]interface{}
	Ins     *EWeb
}

func (ctx *Context) Param(key string) string {
	if val, has := ctx.Params[key]; has {
		return val
	}
	return ""
}

func (ctx *Context) Set(key string, value interface{}) {
	ctx.Data[key] = value
}

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

//类似php $_GET
func (ctx *Context) Query(key string) string {
	return ctx.Request.URL.Query().Get(key)
}

//类似php $_POST
func (ctx *Context) PostFrom(key string) string {
	return ctx.Request.PostFormValue(key)
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

//渲染相关
func (ctx *Context) String(format string, a ...interface{}) {
	ctx.Writer.Write([]byte(fmt.Sprintf(format, a...)))
}
func (ctx *Context) Html(path string, data interface{}) {
	path = ctx.Ins.TemplateDir + "/" + path
	t, err := template.ParseFiles(path)
	if err != nil {
		ctx.Writer.WriteHeader(500)
		log.Println(err.Error())
		return
	}
	err = t.Execute(ctx.Writer, data)
	if err != nil {
		ctx.Writer.WriteHeader(500)
		log.Println(err.Error())
	}
}
func (ctx *Context) Json(data interface{}) {
	val, err := json.Marshal(data)
	if err != nil {
		val = []byte("{}")
		log.Println(err.Error())
		return
	}
	ctx.Writer.Write(val)
}
