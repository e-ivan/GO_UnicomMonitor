package main

import (
	"os/exec"
	"strings"
)

// 检查FFmpeg是否已安装
func CheckFFmpegInstalled() bool {
	_, err := exec.LookPath("ffmpeg")
	return err == nil
}

// 获取FFmpeg版本信息
func GetFFmpegVersion() string {
	cmd := exec.Command("ffmpeg", "-version")
	output, err := cmd.Output()
	if err != nil {
		return "无法获取版本信息"
	}
	
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return lines[0]
	}
	
	return "未知版本"
} 