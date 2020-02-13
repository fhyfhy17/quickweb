package routers

import (
	"github.com/astaxie/beego"
	"quickweb/controllers"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/deploy", &controllers.DeployController{})
	beego.Router("/deploy/doDeploy", &controllers.DeployController{}, "get:DoDeploy")
	beego.Router("/deploy/getBranches", &controllers.DeployController{}, "get:GetBranches")
	beego.Router("/deploy/Execute", &controllers.DeployController{}, "get:Execute")
	beego.Router("/deploy/log", &controllers.DeployController{}, "get:WebSocket")
}
