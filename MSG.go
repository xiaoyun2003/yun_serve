package main

//消息处理包
import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"

	"strconv"
	"time"
)

//处理来自客服端的包，返回数据给客服端
//使用长轮询方式
func getMsg(c net.Conn, cp CPack) int {

	if verSkey(cp.Uid, cp.Skey) {
		m := cp.Data
		var mid string
		if !isempty(m["mid"]) {
			mid = getType(m["mid"])
		}
		if mid == "" {
			mid = "0"
		}

	LOOP:

		var friends string
		var groups string
		err := db.QueryRow("select friends from user where uid=" + strconv.Itoa(cp.Uid)).Scan(&friends)
		checkErr(err)
		err = db.QueryRow("select groups from user where uid=" + strconv.Itoa(cp.Uid)).Scan(&groups)
		checkErr(err)

		//获取消息
		sql := "select type,subtype,msgid,mid,msg,sender,sname,receiver from msg where (receiver in (" + groups + ") or receiver=" + strconv.Itoa(cp.Uid) + ") and mid>" + mid

		rows, err := db.Query(sql)
		msgFlag := 0

		checkErr(err)
		if err == nil {
			var smps []smsgPack

			for rows.Next() {
				//有消息

				msgFlag = 1

				var smp smsgPack
				err := rows.Scan(&smp.Types, &smp.Subtypes, &smp.Msgid, &smp.Mid, &smp.Msg, &smp.Sender, &smp.Sname, &smp.Receiver)
				checkErr(err)
				smps = append(smps, smp)
			}

			//有数据就发给客服端
			if msgFlag == 1 {

				defer c.Close()
				fmt.Println("有消息")
				mdata := make(map[string]interface{})
				mdata["msg"] = smps

				sp := esp(cp.Types, 200, cp.Uid, mdata)
				re, _ := url.QueryUnescape(setSPack(sp))
				c.Write([]byte(re))
				//此时有消息应该退出阻塞,
				return 0

			}

		}
		time.Sleep(time.Second * 1)
		goto LOOP
	} else {
		mdata := make(map[string]interface{})
		mdata["tips"] = "您的skey已失效,请重新登录"
		sp := esp(cp.Types, 102, cp.Uid, mdata)
		re, _ := url.QueryUnescape(setSPack(sp))
		c.Write([]byte(re))
	}
	return 0
}

//处理来自客服端的包
func sendMsg(cmp cmsgPack) SPack {
	if verSkey(cmp.Uid, cmp.Skey) {
		rec := cmp.Receiver
		se := cmp.Uid

		//编码msg
		msg := base64.StdEncoding.EncodeToString([]byte(cmp.Msg))
		//获取用户名
		sname := getUnameByUid(se)
		var vm bool
		if cmp.Types == "groups" {
			vm, _ = verMsgR(2, se, rec)
		}
		if cmp.Types == "friends" {
			vm, _ = verMsgR(1, se, rec)
		}
		if vm {
			var msgid int
			err := db.QueryRow("select count(*) from msg where type='" + cmp.Types + "' and " + "receiver=" + strconv.Itoa(rec)).Scan(&msgid)
			checkErr(err)
			msgid = msgid + 1
			stmt, err := db.Prepare("insert into msg(msgid,msg,type,subtype,sender,sname,receiver,createtime) values(?,?,?,?,?,?,?,?)")
			checkErr(err)
			ct := time.Now()
			res, err := stmt.Exec(msgid, msg, cmp.Types, "text", se, sname, rec, ct)
			checkErr(err)
			if err == nil {
				mid, err := res.LastInsertId()
				checkErr(err)

				mdata := make(map[string]interface{})
				mdata["tips"] = "发送成功"
				mdata["msgid"] = msgid
				mdata["mid"] = mid
				mdata["sender"] = se
				mdata["sname"] = sname
				mdata["msg"] = msg
				mdata["receiver"] = rec
				return esp(cmp.Types, 200, cmp.Uid, mdata)
			}
			mdata := make(map[string]interface{})
			mdata["tips"] = "发送失败"
			return esp(cmp.Types, 101, cmp.Uid, mdata)
		} else {
			mdata := make(map[string]interface{})
			mdata["tips"] = "不合法的请求"
			return esp(cmp.Types, 102, cmp.Uid, mdata)
		}
	} else {
		mdata := make(map[string]interface{})
		mdata["tips"] = "您的skey已失效,请重新登录"
		return esp(cmp.Types, 102, cmp.Uid, mdata)
	}
	mdata := make(map[string]interface{})
	mdata["tips"] = "未知的错误"
	return esp(cmp.Types, 101, cmp.Uid, mdata)
}
