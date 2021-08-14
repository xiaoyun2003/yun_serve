package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//判断文件是否存在
func isFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}

//错误判断
func checkErr(err error) {
	funcName, file, line, _ := runtime.Caller(0)
	if err != nil {
		fmt.Println(file, funcName, line)

		fmt.Printf("%c[1;40;31m%s%s%c[0m\n", 0x1B, "[ERROR]", err.Error(), 0x1B)
	}
}

//封装返回
func esp(types string, code int, uid int, data map[string]interface{}) SPack {
	var sp SPack
	sp.Code = code
	sp.Data = data
	sp.Types = types
	m := make(map[string]interface{})
	m["tips"] = "From 小云云聊服务器"
	sp.Serveinfo = m
	sp.Uid = uid
	return sp
}

//skey生成算法
func getSkey(u string, p string) string {

	time := time.Now().Unix()
	skey := md5.Sum([]byte(u + strconv.FormatInt(time, 10) + p))

	sk := fmt.Sprintf("%x", skey)
	return sk

}

//get uid by user
func getUidByUser(u string) int {
	var uid int
	err := db.QueryRow("select uid from user where u='" + u + "'").Scan(&uid)
	checkErr(err)
	if err == nil {
		return uid
	}
	return -1
}

func stringDeal(s string) string {
	str := bytes.Trim([]byte(s), "\x00")
	return string(str)
}

//过滤数据
func Sdata(s map[string]interface{}) map[string]interface{} {
	for k, v := range s {
		str, ok := v.(string)
		if ok {
			s[k] = url.QueryEscape(str)
		}
	}
	return s
}

func sString(data string, key string) []string {
	if strings.Contains(data, key) {
		return strings.Split(data, key)
	}
	var d []string
	d = append(d, data)
	return d
}

//获取skey
func getSkeyByUid(uid int) string {
	var skey string
	//首先查缓存表
	fmt.Println("缓存表", uid)
	if Skey[uid] != "" {
		fmt.Println(Skey[uid])
		return Skey[uid]
	}
	err := db.QueryRow("select skey from user where uid=" + strconv.Itoa(uid)).Scan(&skey)
	checkErr(err)
	Skey[uid] = skey
	return skey
}

//验证skey
func verSkey(uid int, skey string) bool {
	rskey := getSkeyByUid(uid)
	if rskey == skey && rskey != "" && skey != "" {
		return true
	}
	return false
}

//获取用户
func getUserByUid(uid int) string {
	var user string
	err := db.QueryRow("select u from user where uid=" + strconv.Itoa(uid)).Scan(&user)
	checkErr(err)
	if uid != 0 && user != "" && err == nil {
		return user
	}
	return ""
}

//获取用户名
func getUnameByUid(uid int) string {
	var uname string
	err := db.QueryRow("select uname from user where uid=" + strconv.Itoa(uid)).Scan(&uname)
	checkErr(err)
	if uid != 0 && uname != "" && err == nil {
		return uname
	}
	return ""
}

//验证消息关系
//一个很重要的函数,该函数用来验证发送者和接受者的关系是否合法,需要指明验证类型 1 好友 2 群聊
func verMsgR(types int, sender int, receiver int) (bool, string) {
	if sender == 0 || receiver == 0 {
		return false, ""
	}
	if types == 1 {
		var friends string
		err := db.QueryRow("select friends from user where uid=" + strconv.Itoa(receiver)).Scan(&friends)
		checkErr(err)
		ss := sString(friends, ",")
		for _, v := range ss {
			if v == strconv.Itoa(sender) {
				return true, friends
			}
		}
		return false, friends
	}
	if types == 2 {
		var members string
		fmt.Println("成员", members)
		err := db.QueryRow("select members from groups where gid=" + strconv.Itoa(receiver)).Scan(&members)
		checkErr(err)

		ss := sString(members, ",")
		for _, v := range ss {
			if v == strconv.Itoa(sender) {
				return true, members
			}
		}
		return false, members
	}
	return false, ""
}

//判空
func isempty(data interface{}) bool {
	str, ok := data.(string)
	if data == nil {
		return true
	}
	if str == "" && ok {
		return true
	}
	return false
}

//类型转换
func getType(i interface{}) string {
	switch i.(type) {
	case int:
		return strconv.Itoa(i.(int))
	case string:
		return i.(string)
	case int64:
		return strconv.FormatInt(i.(int64), 10)
	case bool:
		return strconv.FormatBool(i.(bool))
	case float64:
		return strconv.Itoa(int(i.(float64)))
	}
	return ""
}

//
func getGroupData(key string, gid int) string {
	var data interface{}
	err := db.QueryRow("select " + key + " from groups where gid=" + strconv.Itoa(gid)).Scan(&data)
	checkErr(err)
	return getType(data)
}

func getUserData(key string, uid int) string {
	var data interface{}
	err := db.QueryRow("select " + key + " from user where uid=" + strconv.Itoa(uid)).Scan(&data)
	checkErr(err)
	return getType(data)
}

func setAdd(set string, add string) string {
	if set == "" && add == "" {
		return ""
	}
	if add == "" {
		return set
	}
	if set == "" {
		return add
	}
	return set + "," + add
}

//验证用户或群组是否存在
func verExist(types string, id int) bool {
	var u string
	if types == "groups" {
		u = "gid"
	} else if types == "user" {
		u = "uid"
	} else {
		return false
	}
	var data int
	err := db.QueryRow("select " + u + " from " + types + " where " + u + "=" + strconv.Itoa(id)).Scan(&data)
	if err == nil && data != 0 {
		return true
	}
	return false
}

//判断数组当中某个元素是否存在
func verSetExist(data []string, key string) bool {
	for _, v := range data {
		if strings.Contains(v, key) {
			return true
		}
	}
	return false
}

func setUserInfo(uid int, key string, value string) {
	var settings string
	err := db.QueryRow("select settings from user where uid=" + strconv.Itoa(uid)).Scan(&settings)
	checkErr(err)
	m := make(map[string]string)
	err = json.Unmarshal([]byte(settings), &m)
	m[key] = value
	json, err := json.Marshal(m)
	stmt, err := db.Prepare("update user set settings=? where uid=" + strconv.Itoa(uid))
	checkErr(err)
	_, err = stmt.Exec(json)
	checkErr(err)
}

func getUserInfo(uid int, key string) string {
	var settings string
	err := db.QueryRow("select settings from user where uid=" + strconv.Itoa(uid)).Scan(&settings)
	checkErr(err)
	m := make(map[string]string)
	err = json.Unmarshal([]byte(settings), &m)
	if key == "" {
		return settings
	}
	return m[key]
}

func setGroupInfo(gid int, key string, value string) {
	var settings string
	err := db.QueryRow("select settings from user where uid=" + strconv.Itoa(gid)).Scan(&settings)
	checkErr(err)
	m := make(map[string]string)
	err = json.Unmarshal([]byte(settings), &m)
	m[key] = value
	json, err := json.Marshal(m)
	stmt, err := db.Prepare("update groups set settings=? where uid=" + strconv.Itoa(gid))
	checkErr(err)
	_, err = stmt.Exec(json)
	checkErr(err)
}

func getGroupInfo(gid int, key string) string {
	var settings string
	err := db.QueryRow("select settings from groups where uid=" + strconv.Itoa(gid)).Scan(&settings)
	checkErr(err)
	m := make(map[string]string)
	err = json.Unmarshal([]byte(settings), &m)
	if key == "" {
		return settings
	}
	return m[key]
}
