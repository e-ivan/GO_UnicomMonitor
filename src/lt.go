package main

import (
	"crypto/tls"
	"io"
	"io/fs"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

// 开始录制
func GoRecording(config *Config, video *Video) {
	//临时变量
	tempPath := config.Path + "/" + video.Name
	//断开后重连
	for {
		//连接服务器传输数据
		success := linkServerAndRecord(video, tempPath, config)
		//检查连接状态
		if !success {
			FmtPrint(video.Name + "设备连接失败，稍后自动重连(" + strconv.Itoa(config.Sleep) + ")")
			timeout := time.Duration(config.Sleep)
			time.Sleep(timeout * time.Second)
			continue
		}
	}
}

// 带重试机制的消息读取方法
func readMessageWithRetry(conn *websocket.Conn, maxRetries int, retryDelay time.Duration, video *Video) ([]byte, error) {
	var lastErr error
	
	for attempt := 0; attempt <= maxRetries; attempt++ {
		_, response, err := conn.ReadMessage()
		if err == nil {
			return response, nil
		}
		
		lastErr = err
		
		// 如果不是最后一次尝试，则重新连接服务器
		if attempt < maxRetries {
			// 根据重试次数计算延迟时间，使用指数退避策略
			currentDelay := retryDelay * time.Duration(1<<attempt) // 2^attempt 倍延迟
			FmtPrint("读取消息失败，尝试重新连接服务器... (第", attempt+1, "次重试，共", maxRetries+1, "次，延迟", currentDelay, "秒)")
			time.Sleep(currentDelay)
			
			// 关闭当前连接
			conn.Close()
			
			// 重新连接服务器
			uri := url.URL{
				Scheme: "wss",
				Host:   video.WsHost,
				Path:   "/h5player/live",
			}
			//跳过证书验证
			dialer := websocket.Dialer{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: true,
				},
			}
			//发起连接
			newConn, _, err := dialer.Dial(uri.String(), nil)
			if err != nil {
				FmtPrint("重新连接服务器失败", err)
				continue
			}
			
			//发送消息
			message := "_paramStr_=" + video.ParamMsg
			err = newConn.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				FmtPrint("重新发送消息失败：", err)
				newConn.Close()
				continue
			}
			
			// 更新连接
			*conn = *newConn
		}
	}
	
	return nil, lastErr
}


// 连接服务器并持续录制
func linkServerAndRecord(video *Video, tempPath string, config *Config) bool {
	uri := url.URL{
		Scheme: "wss",
		Host:   video.WsHost,
		Path:   "/h5player/live",
	}
	//跳过证书验证
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	//发起连接
	conn, _, err := dialer.Dial(uri.String(), nil)
	if err != nil {
		FmtPrint("无法连接到服务器", err)
		return false
	}
	defer conn.Close()
	
	//发送消息
	message := "_paramStr_=" + video.ParamMsg
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		FmtPrint("发送消息失败：", err)
		return false
	}
	
	//初始化文件相关变量
	var currentFile io.WriteCloser
	var currentFileName string
	var currentFileSize int64
	maxFileSize := int64(video.Size * 1024 * 1024) // 转换为字节

	// 重试配置
	maxRetries := 5
	retryDelay := 2 * time.Second
		
	//持续接收视频流并写入文件
	for {
		response, err := readMessageWithRetry(conn, maxRetries, retryDelay, video)
		if err != nil {
			FmtPrint("接收消息失败：", err)
			
			// 连接断开时，如果有正在录制的文件，进行清理
			if currentFile != nil {
				currentFile.Close()
				FmtPrint("连接断开，完成当前文件录制：" + currentFileName)
			}
			
			return false
		}
		
		//检查数据有效性
		if len(response) > 1 {
			// 如果启用了ONVIF服务，更新视频流
			if config.OnvifPort > 0 {
				UpdateVideoStream(video.Name, response)
			}
			
			// 如果不是仅流模式，则保存文件
			if !video.StreamOnly {
				//检查是否需要创建新文件
				if currentFile == nil {
					//创建第一个文件
					currentFileName = getFileName(tempPath) + ".hevc"
					FmtPrint("开始录制HEVC格式：" + currentFileName)
					currentFile, err = os.OpenFile(currentFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
					
					if err != nil {
						FmtPrint("创建文件失败: ", err)
						return false
					}
					currentFileSize = 0
				} else if currentFileSize >= maxFileSize {
					//关闭当前文件
					currentFile.Close()
					FmtPrint("文件大小达到限制，完成录制：" + currentFileName)
					
					//创建新文件
					currentFileName = getFileName(tempPath) + ".hevc"
					FmtPrint("开始录制新HEVC文件：" + currentFileName)
					currentFile, err = os.OpenFile(currentFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
					
					if err != nil {
						FmtPrint("创建新文件失败: ", err)
						return false
					}
					currentFileSize = 0
				}
				
				//写入数据到当前文件
				bytesWritten, writeErr := currentFile.Write(response)
				if writeErr != nil {
					FmtPrint("写入文件失败：", writeErr)
					return false
				}
				
				//更新文件大小
				currentFileSize += int64(bytesWritten)
				
				//强制刷新缓冲区，确保数据及时写入磁盘
				// 对于io.WriteCloser，我们无法直接调用Sync，所以这里不做任何操作
				// 如果需要同步，可以考虑使用带缓冲的写入器
			}
		}
	}
}

// 连接服务器（保留原函数以备需要）
func linkServer(video *Video) []byte {
	bytes := []byte{}
	uri := url.URL{
		Scheme: "wss",
		Host:   video.WsHost,
		Path:   "/h5player/live",
	}
	//跳过证书验证
	dialer := websocket.Dialer{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	//发起连接
	conn, _, err := dialer.Dial(uri.String(), nil)
	if err != nil {
		FmtPrint("无法连接到服务器", err)
		return bytes
	}
	defer conn.Close()
	//发送消息
	message := "_paramStr_=" + video.ParamMsg
	err = conn.WriteMessage(websocket.TextMessage, []byte(message))
	if err != nil {
		FmtPrint("发送消息失败：", err)
		return bytes
	}
	//接收消息
	for {
		_, response, err := conn.ReadMessage()
		if err != nil {
			FmtPrint("接收消息失败：", err)
			return bytes
		}
		//检查特定条件
		if len(response) > 1 {
			//打印数据的长度
			//FmtPrint("数据长度：", len(bytes))
			//拼接数据
			bytes = append(bytes, response[:]...)
			//结束条件
			if len(bytes) > 1024*1024*video.Size {
				//结束
				return bytes
			}
		}
	}
}

// 获取文件名称
func getFileName(dirPath string) string {
	//添加日期文件夹
	dateFolder := time.Now().Format("20060102")
	fullPath := dirPath + "/" + dateFolder
	//检查文件夹是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		//文件夹不存在，创建它
		err := os.MkdirAll(fullPath, 0755)
		if err != nil {
			FmtPrint("创建文件夹失败：", err)
			os.Exit(0)
		}
	}
	//文件名称
	fileName := time.Now().Format("150405")
	tempPathh := fullPath + "/" + fileName
	return tempPathh
}

// 保存文件
func saveFile(fileName string, bytes *[]byte) {
	file, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		FmtPrint("保存文件失败: ", err)
		os.Exit(0)
	}
	defer file.Close()
	file.Write(*bytes)
}

// 删除文件夹下的旧文件夹
func DeleteOldFiles(config *Config, video *Video) {
	//临时变量
	dirPath := config.Path + "/" + video.Name
	foldersToKeep := video.Count
	//读取文件夹
	var folders []fs.FileInfo
	entries, _ := os.ReadDir(dirPath)
	for _, entry := range entries {
		if entry.IsDir() {
			info, _ := os.Stat(filepath.Join(dirPath, entry.Name()))
			folders = append(folders, info)
		}
	}
	//检查文件夹数量
	if len(folders) <= foldersToKeep {
		return
	}
	//按时间排序
	sort.Slice(folders, func(i, j int) bool {
		return folders[i].ModTime().After(folders[j].ModTime())
	})
	//删除最旧的文件夹
	for i := foldersToKeep; i < len(folders); i++ {
		oldFolder := filepath.Join(dirPath, folders[i].Name())
		_ = os.RemoveAll(oldFolder)
	}
}


