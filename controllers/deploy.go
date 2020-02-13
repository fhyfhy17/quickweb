package controllers

import (
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"io"
	"log"
	"os/exec"
	"strings"
)

type DeployController struct {
	beego.Controller
}

type SelectModel struct {
	Name  string
	Value string
}

func (d *DeployController) Get() {
	d.TplName = "deploy.tpl"
}

func (d *DeployController) DoDeploy() {
	d.TplName = "deploy.tpl"
	pro := d.GetString("pro")
	tar := d.GetString("tar")
	branch := d.GetString("bra")

	if pro == "" || pro == "0" {
		d.Data["msg"] = "项目不能为空"
		return
	}
	if tar == "" || tar == "0" {
		d.Data["msg"] = "目标服务器不能为空"
		return
	}
	if branch == "" || branch == "0" {
		d.Data["msg"] = "分支不能为空"
		return
	}

	fmt.Println(pro)
	fmt.Println(tar)
	fmt.Println(branch)
	d.Data["msg"] = fmt.Sprintf("请求成功 ， 项目:%v , 目标服务器：%v ， 分支: %v", pro, tar, branch)
}

func (d *DeployController) GetBranches() {

	//objs := make([]map[string]interface{}, 3)
	//objs[0]=map[string]interface{}{"name":"name1","value":"value1"}
	//objs[1]=map[string]interface{}{"name":"name2","value":"value2"}
	//objs[2]=map[string]interface{}{"name":"name3","value":"value3"}

	//objs := `[{"name":"name1","value":"value1"},{"name":"name2","value":"value2"},{"name":"name3","value":"value3"}]`

	objs := make([]SelectModel, 3)
	objs[0] = SelectModel{"name1", "value1"}
	objs[1] = SelectModel{"name2", "value2"}
	objs[2] = SelectModel{"name3", "value3"}

	d.Data["json"] = objs
	d.ServeJSON()
}

func asyncLog(uuid string, reader io.ReadCloser) error {
	cache := "" //缓存不足一行的日志信息
	buf := make([]byte, 1024)
	for {
		num, err := reader.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if num > 0 {
			b := buf[:num]
			s := strings.Split(string(b), "\n")
			line := strings.Join(s[:len(s)-1], "\n") //取出整行的日志
			content := fmt.Sprintf("%v%v\n<br />", cache, line)
			fmt.Println(content)
			if ClientMap[uuid] != nil {
				ClientMap[uuid].WriteMessage(1, []byte(content))
			}
			cache = s[len(s)-1]
		}
	}
	return nil
}

func (d *DeployController) Execute() {
	uuid := d.GetString("uuid")
	cmd := exec.Command("sh", "-c", "~/scripts/curl.sh")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s......", err.Error())
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}

	go asyncLog(uuid, stdout)
	go asyncLog(uuid, stderr)

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}
	fmt.Println("执行成了！！！！！")
	d.Data["json"] = "执行成功"
	d.ServeJSON()
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ClientMap map[string]*websocket.Conn = make(map[string]*websocket.Conn)

func (c *DeployController) WebSocket() {
	uuid := c.GetString("uuid")

	conn, err := upgrader.Upgrade(c.Ctx.ResponseWriter, c.Ctx.Request, nil)
	if err != nil {
		fmt.Println(err)
		return
	}

	ClientMap[uuid] = conn
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Sprintf("websocket关闭，code=%d, text=%s", code, text)
		delete(ClientMap, uuid)
		return nil
	})

	//for {
	//	//msgType, msg, err := conn.ReadMessage()
	//	//if err != nil {
	//	//	fmt.Println(err)
	//	//	return
	//	//}
	//
	//		time.Sleep(200 * time.Millisecond)
	//		err = conn.WriteMessage(1, []byte("pong<br />"))
	//		if err != nil {
	//			fmt.Println(err)
	//			return
	//		}
	//
	//}
	c.Ctx.WriteString("")

}
