package controllers

import (
	"bufio"
	"bytes"
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

	c := "svn://192.168.1.105/honeybadger/solitaire/branches"
	cmd := exec.Command("svn", "list", c)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return
	}
	str := out.String()
	s := strings.Split(str, "\n")
	i := len(s) - 1
	split := append(s[:i])
	objs := make([]SelectModel, len(split))
	for i := range split {
		split[i] = "branches/" + split[i]
		fmt.Println(split[i])
		objs[i] = SelectModel{split[i], split[i]}
	}
	objs = append(objs, SelectModel{"trunk/develop/", "trunk/develop/"})
	d.Data["json"] = objs
	d.ServeJSON()
}

func asyncLog(uuid string, reader io.ReadCloser) error {
	rd := bufio.NewReader(reader)
	for {
		content, err := rd.ReadString('\n')

		if err != nil && len(content) == 0 {
			break
		}

		fmt.Println("--------" + content)
		if ClientMap[uuid] != nil {
			ClientMap[uuid].WriteMessage(1, []byte(content+"<br />"))
		}
	}
	return nil
}

func (d *DeployController) Execute() {
	uuid := d.GetString("uuid")
	pro := d.GetString("pro")
	tar := d.GetString("tar")
	branch := d.GetString("bra")

	fmt.Println(pro)
	fmt.Println(tar)
	fmt.Println(branch)

	var shName string

	if pro == "project" {
		shName = "/home/jenkins/genBranches/ddddddd.sh"
	} else {
		shName = "/home/jenkins/genBranches/ddddddd.sh"
	}

	_ = shName
	cmd := exec.Command("sh", "-c", shName+" "+branch+" "+beego.AppConfig.String(tar))
	//cmd := exec.Command("sh", "-c", "~/scripts/curl.sh")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s......", err.Error())
		wsClose(uuid)
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}

	go asyncLog(uuid, stdout)
	go asyncLog(uuid, stderr)

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		wsClose(uuid)
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}
	wsClose(uuid)
	d.Data["json"] = "执行成功"
	d.ServeJSON()
}
func wsClose(uuid string) {
	fmt.Println("ws准备关闭")
	if ClientMap[uuid] != nil {
		ClientMap[uuid].Close()
		fmt.Println("ws关闭")
	}
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
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Sprintf("websocket关闭，code=%d, text=%s", code, text)
		delete(ClientMap, uuid)
		c.Ctx.WriteString("")
		return nil
	})
	ClientMap[uuid] = conn

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			//fmt.Println(err)
			c.Ctx.WriteString("")
			return
		}
		_ = msgType
		_ = msg
	}

	c.Ctx.WriteString("")

}
func (c *DeployController) TT() {
	fmt.Println(beego.AppConfig.String("aa"))
}
