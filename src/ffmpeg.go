package main

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// FFmpeg 转码
func GoFFmpeg(ffmpeg FFmpeg, video Video, tempPath string) {
	//整理并合并文件
	filePath := organizeFiles(tempPath)
	//转码文件
	transcodFiles(ffmpeg, filePath, video.Name)
}

// 整理文件
func organizeFiles(tempPath string) string {
	//读取文件
	filePaths, _ := filepath.Glob(filepath.Join(tempPath, "*.flv"))
	//取出第一个文件名称
	newFileName := strings.Replace(filePaths[0], ".flv", ".bin", -1)
	//合并文件
	mergeFiles(filePaths, newFileName)
	//删除旧的
	deleteFiles(filePaths)
	//返回合并的文件
	return newFileName
}

// 合并文件
func mergeFiles(files []string, newName string) {
	//创建一个目标文件来保存拼接结果
	outputFile, _ := os.Create(newName)
	defer outputFile.Close()
	//遍历文件列表，依次读取并写入目标文件
	for _, file := range files {
		inputFile, _ := os.Open(file)
		defer inputFile.Close()
		//将源文件内容复制到目标文件
		io.Copy(outputFile, inputFile)
	}
}

// 转码文件
func transcodFiles(ffmpeg FFmpeg, filePath string, path string) {
	//FmtPrint("转码文件：" + filePath)
	//默认复制文件
	fileName := filepath.Base(filePath)
	outputFile := path + "/" + strings.Replace(fileName, ".bin", ".mp4", -1)
	cmd := exec.Command(ffmpeg.Exec, "-i", filePath, "-c", "copy", "-map", "0", outputFile)
	//使用CPU编码
	if ffmpeg.Type == "cpu" {
		cmd = exec.Command(ffmpeg.Exec, "-i", filePath, outputFile)
	} else
	//使用GPU编码
	if ffmpeg.Type == "gpu" {
		cmd = exec.Command(ffmpeg.Exec, "-c:v", ffmpeg.Gpu, "-i", filePath, outputFile)
	}
	//执行命令
	err := cmd.Run()
	if err == nil {
		//删除转码前的文件
		os.Remove(filePath)
	}
}

// 删除文件
func deleteFiles(files []string) {
	for _, file := range files {
		os.Remove(file)
	}
}
