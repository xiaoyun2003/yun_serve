package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

//init serve,create datatables
func serveInit(dbFile string) {
	fmt.Println("开始初始化服务器.....")
	db, err := sql.Open("sqlite3", dbFile)
	checkErr(err)

	//设置详情
	db.Exec("create table data(key text,value integer);")
	stmt, err := db.Prepare("insert into data(key,value) values(?,?)")
	checkErr(err)
	_, err = stmt.Exec("mid", 0)
	checkErr(err)

	db.Exec("create table user(uid integer primary key AUTOINCREMENT unique, u text unique, uname text, pwd text, skey text,  friends text, groups text, system text,settings text, createtime timestamp);")

	db.Exec("create table groups(gid integer primary key AUTOINCREMENT unique, gname text, des text, members text, admin text, settings text, createtime timestamp);")

	//创建初始化账号
	stmt, err = db.Prepare("insert into groups(gname,des,admin,createtime) values(?,?,?,?)")
	checkErr(err)

	ct := time.Now()
	_, err = stmt.Exec("云聊官方聊天室", "From 云聊官方", "1", ct)
	checkErr(err)

	db.Exec("create table msg(mid integer primary key AUTOINCREMENT unique, msgid interger,msg text, type text, subtype text, sender integer, sname text, receiver interge, createtime timestamp);")

	db.Exec("create unique index mid on msg (mid)")
	db.Close()
}

//账号注册初始化操作,对所有账号具有普遍性
func userInit(uid int) {

	stmt, err := db.Prepare("update user set system=? ,createtime=? ,groups=? ,friends=? where uid=" + strconv.Itoa(uid))
	checkErr(err)

	ct := time.Now()
	_, err = stmt.Exec("100", ct, "1", "1")
	checkErr(err)

	data2 := getGroupData("members", 1)
	stmt, err = db.Prepare("update groups set members=? where gid=1")
	checkErr(err)
	_, err = stmt.Exec(setAdd(data2, strconv.Itoa(uid)))
	checkErr(err)

	if uid != 1 {

		data1 := getUserData("friends", 1)
		stmt, err = db.Prepare("update user set friends=? where uid=1")
		checkErr(err)
		_, err = stmt.Exec(setAdd(data1, strconv.Itoa(uid)))
		checkErr(err)
	}
}
