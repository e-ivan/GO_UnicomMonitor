package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// 启动网站服务
func StartHttp(config *Config) {
	// 解析用户名和密码
	username := strings.Split(config.User, ":")[0]
	password := strings.Split(config.User, ":")[1]
	// 静态目录需要认证
	http.Handle("/", basicAuth(http.FileServer(http.Dir("static")).ServeHTTP, username, password))
	// 文件列表需要认证
	http.HandleFunc("/files", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		handleFileList(w, r, config.Path)
	}, username, password))
	// 文件内容需要认证
	http.HandleFunc("/get", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		handleFileContent(w, r, config.Path)
	}, username, password))
	// 启动服务器
	http.ListenAndServe(config.Host, nil)
}

// 身份验证中间件
func basicAuth(next http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != username || pass != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// 处理文件列表请求
func handleFileList(w http.ResponseWriter, r *http.Request, dirPath string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 获取文件列表
	files, err := listFiles(dirPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 返回文件列表
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(files)
}

// 处理文件内容请求
func handleFileContent(w http.ResponseWriter, r *http.Request, dirPath string) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// 获取文件名
	filename := r.URL.Query().Get("file")
	if filename == "" {
		http.Error(w, "File parameter is required", http.StatusBadRequest)
		return
	}
	// 打开文件
	fullPath := filepath.Join(dirPath, filename)
	file, err := os.Open(fullPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
	// 获取文件信息以设置Content-Length
	fileInfo, err := file.Stat()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// 设置响应头
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment; filename="+filepath.Base(filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))
	// 直接将文件流复制到响应流中
	http.ServeContent(w, r, filename, fileInfo.ModTime(), file)
}

// 获取文件列表
func listFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// 将路径转换为相对于start目录的路径
			relPath, err := filepath.Rel(dir, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}
		return nil
	})
	return files, err
}
