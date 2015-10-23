package eweb

import (
	"fmt"
	"github.com/afocus/eweb/render"
	"html/template"
	"net/http"
	"reflect"
	"regexp"
	"runtime/debug"
	"strings"
)

type ActionFunc func(*Context)

type routerMapStruct struct {
	Params []string
	Method string
	Action ActionFunc
}
type routerMaps struct {
	Control Controller
	List    map[string]routerMapStruct
}

type EWeb struct {
	//分组
	groupRouter map[string][]string
	//路由map
	routers map[string]routerMaps
	//静态文件路径
	StaticDir map[string]string
	//基础路径 目前还没实现 后期实现在公开
	basePath          string
	notFoundHandlFunc ActionFunc
	//默认控制器名称
	DefaultControlName string
	//
	render *render.Render
	debug  bool
}

func New() *EWeb {
	web := new(EWeb)
	web.routers = make(map[string]routerMaps, 0)
	web.StaticDir = map[string]string{
		"/static": "static/",
	}
	web.basePath = "/"
	web.DefaultControlName = "index"
	enableDebug := GetConfig("default").GetBool("app", "debug", true)
	web.render = render.New("templates/", enableDebug)
	web.debug = enableDebug
	return web
}

func GetVersion() string {
	return "0.0.5"
}
func (e *EWeb) RegisterTplFuncs(funcs template.FuncMap) {
	e.render.RegisterFuncs(funcs)
}

//处理静态文件
//使用http包的默认处理方式
func (e *EWeb) staticFile(path string, w http.ResponseWriter, r *http.Request) bool {
	if path == "/favicon.ico" {
		return true
	} else {
		for prefix, staticDir := range e.StaticDir {
			if strings.HasPrefix(path, prefix) {
				file := staticDir + path[len(prefix):]
				http.ServeFile(w, r, file)
				return true
			}
		}
		return false
	}
}

//统一清理不符合标准的路径
func cleanPath(path string) string {
	if path == "" {
		return "/"
	}
	if path[0] != '/' {
		path = "/" + path
	}
	strlen := len(path)
	buff := make([]rune, strlen)
	var i int
	var hasF bool
	for _, v := range path {
		if v == '/' {
			if !hasF {
				hasF = true
				buff[i] = '/'
				i++
			}
		} else {
			hasF = false
			buff[i] = v
			i++
		}
	}
	if buff[i-1] == '/' {
		return string(buff[:i-1])
	}
	return string(buff[:i])
}

func (e *EWeb) parseUrl(url string) (controlName, path string) {
	p := cleanPath(url[len(e.basePath):])
	paths := strings.Split(p, "/")
	controlName = e.DefaultControlName
	if len(paths) >= 2 && paths[1] != "" {
		controlName = paths[1]
	}
	path = "/" + strings.Join(paths[2:], "/")
	return
}

func panicCatch(e *EWeb, ctx *Context) {
	if err := recover(); err != nil {
		w := ctx.Writer
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.WriteHeader(http.StatusInternalServerError)
		LogError("%v", err)
		tpl := `
		<html>
		<head><meta charset="utf-8"/><title>服务器错误</title>
		<style>
		*{font-family:"微软雅黑";line-height:1.8}
		pre{background:#fefefe;padding:20px;margin-top:20px;color:#888;font-size:12px}
		h1{text-align:center;font-size:148px;line-height:1;padding-top:60px;margin:20px}
		h1>small{display:block;font-size:32px;color:#444}
		</style>
		</head>
		<body><h1>500<small>内部服务器错误</small></h1>
		<p style="font-weight:bold;font-size:14px;color:red;background:#ffe;padding:20px;">错误信息: %v</p>
		%s
		</body>
		</html>
		`
		stackInfo := ""
		if e.debug {
			stackstr := string(debug.Stack())
			stackInfo = fmt.Sprintf(`<pre>%s</pre>`, stackstr)
		}
		fmt.Fprintln(w, fmt.Sprintf(tpl, err, stackInfo))
		flusher, ok := w.(http.Flusher)
		if ok {
			flusher.Flush()
		}
	}
}

func doAction(control Controller, fun ActionFunc, ctx *Context) {
	if before, ok := control.(ControlBeforer); ok {
		if !before.Before(ctx) {
			return
		}
	}
	fun(ctx)
	if after, ok := control.(ControlAfter); ok {
		after.After(ctx)
	}
}

func (e *EWeb) routerToAction(ctx *Context, comap routerMaps, uripath, method string) bool {
	if co, has := comap.List[uripath]; has {
		//普通匹配
		if co.Method == method || co.Method == "*" {
			doAction(comap.Control, co.Action, ctx)
			return true
		}
	} else {
		//正则匹配
		for path, v := range comap.List {
			if v.Params != nil && (v.Method == method || v.Method == "*") {
				reg := regexp.MustCompile(path)
				if p := reg.FindStringSubmatch(uripath); p != nil {
					for k, v := range v.Params {
						ctx.Params[v] = p[k+1]
					}
					doAction(comap.Control, v.Action, ctx)
					return true
				}
			}
		}
	}
	return false
}

//http handler
func (e *EWeb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if e.staticFile(r.URL.Path, w, r) {
		return
	}
	LogInfo("request [%s] %s", r.Method, r.URL.Path)
	cname, uripath := e.parseUrl(r.URL.Path)
	ctx := &Context{
		Writer:      w,
		Request:     r,
		Params:      make(map[string]string),
		Data:        make(map[string]interface{}),
		Ins:         e,
		ControlName: cname,
	}
	//捕获异常进行崩溃恢复
	defer panicCatch(e, ctx)
	if comap, has := e.routers[cname]; has {
		if e.routerToAction(ctx, comap, uripath, r.Method) {
			return
		}
	}
	cname2, uripath2 := e.parseUrl(uripath)
	if comap, has := e.routers[cname+"/"+cname2]; has {
		if e.routerToAction(ctx, comap, uripath2, r.Method) {
			return
		}
	}
	e.NotFound(ctx)
}

func (e *EWeb) Register(groupname string, cs ...Controller) {
	groupname = strings.Split(groupname, "/")[1]
	for _, c := range cs {
		cname := strings.ToLower(c.GetName())
		if cname == "" {
			type_ := reflect.TypeOf(c)
			cname = strings.ToLower(type_.Elem().Name())
		}
		if groupname != "" {
			cname = groupname + "/" + cname
		}
		if _, has := e.routers[cname]; has {
			LogError("controlName:" + cname + " alreay registed")
			panic("controlName:" + cname + " alreay registed")
		}
		e.routers[cname] = routerMaps{
			Control: c,
			List:    make(map[string]routerMapStruct, 0),
		}

		for _, r := range c.GetRouter() {
			LogInfo("router>> %s %s", cname, r.Path)
			comstr := `([^/^\s.]+)`
			x := regexp.MustCompile(fmt.Sprintf("/:%s", comstr))
			params := x.FindAllString(r.Path, -1)
			if params != nil {
				//存在自定义参数
				for k, v := range params {
					params[k] = v[2:]
				}
				r.Path = x.ReplaceAllString(r.Path, fmt.Sprintf("/${n}%s", comstr))
			}
			e.routers[cname].List[strings.ToLower(r.Path)] = routerMapStruct{
				Params: params,
				Method: strings.ToUpper(r.Mehod),
				Action: r.Action,
			}
		}
	}
}

//设置404处理函数
func (e *EWeb) SetNotFound(handler ActionFunc) {
	e.notFoundHandlFunc = handler
}

func (e *EWeb) NotFound(ctx *Context) {
	ctx.Writer.WriteHeader(http.StatusNotFound)
	ctx.Writer.Header().Set("Content-type", render.CONTENTTYPE_HTML)
	if e.notFoundHandlFunc != nil {
		e.notFoundHandlFunc(ctx)
	} else {
		tpl := `
		<html>
		<head><meta charset="utf-8"/><title>未找到网页</title>
		<style>
		*{font-family:"微软雅黑";line-height:1.8}
		h1{text-align:center;font-size:148px;line-height:1;padding-top:60px;margin:20px}
		h1>small{display:block;font-size:32px;color:#444}
		</style>
		</head>
		<body><h1>404<small>未找到该网页 file not found</small></h1>
		</body>
		</html>
		`
		ctx.Writer.Write([]byte(tpl))
	}
}

func (e *EWeb) Run() {
	cfg := GetConfig("default")
	address := cfg.GetString("app", "address", ":8088")
	Log("---> run at %s", address)
	err := http.ListenAndServe(address, e)
	if err != nil {
		LogError("%v", err)
	}
}

func (e *EWeb) GetController(controlname string) Controller {
	if comap, has := e.routers[controlname]; has {
		return comap.Control
	}
	return nil
}

func (e *EWeb) SetDebug(enable bool) {
	e.debug = enable
}

type D struct {
	Message string      `json:"message"`
	Code    int         `json:"code"`
	Data    interface{} `json:"data"`
}
