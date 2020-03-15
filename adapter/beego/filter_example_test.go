package beego

import (
	"github.com/astaxie/beego"
)

type MainController struct {
	beego.Controller
}

func (ctl *MainController) Get() {
	ctl.Ctx.WriteString("hello world")
}

func Example() {
	beego.Router("/hello", &MainController{})
	beforeExec, finishRouter := SentinelFilters()
	beego.InsertFilter("/hello", beego.BeforeExec, beforeExec, false)
	beego.InsertFilter("/hello", beego.FinishRouter, finishRouter, false)
	beego.Run()
}
