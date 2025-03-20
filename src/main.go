package main

import "time"

// 主函数
func main() {
	FmtPrint("开源：https://github.com/zgcwkjOpenProject/GO_UnicomMonitor")
	FmtPrint("作者：zgcwkj")
	FmtPrint("版本：20250320_001")
	FmtPrint("请尊重开源协议，保留作者信息！")
	FmtPrint("")
	//读取配置文件
	config := GetConfig()
	if config.Path == "" {
		config.Path = "./"
	}
	//启动录制协程
	FmtPrint("启动录制服务，存储路径：" + config.Path)
	for _, video := range config.Video {
		go GoRecording(&config, &video)
	}
	//删除文件协程
	go func() {
		for {
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
			//FmtPrint("执行删除文件")
			//删除旧文件
			for _, video := range config.Video {
				DeleteOldFiles(video.Name, video.Count)
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
