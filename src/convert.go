package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// 转换配置选项
type ConvertOptions struct {
	DeleteOriginal bool   // 是否删除原文件
	Quality        string // 视频质量 (high, medium, low)
	OutputFormat   string // 输出格式 (mp4, avi, mkv)
}

// 批量转换指定目录下的所有HEVC文件
func BatchConvertHevcFiles(dirPath string, options ConvertOptions) {
	FmtPrint("开始批量转换目录: " + dirPath)
	
	// 检查FFmpeg是否可用
	_, err := exec.LookPath("ffmpeg")
	if err != nil {
		FmtPrint("FFmpeg未找到，无法转换视频格式。请安装FFmpeg: ", err)
		return
	}
	
	// 遍历目录查找HEVC文件
	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// 检查是否为HEVC文件
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".hevc") {
			convertSingleFile(path, options)
		}
		
		return nil
	})
	
	if err != nil {
		FmtPrint("遍历目录失败: ", err)
	}
	
	FmtPrint("批量转换完成")
}

// 转换单个文件
func convertSingleFile(hevcFilePath string, options ConvertOptions) {
	// 生成输出文件路径
	outputExt := options.OutputFormat
	if outputExt == "" {
		outputExt = "mp4"
	}
	outputFilePath := strings.TrimSuffix(hevcFilePath, ".hevc") + "." + outputExt
	
	FmtPrint("转换文件: " + hevcFilePath + " -> " + outputFilePath)
	
	// 构建FFmpeg命令参数
	args := []string{
		"-i", hevcFilePath,
		"-y", // 覆盖输出文件
	}
	
	// 根据质量设置编码参数
	switch options.Quality {
	case "high":
		args = append(args, "-c:v", "libx264", "-crf", "18", "-preset", "slow")
	case "medium":
		args = append(args, "-c:v", "libx264", "-crf", "23", "-preset", "medium")
	case "low":
		args = append(args, "-c:v", "libx264", "-crf", "28", "-preset", "fast")
	default:
		// 默认使用复制模式（最快，保持原质量）
		args = append(args, "-c", "copy")
	}
	
	// 添加输出文件
	args = append(args, outputFilePath)
	
	// 执行转换
	cmd := exec.Command("ffmpeg", args...)
	err := cmd.Run()
	if err != nil {
		FmtPrint("转换失败: ", err)
		return
	}
	
	FmtPrint("转换完成: " + outputFilePath)
	
	// 如果配置了删除原文件
	if options.DeleteOriginal {
		err = os.Remove(hevcFilePath)
		if err != nil {
			FmtPrint("删除原文件失败: ", err)
		} else {
			FmtPrint("已删除原文件: " + hevcFilePath)
		}
	}
}

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