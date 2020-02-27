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
	beego.Router("/deploy/TT", &controllers.DeployController{}, "get:TT")
	beego.Router("/deploy/log", &controllers.DeployController{}, "get:WebSocket")
	beego.Router("/deploy/ReceiveFile", &controllers.DeployController{}, "get:ReceiveFile")
	beego.Router("/deploy/SendFile", &controllers.DeployController{}, "post:SendFile")
	beego.Router("/deploy/PushToFormal", &controllers.DeployController{}, "get:PushToFormal")

}
