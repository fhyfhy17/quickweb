package main

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/config"
	_ "quickweb/routers"
)

func main() {
	initConfig()
	beego.Run()
}

func initConfig() {
	initConf, err := config.NewConfig("ini", "conf/init.conf")
	if err != nil {
		fmt.Println(err.Error())
	}
	confName := initConf.String("confName")
	beego.LoadAppConfig("ini", "conf/"+confName)
}
