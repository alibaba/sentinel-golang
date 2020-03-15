package beego

import (
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/context"
)

func NewSentinelFilterFunc(filterFunc beego.FilterFunc, opts ...Option) beego.FilterFunc {
	options := evaluateOptions(opts)
	return func(ctx *context.Context) {
		var resourceName = ctx.Input.Method() + ":" + ctx.Input.URL()
		if routerPattern, ok := ctx.Input.GetData("RouterPattern").(string); ok {
			resourceName = ctx.Input.Method() + ":" + routerPattern
		}

		if options.resourceExtract != nil {
			resourceName = options.resourceExtract(ctx)
		}

		entry, err := sentinel.Entry(
			resourceName,
			sentinel.WithResourceType(base.ResTypeWeb),
			sentinel.WithTrafficType(base.Inbound),
		)

		if err != nil {
			if options.blockFallback != nil {
				options.blockFallback(ctx)
			} else {
				ctx.Abort(http.StatusTooManyRequests, err.Error())
			}
			return
		}
		defer entry.Exit()
		filterFunc(ctx)
	}
}

func NewSentinelController(controllerInterface beego.ControllerInterface, opts ...Option) beego.ControllerInterface {
	options := evaluateOptions(opts)
	return &sentinelController{
		ControllerInterface: controllerInterface,
		Opts:                options,
	}
}

type sentinelController struct {
	beego.ControllerInterface
	Ctx  *context.Context
	Opts *options
}

func (sc *sentinelController) intercept(handler func()) {
	var resourceName = sc.Ctx.Input.Method() + ":" + sc.Ctx.Input.URL()
	if routerPattern, ok := sc.Ctx.Input.GetData("RouterPattern").(string); ok {
		resourceName = sc.Ctx.Input.Method() + ":" + routerPattern
	}

	if sc.Opts.resourceExtract != nil {
		resourceName = sc.Opts.resourceExtract(sc.Ctx)
	}

	entry, err := sentinel.Entry(
		resourceName,
		sentinel.WithResourceType(base.ResTypeWeb),
		sentinel.WithTrafficType(base.Inbound),
	)

	if err != nil {
		if sc.Opts.blockFallback != nil {
			sc.Opts.blockFallback(sc.Ctx)
		} else {
			sc.Ctx.Abort(http.StatusTooManyRequests, err.Error())
		}
		return
	}
	defer entry.Exit()
	handler()
}

func (sc *sentinelController) Init(ct *context.Context, controllerName, actionName string, app interface{}) {
	sc.ControllerInterface.Init(ct, controllerName, actionName, app)
	sc.Ctx = ct
}

func (sc *sentinelController) Get() {
	sc.intercept(sc.ControllerInterface.Get)
}
func (sc *sentinelController) Post() {
	sc.intercept(sc.ControllerInterface.Post)
}
func (sc *sentinelController) Delete() {
	sc.intercept(sc.ControllerInterface.Delete)
}
func (sc *sentinelController) Put() {
	sc.intercept(sc.ControllerInterface.Put)
}
func (sc *sentinelController) Head() {
	sc.intercept(sc.ControllerInterface.Head)
}
func (sc *sentinelController) Patch() {
	sc.intercept(sc.ControllerInterface.Patch)
}
func (sc *sentinelController) Options() {
	sc.intercept(sc.ControllerInterface.Options)
}
