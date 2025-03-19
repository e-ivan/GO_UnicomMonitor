package main

import (
	"crypto/tls"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 录制
func GoRecording(config *Config, video *Video) {
	//临时变量
	tempPath := config.Path + "/" + video.Name
	//断开后重连
	for {
		//连接服务器传输数据
		bytes := linkServer(video)
		//检查数据
		if len(bytes) == 0 {
			FmtPrint(video.Name + "设备连接失败，稍后自动重连(" + strconv.Itoa(config.Sleep) + ")")
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
			continue
		}
		//文件名称
		fileName := getFileName(tempPath) + ".hevc"
		//保存文件
		saveFile(fileName, &bytes)
		//录制完成
		FmtPrint("录制完成：" + fileName)
	}
}

// 连接服务器
func linkServer(video *Video) []byte {
	bytes := []byte{}
	uri := url.URL{
		Scheme: "wss",
		Host:   video.WsHost,
		Path:   "/h5player/live",
	}
	//跳过证书验证
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	//发起连接
	conn, _, err := dialer.Dial(uri.String(), nil)
	if err != nil {
		FmtPrint("无法连接到服务器", err)
		return bytes
	}
	defer conn.Close()
	//发送消息
	message := "_paramStr_=" + video.ParamMsg
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		FmtPrint("发送消息失败：", err)
		return bytes
	}
	//接收消息
	for {
		_, response, err := conn.ReadMessage()
		if err != nil {
			FmtPrint("接收消息失败：", err)
			return bytes
		}
		//检查特定条件
		if len(response) > 1 {
			//打印数据的长度
			//FmtPrint("数据长度：", len(bytes))
			//拼接数据
			bytes = append(bytes, response[:]...)
			//结束条件
			if len(bytes) > 1024*1024*video.Size {
				//结束
				return bytes
			}
		}
	}
}

// 获取文件名称
func getFileName(dirPath string) string {
	//检查文件夹是否存在
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		//文件夹不存在，创建它
		err := os.MkdirAll(dirPath, 0755)
		if err != nil {
			FmtPrint("创建文件夹失败：", err)
			os.Exit(0)
		}
		FmtPrint("创建文件夹：" + dirPath)
	}
	//文件名称
	fileName := time.Now().Format("20060102_150405")
	tempPathh := dirPath + "/" + fileName
	return tempPathh
}

// 保存文件
func saveFile(fileName string, bytes *[]byte) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		FmtPrint("保存文件失败: ", err)
		os.Exit(0)
	}
	defer file.Close()
	file.Write(*bytes)
}
