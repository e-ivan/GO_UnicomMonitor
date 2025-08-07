package main

import "time"

// 全局流媒体服务器实例
var streamServer *StreamServer
var rtspServer *RTSPServer
var rtmpServer *RTMPServer

// 主函数
func main() {
	FmtPrint("开源：https://github.com/zgcwkjOpenProject/GO_UnicomMonitor")
	FmtPrint("作者：zgcwkj")
	FmtPrint("版本：20250325_003 - 多协议流媒体服务器版本")
	FmtPrint("请尊重开源协议，保留作者信息！")
	FmtPrint("")
	
	//读取配置文件
	config := GetConfig()
	if config.Path == "" {
		config.Path = "./"
	}
	
	// 初始化流媒体服务器
	streamServer = NewStreamServer()
	
	// 初始化RTSP服务器
	if config.RTSPPort > 0 {
		rtspServer = NewRTSPServer(config.RTSPPort, streamServer)
	}
	
	// 初始化RTMP服务器
	if config.RTMPPort > 0 {
		rtmpServer = NewRTMPServer(config.RTMPPort, streamServer)
	}
	
	//启动录制协程
	FmtPrint("启动录制服务，存储路径：" + config.Path)
	for _, video := range config.Video {
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
	if config.Host == "" && config.StreamHost == "" && config.RTSPPort == 0 && config.RTMPPort == 0 {
		//后台运行
		for {
			FmtPrint("程序运行正常")
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
		}
	} else {
		//启动RTSP服务器
		if config.RTSPPort > 0 {
			FmtPrint("启动RTSP服务器，端口：" + fmt.Sprintf("%d", config.RTSPPort))
			go func() {
				if err := rtspServer.Start(); err != nil {
					FmtPrint("RTSP服务器启动失败：", err)
				}
			}()
		}
		
		//启动RTMP服务器
		if config.RTMPPort > 0 {
			FmtPrint("启动RTMP服务器，端口：" + fmt.Sprintf("%d", config.RTMPPort))
			go func() {
				if err := rtmpServer.Start(); err != nil {
					FmtPrint("RTMP服务器启动失败：", err)
				}
			}()
		}
		
		//启动流媒体服务器
		if config.StreamHost != "" {
			FmtPrint("启动HTTP流媒体服务器：" + config.StreamHost)
			go func() {
				if err := streamServer.Start(config.StreamHost); err != nil {
					FmtPrint("流媒体服务器启动失败：", err)
				}
			}()
		}
		
		//启动网站服务
		if config.Host != "" {
			FmtPrint("启动网站服务：" + config.Host)
			StartHttp(&config)
		} else {
			// 只启动流媒体服务器时，保持程序运行
			for {
				FmtPrint("流媒体服务器运行正常")
				timeout := time.Duration(config.Sleep)
				time.Sleep(timeout * time.Second)
			}
		}
	}
}
