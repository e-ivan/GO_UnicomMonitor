package main

import "time"

// 主函数
func main() {
	FmtPrint("开源：https://github.com/zgcwkjOpenProject/GO_UnicomMonitor")
	FmtPrint("作者：zgcwkj")
	FmtPrint("版本：20250325_001")
	FmtPrint("请尊重开源协议，保留作者信息！")
	FmtPrint("")
	
	// 检查FFmpeg是否已安装
	if CheckFFmpegInstalled() {
		FmtPrint("FFmpeg已安装: " + GetFFmpegVersion())
		FmtPrint("支持实时HEVC到MP4转换功能")
	} else {
		FmtPrint("警告: 未检测到FFmpeg，无法使用实时MP4转换功能")
		FmtPrint("请安装FFmpeg以启用实时HEVC到MP4转换: https://ffmpeg.org/download.html")
	}
	FmtPrint("")
	
	//读取配置文件
	config := GetConfig()
	if config.Path == "" {
		config.Path = "./"
	}
	//启动录制协程
	FmtPrint("启动录制服务，存储路径：" + config.Path)
	for _, video := range config.Video {
		if video.ConvertToMp4 {
			FmtPrint("设备 " + video.Name + " 已启用实时MP4转换功能")
		} else {
			FmtPrint("设备 " + video.Name + " 录制HEVC格式视频")
		}
		go GoRecording(&config, &video)
	}
	//删除旧文件协程
	go func() {
		for {
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
			//FmtPrint("执行删除旧文件录像")
			for _, video := range config.Video {
				DeleteOldFiles(&config, &video)
			}
		}
	}()
	//运行类型
	if config.Host == "" {
		//后台运行
		for {
			FmtPrint("程序运行正常")
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
		}
	} else {
		//网站服务
		FmtPrint("启动网站服务：" + config.Host)
		//启动网站服务
		StartHttp(&config)
	}
}
