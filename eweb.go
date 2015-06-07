package eweb

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"regexp"
	"strings"
)

type routerMapStruct struct {
	Params []string
	Method string
	Action func(*Context)
}
type routerMaps struct {
	Control Controller
	List    map[string]routerMapStruct
}

type EWeb struct {
	//路由map
	routers map[string]routerMaps
	//静态文件路径
	StaticDir map[string]string
	//基础路径 目前还没实现 后期实现在公开
	//模板路径
	TemplateDir string
	basePath    string
}

func New() *EWeb {
	web := new(EWeb)
	web.routers = make(map[string]routerMaps)
	web.StaticDir = make(map[string]string, 0)
	web.basePath = "/"
	web.TemplateDir = "./view"
	return web
}

func GetVersion() string {
	return "0.0.1"
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
	controlName = "index"
	if len(paths) >= 2 && paths[1] != "" {
		controlName = paths[1]
	}
	path = "/" + strings.Join(paths[2:], "/")
	return
}

//http handler
func (e *EWeb) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.ToLower(r.URL.Path)
	log.Printf("[%s] %s\r\n", r.Method, path)
	if e.staticFile(path, w, r) {
		return
	}
	ctx := &Context{
		Writer:  w,
		Request: r,
		Params:  make(map[string]string),
		Data:    make(map[string]interface{}),
		Ins:     e,
	}
	cname, uripath := e.parseUrl(path)
	if comap, has := e.routers[cname]; has {
		if co, has := comap.List[uripath]; has {
			//普通匹配
			if co.Method == r.Method || co.Method == "*" {
				if comap.Control.Before(ctx) {
					co.Action(ctx)
				}
				comap.Control.After(ctx)
				return
			}
		} else {
			//正则匹配
			for path, v := range comap.List {
				if v.Params != nil && (v.Method == r.Method || v.Method == "*") {
					reg := regexp.MustCompile(path)
					if p := reg.FindStringSubmatch(uripath); p != nil {
						for k, v := range v.Params {
							ctx.Params[v] = p[k+1]
						}
						if comap.Control.Before(ctx) {
							v.Action(ctx)
						}
						comap.Control.After(ctx)
						return
					}
				}
			}
		}
	}
	http.NotFound(w, r)
}

func (e *EWeb) Register(cs ...Controller) {
	for _, c := range cs {
		cname := strings.ToLower(c.GetName())
		if cname == "" {
			type_ := reflect.TypeOf(c)
			cname = strings.ToLower(type_.Elem().Name())
		}
		if _, has := e.routers[cname]; has {
			panic("controlName:" + cname + " alreay registed")
		}
		e.routers[cname] = routerMaps{
			Control: c,
			List:    make(map[string]routerMapStruct, 0),
		}
		for _, r := range c.GetRouter() {
			log.Println("router>>", cname, r.Path)
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

func (e *EWeb) Run(addr string) {
	http.ListenAndServe(addr, e)
}
