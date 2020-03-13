package controllers

import (
	"archive/zip"
	"bufio"
	"bytes"
	"fmt"
	"github.com/astaxie/beego"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
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
	//s = append(s[:len(s)-1])
	s = removeEmpty(s)
	objs := make([]SelectModel, len(s))
	for i := range s {
		s[i] = "branches/" + s[i]
		fmt.Println(s[i])
		objs[i] = SelectModel{s[i], s[i]}
	}
	objs = append(objs, SelectModel{"trunk/develop/", "trunk/develop/"})
	d.Data["json"] = objs
	d.ServeJSON()
}

func removeEmpty(s []string) []string {
	i := 0
	for v := range s {
		if s[v] == "" {
			continue
		}
		s[i] = s[v]
		i++
	}
	return s[:i]
}

func removeEmpty2(s []string) []string {
	newSlice := s[:0]
	for _, v := range s {
		if v != "" {
			newSlice = append(newSlice, v)
		}
	}
	return newSlice
}

func asyncLog(uuid string, reader io.ReadCloser) error {
	rd := bufio.NewReader(reader)
	for {
		content, err := rd.ReadString('\n')

		if err != nil && len(content) == 0 {
			break
		}
		fmt.Println(content)
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

	if pro == "" || pro == "0" || pro == "null" {
		returnMsg(d, "项目不能为空")
		//d.Data["msg"] = "项目不能为空"
		return
	}
	if tar == "" || tar == "0" || tar == "null" {
		returnMsg(d, "目标服务器不能为空")
		return
	}
	if branch == "" || branch == "0" || branch == "null" {
		returnMsg(d, "分支不能为空")
		return
	}

	fmt.Println(pro)
	fmt.Println(tar)
	fmt.Println(branch)

	var shName string

	if pro == "project" {
		shName = "/home/jenkins/genBranches/deploy2.sh"
	} else {
		shName = "/home/jenkins/genBranches/deployGM2.sh"
	}

	_ = shName
	cmd := exec.Command("sh", "-c", shName+" "+branch+" "+tar) //beego.AppConfig.String(tar))
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

func (c *DeployController) ReceiveFile() {

	host := c.GetString("remote_ip_down")
	downPath := c.GetString("downPath")

	if host == "" {
		host = "192.168.1.35"
	}
	port := getPort(host)

	var (
		err        error
		sftpClient *sftp.Client
	)
	sshKeyPath := beego.AppConfig.String("sshKeyPath")

	// SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err = connect("root", "123.com", host,
		port, "key", sshKeyPath)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	// 完整的远程文件路径 和 本地文件夹
	var remoteFilePath = downPath
	cachePath := "static/down"

	msg := ""

	//srcFile, err := sftpClient.Open(remoteFilePath)
	//if err != nil {
	//	msg = fmt.Sprintln(err)
	//	returnMsg(c, msg)
	//	return
	//}
	os.Mkdir(cachePath, os.ModePerm)

	walk := sftpClient.Walk(remoteFilePath)
	for walk.Step() {
		stat := walk.Stat()
		if stat.IsDir() {
			os.Mkdir(filepath.Join(cachePath, walk.Path()), os.ModePerm)
		} else {
			fmt.Println(walk.Path(), "  ----")
			fmt.Println(stat.Name(), " =====")
			srcFile, err := sftpClient.Open(walk.Path())
			if err != nil {
				msg = fmt.Sprintln(err)
				returnMsg(c, msg)
				return
			}

			defer srcFile.Close()
			addr := filepath.Join(cachePath, filepath.Base(walk.Path()))
			dstFile, err := os.Create(addr)
			if err != nil {
				msg = fmt.Sprintln(err)
				returnMsg(c, msg)
				return
			}
			defer dstFile.Close()

			if _, err = srcFile.WriteTo(dstFile); err != nil {
				msg = fmt.Sprintln(err)
				returnMsg(c, msg)
				return
			}
		}
	}

	msg = fmt.Sprintln("取得文件成功!")
	fmt.Println(msg)
	downFileName := filepath.Base(downPath) + ".tar"
	err = Zip(cachePath, downFileName, "")
	if err != nil {
		msg = fmt.Sprintln(err)
		returnMsg(c, msg)
		return
	}
	err = os.RemoveAll(cachePath)
	fmt.Println(err)
	defer os.Remove(downFileName)
	c.Ctx.Output.Download(downFileName, filepath.Base(downPath)+".tar")
}

func delAllFilesOfDir(removePath string) {
	filepath.Walk(removePath, func(path string, info os.FileInfo, err error) error {

		if path == removePath {
			return nil
		}
		if !Exists(path) {
			return nil
		}
		if info.IsDir() {
			os.RemoveAll(path)
			return nil
		}
		fmt.Println(path + "~~~~")
		os.Remove(path)

		return nil
	})
}
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

func (c *DeployController) SendFile() {
	var (
		err        error
		sftpClient *sftp.Client
	)
	files, err := c.GetFiles("upFile")
	if err != nil {
		msg := fmt.Sprintln(err)
		returnMsg(c, msg)
		return
	}
	remoteIp := c.GetString("remote_ip")
	upPath := c.GetString("upPath")

	//ext := path.Ext(h.Filename)
	//验证后缀名是否符合要求
	//var AllowExtMap map[string]bool = map[string]bool{
	//	".jpg":  true,
	//	".jpeg": true,
	//	".png":  true,
	//	".gz":   true,
	//}
	//if _, ok := AllowExtMap[ext]; !ok {
	//	c.Ctx.WriteString("后缀名不符")
	//	return
	//}

	sshKeyPath := beego.AppConfig.String("sshKeyPath")
	msg := ""
	port := getPort(remoteIp)
	// 这里换成实际的 SSH 连接的 用户名，密码，主机名或IP，SSH端口
	sftpClient, err = connect("root", "123.com", remoteIp,
		port, "key", sshKeyPath)
	if err != nil {
		msg = fmt.Sprintln(err)
		returnMsg(c, msg)
		return
	}
	defer sftpClient.Close()

	// 用来测试的本地文件路径 和 远程机器上的文件夹

	var remoteDir = upPath

	for _, fh := range files {
		f, err := fh.Open()
		if err != nil {
			msg = fmt.Sprintln(err)
			returnMsg(c, msg)
			return
		}

		defer f.Close()
		slash := filepath.ToSlash(filepath.Join(remoteDir, filepath.Base(fh.Filename)))
		dstFile, err := sftpClient.Create(slash)
		if err != nil {
			msg = fmt.Sprintln(err)
			returnMsg(c, msg)
			return
		}
		defer dstFile.Close()

		bs, err := ioutil.ReadAll(f)
		if err != nil {
			msg = fmt.Sprintln(err)
			returnMsg(c, msg)
			return
		}
		dstFile.Write(bs)
	}

	msg = fmt.Sprintln("发送文件成功!")
	fmt.Println(msg)
	returnMsg(c, msg)
}

func getPort(ip string) int {
	port := 52236
	if ip == "192.168.1.35" || ip == "192.168.1.105" {
		port = 22
	}
	return port
}

func returnMsg(c *DeployController, msg string) {
	fmt.Println(msg)
	c.Ctx.WriteString(msg)
}

func connect(user, password, host string, port int, sshType string, sshKeyPath string) (*sftp.Client, error) {
	var (
		auth         []ssh.AuthMethod
		addr         string
		clientConfig *ssh.ClientConfig
		sshClient    *ssh.Client
		sftpClient   *sftp.Client
		err          error
	)
	// get auth method
	auth = make([]ssh.AuthMethod, 0)
	auth = append(auth, ssh.Password(password))

	clientConfig = &ssh.ClientConfig{
		User:    user,
		Auth:    auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	if sshType != "password" {
		clientConfig.Auth = []ssh.AuthMethod{publicKeyAuth(sshKeyPath)}
	}

	// ssh连接
	addr = fmt.Sprintf("%s:%d", host, port)
	fmt.Println("准备连接ssh服务器，sshType=", sshType, " sshKeyPath=", sshKeyPath, " ip:port=", addr)
	if sshClient, err = ssh.Dial("tcp", addr, clientConfig); err != nil {
		return nil, err
	}

	// create sftp client
	if sftpClient, err = sftp.NewClient(sshClient); err != nil {
		return nil, err
	}

	return sftpClient, nil
}

//密钥登录
func publicKeyAuth(keyPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(keyPath)
	if err != nil {
		log.Fatal("ssh key file read failed", err)
	}
	// Create the Signer for this private key.
	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatal("ssh key signer failed", err)
	}
	return ssh.PublicKeys(signer)
}

func (d *DeployController) PushToFormal() {
	host := d.GetString("push_remote_ip")
	if host == "" || host == "0" || host == "null" {
		returnMsg(d, "目标IP不能为空")
		return
	}
	fmt.Println(host)
	var shName string
	if host == "35.196.251.9" {
		shName = "/home/jenkins/genBranches/pushTiShenToFormal.sh"
	} else {
		shName = "/home/jenkins/genBranches/pushawstestToFormal.sh"
	}

	cmd := exec.Command("sh", "-c", shName+" "+host)
	//cmd := exec.Command("sh", "-c", "~/scripts/curl.sh")

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting command: %s......", err.Error())
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}

	go func() {
		rd := bufio.NewReader(stdout)
		for {
			content, err := rd.ReadString('\n')

			if err != nil && len(content) == 0 {
				break
			}
			fmt.Println(content)
		}
	}()
	go func() {
		rd := bufio.NewReader(stderr)
		for {
			content, err := rd.ReadString('\n')

			if err != nil && len(content) == 0 {
				break
			}
			fmt.Println(content)
		}
	}()

	if err := cmd.Wait(); err != nil {
		log.Printf("Error waiting for command execution: %s......", err.Error())
		d.Data["json"] = "执行出错"
		d.ServeJSON()
		return
	}
	d.Data["json"] = "执行成功"
	d.ServeJSON()
}

func Zip(source, target, filter string) error {
	var err error
	if isAbs := filepath.IsAbs(source); !isAbs {
		source, err = filepath.Abs(source) // 将传入路径直接转化为绝对路径
		if err != nil {
			return errors.WithStack(err)
		}
	}
	//创建zip包文件
	zipfile, err := os.Create(target)
	if err != nil {
		return errors.WithStack(err)
	}

	defer func() {
		if err := zipfile.Close(); err != nil {
			log.Fatalf("*File close error: %s, file: %s", err.Error(), zipfile.Name())
		}
	}()

	//创建zip.Writer
	zw := zip.NewWriter(zipfile)

	defer func() {
		if err := zw.Close(); err != nil {
			log.Fatalf("zipwriter close error: %s", err.Error())
		}
	}()

	info, err := os.Stat(source)
	if err != nil {
		return errors.WithStack(err)
	}

	var baseDir string
	if info.IsDir() {
		baseDir = filepath.Base(source)
	}

	err = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {

		if err != nil {
			return errors.WithStack(err)
		}

		if filter != "" {
			ism, err := filepath.Match(filter, info.Name())

			if err != nil {
				return errors.WithStack(err)
			}
			if ism {
				return nil
			}
		}

		//创建文件头
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return errors.WithStack(err)
		}

		if baseDir != "" {
			header.Name = filepath.Join(baseDir, strings.TrimPrefix(path, source))
		}

		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		//写入文件头信息
		writer, err := zw.CreateHeader(header)
		if err != nil {
			return errors.WithStack(err)
		}

		if info.IsDir() {
			return nil
		}
		//写入文件内容
		file, err := os.Open(path)
		if err != nil {
			return errors.WithStack(err)
		}

		defer file.Close()

		_, err = io.Copy(writer, file)

		return errors.WithStack(err)
	})

	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}
