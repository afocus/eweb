package eweb

import (
	"bytes"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

type Context struct {
	Writer      http.ResponseWriter
	Request     *http.Request
	Params      map[string]string
	Data        map[string]interface{}
	Ins         *EWeb
	ControlName string
}

func (ctx *Context) HTMLEscapeString(src string) string {
	return template.HTMLEscapeString(src)
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

func (ctx *Context) GetCookie(key string) string {
	ck, err := ctx.Request.Cookie(key)
	if err != nil {
		return ""
	}
	return ck.Value
}

//setcookie come from beego
var cookieNameSanitizer = strings.NewReplacer("\n", "-", "\r", "-")
var cookieValueSanitizer = strings.NewReplacer("\n", " ", "\r", " ", ";", " ")

func sanitizeName(n string) string {
	return cookieNameSanitizer.Replace(n)
}
func sanitizeValue(v string) string {
	return cookieValueSanitizer.Replace(v)
}
func (ctx *Context) SetCookie(name string, value string, others ...interface{}) {
	var b bytes.Buffer
	fmt.Fprintf(&b, "%s=%s", sanitizeName(name), sanitizeValue(value))

	//fix cookie not work in IE
	if len(others) > 0 {
		switch v := others[0].(type) {
		case int:
			if v > 0 {
				fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d", time.Now().Add(time.Duration(v)*time.Second).UTC().Format(time.RFC1123), v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int64:
			if v > 0 {
				fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d", time.Now().Add(time.Duration(v)*time.Second).UTC().Format(time.RFC1123), v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		case int32:
			if v > 0 {
				fmt.Fprintf(&b, "; Expires=%s; Max-Age=%d", time.Now().Add(time.Duration(v)*time.Second).UTC().Format(time.RFC1123), v)
			} else if v < 0 {
				fmt.Fprintf(&b, "; Max-Age=0")
			}
		}
	}

	// the settings below
	// Path, Domain, Secure, HttpOnly
	// can use nil skip set

	// default "/"
	if len(others) > 1 {
		if v, ok := others[1].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Path=%s", sanitizeValue(v))
		}
	} else {
		fmt.Fprintf(&b, "; Path=%s", "/")
	}
	// default empty
	if len(others) > 2 {
		if v, ok := others[2].(string); ok && len(v) > 0 {
			fmt.Fprintf(&b, "; Domain=%s", sanitizeValue(v))
		}
	}
	// default empty
	if len(others) > 3 {
		var secure bool
		switch v := others[3].(type) {
		case bool:
			secure = v
		default:
			if others[3] != nil {
				secure = true
			}
		}
		if secure {
			fmt.Fprintf(&b, "; Secure")
		}
	}
	// default false. for session cookie default true
	httponly := false
	if len(others) > 4 {
		if v, ok := others[4].(bool); ok && v {
			// HttpOnly = true
			httponly = true
		}
	}
	if httponly {
		fmt.Fprintf(&b, "; HttpOnly")
	}
	ctx.Writer.Header().Add("Set-Cookie", b.String())
}
