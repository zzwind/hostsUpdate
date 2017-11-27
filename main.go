package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

//这是一个go 写的hosts的更新程序
var hostsPath = "C:/Windows/System32/drivers/etc/hosts"
var localHostsPath = "./hostsdb"
var logger *log.Logger

//验证信息
var mac, initDate, license string = "C8:60:00:C9:8A:95", "2027-03-24", "pay"

func init() {
	//初始化标签
	flagGooglehosts()
	logFile, err := os.OpenFile("update.log", os.O_APPEND|os.O_CREATE, 777)
	if err != nil {
		log.Fatalln("faild to create upupdate.log")
	}
	logger = log.New(logFile, "", log.LstdFlags)

}
func main() {
	licenseCheck()
	//os.Exit(0)

	//检测是否能连接谷歌，能就退出程序
	if envCheck() {
		logger.Println("环境检测正常，无需更新")
		os.Exit(0)
	}
	logger.Println("环境检测失败，从远程获取hosts中...")
	//获取远程hosts，获取失败载入自带的hosts
	hostsByte, err := getRemoteHosts()
	if err != nil {
		logger.Println("远程获取hosts失败,从本地获取...", err)
		hostsByte, err = loadLocalFile()
		if err != nil {
			logger.Println("载入本地hosts失败", err)
			os.Exit(0)
		}
	}
	err = writeHosts(hostsByte)
	if err != nil {
		logger.Println("hosts写入失败:", err)
	} else {
		logger.Println("hosts写入完成")
	}
	flushdns()
	logger.Println("hosts刷新完成，程序退出")
}
func loadLocalFile() ([]byte, error) {
	b, err := ioutil.ReadFile(localHostsPath)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//获取远程hosts文件
func getRemoteHosts() ([]byte, error) {

	//获取最新的hosts文件
	logger.Println("开始更新")
	//这个是老的hosts 地址
	//resp, err := http.Get("https://raw.githubusercontent.com/racaljk/hosts/master/hosts")
	resp, err := http.Get("https://raw.githubusercontent.com/googlehosts/hosts/master/hosts-files/hosts")

	if err != nil {
		return nil, err
	}
	//预关闭资源
	defer resp.Body.Close()
	//读取所有的字节数
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

//本地环境检测
func envCheck() bool {
	r, err := http.Get("https://www.google.com.hk")
	if err != nil {
		logger.Println(err)
		return false
	}
	return r.StatusCode == 200
}

//写入hosts
func writeHosts(hb []byte) error {
	b, err := ioutil.ReadFile(hostsPath)
	if err != nil {
		logger.Println(err)
		return err
	}
	gb := []byte("#googlehosts#")
	index := bytes.Index(b, gb)
	var ob bytes.Buffer
	ob.Write(b[0 : index+len(gb)])
	ob.WriteString("\n")
	ob.Write(hb)
	return ioutil.WriteFile(hostsPath, ob.Bytes(), os.ModePerm)
}

//写入hosts flag
func flagGooglehosts() {

	if _, err := os.Stat("./update.log"); os.IsNotExist(err) {
		f, _ := os.OpenFile(hostsPath, os.O_APPEND, 777)
		f.Write([]byte("#googlehosts#"))
		f.Close()
	}
}

//刷新dns
func flushdns() {
	cmd := exec.Command("ipconfig", "/flushdns")
	_, err := cmd.Output()
	if err != nil {
		logger.Println(err)
	}
}

//许可检测
func licenseCheck() {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}

	if len(interfaces) > 0 {

		hardWareAddr := interfaces[0].HardwareAddr
		stime, err := time.Parse("2006-01-02", initDate)
		if err != nil {
			logger.Println("初始时间格式化失败")
			os.Exit(0)
		}
		if license == "pay" {
			if stime.AddDate(1, 0, 0).Before(time.Now()) {
				logger.Println("许可时间过期")
				os.Exit(0)
			}
		} else if license == "trial" {
			if stime.AddDate(0, 0, 3).Before(time.Now()) {
				logger.Println("试用时间过期")
				os.Exit(0)
			}
		}

		if mac != strings.ToUpper(hardWareAddr.String()) {
			logger.Println("mac不匹配，非法许可", mac, hardWareAddr.String())
			os.Exit(0)
		}

	} else {
		logger.Println("mac地址获取失败，无法进行许可验证")
	}
}
