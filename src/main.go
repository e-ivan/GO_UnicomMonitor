package main

import (
	"net/http"
	"strings"
	"time"
)

// 主函数
func main() {
	FmtPrint("开源：https://github.com/zgcwkjOpenProject/GO_UnicomMonitor")
	FmtPrint("作者：zgcwkj")
	FmtPrint("版本：20250110_001")
	FmtPrint("请尊重开源协议，保留作者信息！")
	FmtPrint("")
	//读取配置文件
	config := GetConfig()
	//启动录制协程
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
	//等待退出
	if config.Host != "" {
		FmtPrint("启动网站服务：" + config.Host)
		fs := http.FileServer(http.Dir("./"))
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			//排除某些文件
			if strings.HasSuffix(r.URL.Path, ".go") {
				http.NotFound(w, r)
				return
			}
			//排除某个特定的文件
			if r.URL.Path == "/config.json" {
				http.NotFound(w, r)
				return
			}
			//如果没有被排除，继续执行文件服务器
			fs.ServeHTTP(w, r)
		})
		http.ListenAndServe(config.Host, nil)
	} else {
		for {
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
			FmtPrint("程序运行正常")
		}
	}
}
