package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// 配置文件
type Config struct {
	Host         string  `json:"host"`         // 监听地址
	User         string  `json:"user"`         // 用户信息
	Path         string  `json:"path"`         // 保存路径
	Sleep        int     `json:"sleep"`        // 重连间隔
	OnvifPort    int     `json:"onvifPort"`    // ONVIF服务端口
	OnvifAuth    bool    `json:"onvifAuth"`    // ONVIF认证开关
	Video        []Video `json:"video"`        // 视频录制配置
}

// 视频录制配置
type Video struct {
	WsHost      string `json:"wsHost"`      // 连接地址
	ParamMsg    string `json:"paramMsg"`    // 连接参数
	Name        string `json:"name"`        // 设备名称
	Size        int    `json:"size"`        // 截断大小
	Count       int    `json:"count"`       // 保留天数
	StreamOnly  bool   `json:"streamOnly"`  // 仅流模式，不保存文件
}

//go:embed config.json
var defaultConfig []byte // 默认配置

// 获取配置
func GetConfig() Config {
	var config Config
	filePath := "config.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		//不存在时，生成一个配置文件
		err = os.WriteFile(filePath, defaultConfig, 0666)
		if err != nil {
			FmtPrint("配置文件创建失败", err)
			os.Exit(0)
		}
		FmtPrint("已生成默认配置文件，请更改配置文件后再启动程序！")
		//等待用户输入
		FmtPrint("按回车键退出程序...")
		var input string
		fmt.Scanln(&input)
		//退出程序
		os.Exit(0)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		FmtPrint("读取配置文件出错", err)
		os.Exit(0)
	}
	return config
}

// 定义内置的打印语句
func FmtPrint(data ...any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	processedData, hasFormat, formatStr := processArgs(data...)
	// 输出
	if len(data) == 1 {
		fmt.Printf("%s: %v\n", date, processedData[0])
	} else if hasFormat {
		fmt.Printf("%s: "+formatStr+"\n", append([]any{date}, processedData[1:]...)...)
	} else {
		fmt.Printf("%s: %v\n", date, processedData)
	}
}

// 写日志
func LogWrite(data ...any) {
	date := time.Now().Format("2006-01-02 15:04:05")
	processedData, hasFormat, formatStr := processArgs(data...)
	// 检查日志文件夹
	dirPath := "logs"
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		err = os.MkdirAll(dirPath, 0777)
		if err != nil {
			FmtPrint("日志文件夹创建失败", err)
			return
		}
	}
	// 打开日志文件
	fileName := time.Now().Format("2006-01-02")
	filePath := dirPath + "/" + fileName + ".log"
	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		FmtPrint("日志文件创建失败", err)
		return
	}
	defer file.Close()
	// 写入日志
	if len(data) == 1 {
		file.WriteString(date + ": " + fmt.Sprintf("%v", processedData[0]) + "\n")
	} else if hasFormat {
		file.WriteString(date + ": " + fmt.Sprintf(formatStr, processedData[1:]...) + "\n")
	} else {
		file.WriteString(date + ": " + fmt.Sprintf("%v", processedData) + "\n")
	}
}

// 处理参数列表
func processArgs(data ...any) ([]any, bool, string) {
	processedData := make([]any, len(data))
	for i, item := range data {
		if bytes, ok := item.([]byte); ok {
			processedData[i] = string(bytes)
		} else {
			processedData[i] = item
		}
	}
	// 检查是否是格式化字符串
	hasFormat := false
	formatStr := "%v"
	if len(data) > 1 {
		if format, ok := processedData[0].(string); ok {
			hasFormat = true
			formatStr = format
		}
	}
	return processedData, hasFormat, formatStr
}
