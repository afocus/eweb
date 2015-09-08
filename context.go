package eweb

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
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

//优先url的参数 返回字符串
//如果具体的值是字符串数组 请使用ctx.request.url.query()[key]类似的方式
func (ctx *Context) Query(key string) string {
	if val := ctx.Request.URL.Query().Get(key); val != "" {
		return val
	}
	if ctx.Request.Form == nil {
		ctx.Request.ParseForm()
	}
	return ctx.Request.Form.Get(key)
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

func (ctx *Context) IsAjax() bool {
	return ctx.Request.Header.Get("X-Requested-With") == "XMLHttpRequest"
}

func (ctx *Context) IsUpload() bool {
	return ctx.Request.Header.Get("Content-Type") == "multipart/form-data"
}

//渲染相关
func (ctx *Context) Render(val []byte, code int, typestr string) {
	ctx.Writer.WriteHeader(code)
	ctx.Writer.Header().Set("Content-Type", typestr)
	_, err := ctx.Writer.Write(val)
	if err != nil {
		http.Error(ctx.Writer, err.Error(), http.StatusInternalServerError)
	}
}

func (ctx *Context) String(code int, format string, a ...interface{}) {
	ctx.Render([]byte(fmt.Sprintf(format, a...)), code, "text/plain")
}

func (ctx *Context) Html(code int, path string, data interface{}) {
	path = ctx.Ins.TemplateDir + "/" + path
	t, err := template.ParseFiles(path)
	if err == nil {
		err = t.Execute(ctx.Writer, data)
		if err == nil {
			return
		}
	}
	ctx.Writer.Write([]byte(err.Error()))
}

func (ctx *Context) Json(code int, data interface{}) {
	val, err := json.Marshal(data)
	if err != nil {
		ctx.Render([]byte(err.Error()), http.StatusOK, "text/plain")
		return
	}
	ctx.Render(val, code, "application/json")
}

func (ctx *Context) Xml(code int, data interface{}) {
	val, err := xml.Marshal(data)
	if err != nil {
		ctx.Render([]byte(err.Error()), http.StatusOK, "text/plain")
		return
	}
	//xml默认不添加头部 这里补上
	xmls := []byte(xml.Header)
	ctx.Render(append(xmls, val...), code, "application/xml")
}
