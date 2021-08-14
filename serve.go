package main

import (
	"database/sql"
	"fmt"
	"net"
	"net/url"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var dbFile string
var db *sql.DB
var msgdb *sql.DB
var Skey map[int]string

func main() {

	//数据库文件目录
	dbFile = "./yunchat.db"
	//监听端口
	listenPort := "3521"
	listenPort1 := "3522"
	Skey = make(map[int]string)
	//判断db是否存在
	if !isFileExist(dbFile) {
		serveInit(dbFile)
	}

	var err error
	db, err = sql.Open("sqlite3", dbFile)
	checkErr(err)

	fmt.Println("开始监听.....")
	tcpAd, _ := net.ResolveTCPAddr("tcp4", ":"+listenPort)
	listener, err := net.ListenTCP("tcp", tcpAd)
	checkErr(err)

	tcpAd1, _ := net.ResolveTCPAddr("tcp4", ":"+listenPort1)
	listener1, err := net.ListenTCP("tcp", tcpAd1)
	checkErr(err)

	go func() {
		for {
			c, err := listener.Accept()
			checkErr(err)
			go handle(c)
		}

	}()

	go func() {
		for {
			c1, err := listener1.Accept()
			checkErr(err)
			go handle1(c1)
		}

	}()

	//a cmd to control our serve
	var cmd string
	for {
		fmt.Scanln(&cmd)
		if cmd == "q" {
			os.Exit(0)
		}
	}

}

//handle functuon
func handle(c net.Conn) {
	defer c.Close()

	buf := make([]byte, 1024)
	_, err := c.Read(buf)
	checkErr(err)

	cpack := getCPack(string(buf))

	var sp SPack
	switch cpack.Types {
	case "register":
		//注册服务
		sp = register(cpack)

	case "login":
		sp = login(cpack)

	case "getgroups":
		sp = getGroups(cpack)
	case "addEvent":
		sp = addEvent(cpack)

	case "getfriends":
		sp = getFriends(cpack)
	//no type or type is error
	default:

		//404 找不到服务
		mdata := make(map[string]interface{})
		mdata["tips"] = "未知的服务"
		sp = esp("unknown", 404, -1, mdata)
	}

	re, _ := url.QueryUnescape(setSPack(sp))

	fmt.Println(re)
	c.Write([]byte(re))

}

func handle1(c net.Conn) {

	buf := make([]byte, 2048)
	_, err := c.Read(buf)
	checkErr(err)

	cp := getCPack(string(buf))

	switch cp.Types {
	case "getmsg":
		//使用cp,启动长轮询，需要转交conn
		getMsg(c, cp)

	//不获取信息就发送信息
	default:
		defer c.Close()
		cmp := getcmsgPack(string(buf))
		sp := sendMsg(cmp)
		re, _ := url.QueryUnescape(setSPack(sp))
		c.Write([]byte(re))
	}
}
