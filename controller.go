/*
定义一个基础的control实现controller接口 以后所有的控制器可以继承此control
*/
package eweb

type Controller interface {
	GetName() string
	GetRouter() []ControlRouter
	Before(ctx *Context) bool
	After(ctx *Context)
}

type ControlRouter struct {
	Mehod  string
	Path   string
	Action ActionFunc
}

type Control struct {
}

func (c *Control) GetName() string {
	return ""
}

func (c *Control) Before(ctx *Context) bool {
	return true
}
func (c *Control) After(ctx *Context) {
}
