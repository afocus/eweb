package eweb

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

const (
	CONTENTTYPE_TEXT = "text/plain"
	CONTENTTYPE_JSON = "application/json"
	CONTENTTYPE_XML  = "application/xml"
)

type Context struct {
	Writer  http.ResponseWriter
	Request *http.Request
	Params  map[string]string
	Data    map[string]interface{}
	Ins     *EWeb
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
	ctx.Render([]byte(fmt.Sprintf(format, a...)), code, CONTENTTYPE_TEXT)
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
		ctx.Render([]byte(err.Error()), http.StatusOK, CONTENTTYPE_TEXT)
		return
	}
	ctx.Render(val, code, CONTENTTYPE_JSON)
}

func (ctx *Context) Xml(code int, data interface{}) {
	val, err := xml.Marshal(data)
	if err != nil {
		ctx.Render([]byte(err.Error()), http.StatusOK, CONTENTTYPE_TEXT)
		return
	}
	//xml默认不添加头部 这里补上
	xmls := []byte(xml.Header)
	ctx.Render(append(xmls, val...), code, CONTENTTYPE_XML)
}
