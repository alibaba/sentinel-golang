package beego

import (
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

type MainController struct {
	beego.Controller
}

func (ctl *MainController) Get() {
	ctl.Ctx.WriteString("bar " + ctl.Ctx.Input.Param("id"))
}

func Example() {
	beego.Router("/bar/:id", NewSentinelController(&MainController{}))
	beego.Get("/foo/:id", NewSentinelFilterFunc(func(ctx *context.Context) {
		ctx.WriteString("foo " + ctx.Input.Param("id"))
	}))
	beego.Run()
}
