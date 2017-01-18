package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
)

//这是一个go 写的hosts的更新程序

var hostsPath = "C:/Windows/System32/drivers/etc/hosts"

func main() {
	//获取最新的hosts文件
	resp, err := http.Get("https://raw.githubusercontent.com/racaljk/hosts/master/hosts")
	if err != nil {
		panic(err)
	}
	//预关闭资源
	defer resp.Body.Close()
	//读取所有的字节数
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	//写入到hosts文件
	openHosts(b)
}
func openHosts(hb []byte) {
	b, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		panic(err)
	}
	gb := []byte("#googlehosts#")
	index := bytes.Index(b, gb)
	var ob bytes.Buffer
	ob.Write(b[0 : index+len(gb)])
	ob.WriteString("\n")
	ob.Write(hb)
	ioutil.WriteFile(hostsPath, ob.Bytes(), os.ModePerm)
}
