package main

//业逻辑处理包
//注册事件
import (
	"strconv"
	//"fmt"
)

func register(cp CPack) SPack {
	var sp SPack

	sp.Uid = -1
	//data封包实体化
	m := cp.Data

	mdata := make(map[string]interface{})

	if isempty(m["user"]) {
		mdata["tips"] = "账号不可以留空"
		//101已存在 102验证失败 103信息不完整
		sp.Code = 103
	} else if isempty(m["pwd"]) {
		mdata["tips"] = "密码不可以留空"
		//101已存在 102验证失败 103信息不完整
		sp.Code = 103
	} else if isempty(m["username"]) {
		mdata["tips"] = "昵称不可以留空"
		//101已存在 102验证失败 103信息不完整
		sp.Code = 103
	} else {
		stmt, err := db.Prepare("insert into user(u,uname,pwd) values(?,?,?)")
		checkErr(err)

		res, err := stmt.Exec(m["user"], m["username"], m["pwd"])
		checkErr(err)

		if err == nil {
			id, err := res.LastInsertId()
			checkErr(err)

			//201成功
			sp.Code = 200
			sp.Uid = int(id)

			mdata["username"] = m["username"]
			mdata["user"] = m["user"]

			//注册成功
			userInit(sp.Uid)

		} else {
			//101已存在 102验证失败
			sp.Code = 101
			mdata["tips"] = "当前账号已存在,无法注册"
		}
	}
	return esp(cp.Types, sp.Code, sp.Uid, mdata)
}

//登录事件
func login(cp CPack) SPack {
	//data封包实体化
	m := cp.Data
	mdata := make(map[string]interface{})
	var rpwd string
	var uname string
	var u string
	var uid int

	//修改登录方式
	if !isempty(m["user"]) {
		user, ok := m["user"].(string)
		if ok {
			uid = getUidByUser(user)
		}
	} else {
		uid = cp.Uid
	}

	//get user
	u = getUserByUid(uid)

	//get rpwd
	err := db.QueryRow("select pwd from user where uid=" + strconv.Itoa(uid)).Scan(&rpwd)
	checkErr(err)

	if err == nil {
		if rpwd == m["pwd"] {

			err := db.QueryRow("select uname from user where uid=" + strconv.Itoa(uid)).Scan(&uname)
			checkErr(err)

			//生成skey
			skey := getSkey(u, rpwd)

			stmt, err := db.Prepare("update user set skey=? where uid=" + strconv.Itoa(uid))
			checkErr(err)

			_, err = stmt.Exec(skey)
			checkErr(err)

			mdata["tips"] = "登录成功"
			mdata["user"] = u
			mdata["username"] = uname
			mdata["skey"] = skey
			Skey[uid] = skey
			return esp(cp.Types, 200, uid, mdata)
		}
	}
	mdata["tips"] = "登录失败,用户名或密码错误"
	return esp(cp.Types, 102, uid, mdata)
}

//信息设置
func setInfo(cp CPack) {

}

//获取群列表
func getGroups(cp CPack) SPack {

	mdata := make(map[string]interface{})
	skey := cp.Skey

	if verSkey(cp.Uid, skey) {
		var groups string
		err := db.QueryRow("select groups from user where uid=" + strconv.Itoa(cp.Uid)).Scan(&groups)
		checkErr(err)

		w := "(" + groups + ")"
		rows, err := db.Query("select gid,gname,des from groups where gid in " + w)

		checkErr(err)
		if err == nil {

			type Groups struct {
				Gid   int
				Gname string
				Des   string
			}

			var gs []Groups
			for rows.Next() {
				var g Groups
				err = rows.Scan(&g.Gid, &g.Gname, &g.Des)
				checkErr(err)
				gs = append(gs, g)
			}

			mdata["groups"] = gs

			return esp(cp.Types, 200, cp.Uid, mdata)
		}
	} else {
		mdata["tips"] = "您的skey已失效,请重新登录"
		return esp(cp.Types, 102, cp.Uid, mdata)
	}
	mdata["tips"] = "当前账号未找到相关信息"
	return esp(cp.Types, 102, cp.Uid, mdata)
}

//获取好友列表
func getFriends(cp CPack) SPack {

	mdata := make(map[string]interface{})
	skey := cp.Skey

	if verSkey(cp.Uid, skey) {
		var friends string
		err := db.QueryRow("select friends from user where uid=" + strconv.Itoa(cp.Uid)).Scan(&friends)
		checkErr(err)

		w := "(" + friends + ")"
		rows, err := db.Query("select uid,uname from user where uid in " + w)

		checkErr(err)
		if err == nil {

			type Friends struct {
				Uid   int
				Uname string
			}

			var fs []Friends
			for rows.Next() {
				var f Friends
				err = rows.Scan(&f.Uid, &f.Uname)
				checkErr(err)
				fs = append(fs, f)
			}

			mdata["friends"] = fs

			return esp(cp.Types, 200, cp.Uid, mdata)
		}
	} else {
		mdata["tips"] = "您的skey已失效,请重新登录"
		return esp(cp.Types, 102, cp.Uid, mdata)
	}
	mdata["tips"] = "当前账号未找到相关信息"
	return esp(cp.Types, 102, cp.Uid, mdata)
}

//添加事件
func addEvent(cp CPack) SPack {
	mdata := make(map[string]interface{})
	m := cp.Data
	if verSkey(cp.Uid, cp.Skey) {
		tuid := getType(m["target"])
		ttype := getType(m["type"])
		if tuid == "" {
			mdata["tips"] = "目标对象不能为空"
			return esp(cp.Types, 103, cp.Uid, mdata)
		}
		if ttype != "groups" && ttype != "friends" {
			mdata["tips"] = "未知的添加对象"
			return esp(cp.Types, 103, cp.Uid, mdata)
		}
		if tuid == strconv.Itoa(cp.Uid) {
			mdata["tips"] = "不允许添加自己"
			return esp(cp.Types, 103, cp.Uid, mdata)
		}

		ituid, err := strconv.Atoi(tuid)
		checkErr(err)
		var name string
		var user string
		if ttype == "groups" {
			name = getGroupData("gname", ituid)
		}
		if ttype == "friends" {
			name = getUserData("uname", ituid)
			user = getUserData("user", ituid)
		}

		if name != "" {

			data := getUserData(ttype, cp.Uid)
			if verSetExist(sString(data, ","), tuid) {
				mdata["tips"] = "当前添加对象已存在"
				return esp(cp.Types, 103, cp.Uid, mdata)
			}

			stmt, err := db.Prepare("update user set " + ttype + "=? where uid=" + strconv.Itoa(cp.Uid))
			checkErr(err)
			_, err = stmt.Exec(setAdd(data, tuid))
			checkErr(err)

			if ttype == "friends" {
				data1 := getUserData("friends", ituid)
				stmt, err := db.Prepare("update user set friends=? where uid=" + tuid)
				checkErr(err)
				_, err = stmt.Exec(setAdd(data1, strconv.Itoa(cp.Uid)))
				checkErr(err)
			}

			if ttype == "groups" {
				data2 := getGroupData("members", ituid)
				stmt, err := db.Prepare("update groups set members=? where gid=" + tuid)
				checkErr(err)
				_, err = stmt.Exec(setAdd(data2, strconv.Itoa(cp.Uid)))
				checkErr(err)
			}

			mdata["tips"] = "添加成功"
			mdata["target"] = tuid
			mdata["ttype"] = ttype
			mdata["tname"] = name
			if user != "" {
				mdata["user"] = user
			}
			return esp(cp.Types, 200, cp.Uid, mdata)
		}
		mdata["tips"] = "您的添加对象不存在"
		return esp(cp.Types, 103, cp.Uid, mdata)
	} else {
		mdata["tips"] = "您的skey已失效,请重新登录"
		return esp(cp.Types, 102, cp.Uid, mdata)
	}
	mdata["tips"] = "您的skey已失效,请重新登录"
	return esp(cp.Types, 102, cp.Uid, mdata)
}
