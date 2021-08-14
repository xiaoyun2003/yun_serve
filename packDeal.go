package main

import (
	"encoding/json"
)

type CPack struct {
	Types  string
	Data   map[string]interface{}
	Osinfo map[string]interface{}
	Skey   string
	Uid    int
}

type SPack struct {
	Types     string
	Code      int
	Data      map[string]interface{}
	Serveinfo map[string]interface{}
	Uid       int
}

type cmsgPack struct {
	Types    string
	Subtypes string
	Sname    string
	Msg      string
	Skey     string
	Uid      int
	Receiver int
}

type smsgPack struct {
	Types    string
	Subtypes string
	Msgid    int
	Mid      int
	Sname    string
	Msg      string
	Sender   int
	Receiver int
}

func getCPack(data string) CPack {
	data = stringDeal(data)
	var a CPack
	err := json.Unmarshal([]byte(data), &a)
	checkErr(err)
	a.Data = Sdata(a.Data)
	return a
}

func setSPack(a SPack) string {
	data, err := json.Marshal(a)
	checkErr(err)
	return string(data)
}

func getcmsgPack(data string) cmsgPack {
	data = stringDeal(data)
	var a cmsgPack
	err := json.Unmarshal([]byte(data), &a)
	checkErr(err)
	return a
}

func setsmsgPack(a smsgPack) string {
	data, err := json.Marshal(a)
	checkErr(err)
	return string(data)
}
