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
	Host   string  `json:"host"`
	Sleep  int     `jsson:"sleep"`
	FFmpeg FFmpeg  `json:"ffmpeg"`
	Video  []Video `json:"video"`
}

// 视频转码配置
type FFmpeg struct {
	Exec string `json:"exec"`
	Type string `json:"type"`
	Gpu  string `json:"gpu"`
}

// 视频录像机配置
type Video struct {
	WsHost   string `json:"wsHost"`
	ParamMsg string `json:"paramMsg"`
	Name     string `json:"name"`
	Size     int    `json:"size"`
	Count    int    `json:"count"`
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
	if len(data) == 1 {
		fmt.Println(date+": ", data[0])
	} else {
		fmt.Println(date+": ", data)
	}
}
