package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// 配置文件
type Config struct {
	Host  string  `json:"host"`  // 监听地址
	User  string  `json:"user"`  // 用户信息
	Path  string  `json:"path"`  // 保存路径
	Sleep int     `json:"sleep"` // 重连间隔
	Video []Video `json:"video"` // 视频录制配置
}

// 视频录制配置
type Video struct {
	WsHost   string `json:"wsHost"`   // 连接地址
	ParamMsg string `json:"paramMsg"` // 连接参数
	Name     string `json:"name"`     // 设备名称
	Size     int    `json:"size"`     // 截断大小
	Count    int    `json:"count"`    // 存储数量
}

// 获取配置
func GetConfig() Config {
	var config Config
	filePath := "config.json"
	data, err := os.ReadFile(filePath)
	if err != nil {
		FmtPrint("配置文件不存在", err)
		os.Exit(0)
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		FmtPrint("读取配置文件出错", err)
		os.Exit(0)
	}
	return config
}

// 删除文件夹下的旧文件
func DeleteOldFiles(dirPath string, filesToKeep int) {
	//读取文件
	fileInfos := []fs.FileInfo{}
	filePaths, _ := filepath.Glob(filepath.Join(dirPath, "*"))
	for _, filePath := range filePaths {
		info, _ := os.Stat(filePath)
		if info.Mode().IsRegular() {
			fileInfos = append(fileInfos, info)
		}
	}
	//检查文件数量
	if len(fileInfos) <= filesToKeep {
		return
	}
	//按时间排序
	sort.Slice(fileInfos, func(i, j int) bool {
		return fileInfos[i].ModTime().After(fileInfos[j].ModTime())
	})
	//删除最旧的文件
	for i := filesToKeep; i < len(fileInfos); i++ {
		oldFile := filepath.Join(dirPath, fileInfos[i].Name())
		_ = os.Remove(oldFile)
	}
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
