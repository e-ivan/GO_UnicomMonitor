# 实时MP4录制功能说明

## 功能概述
本程序支持录制HEVC格式视频流，并可选择实时转换为MP4格式以提供更好的兼容性。

## 实现原理

### 1. 视频流数据格式
- 从WebSocket接收到的数据是HEVC编码的原始视频流
- 这些数据不能直接保存为MP4文件，因为缺少容器格式信息

### 2. 实时转换流程
1. **接收阶段**：从WebSocket接收HEVC原始视频流数据
2. **转换阶段**：使用FFmpeg管道实时将HEVC流转换为MP4格式
3. **保存阶段**：直接将转换后的MP4数据写入文件

### 3. 实时转换的优势
- **无需临时文件**：不需要先保存HEVC文件再转换
- **节省磁盘空间**：直接生成MP4文件，不产生中间文件
- **更高效**：实时转换，减少磁盘I/O操作
- **即时可用**：生成的MP4文件可以立即播放

## 配置方法

### 1. 修改配置文件
在 `config.json` 中为每个视频设备添加 `convertToMp4` 选项：

```json
{
  "host": ":25678",
  "user": "root:root", 
  "path": "./videos/",
  "sleep": 60,
  "video": [
    {
      "wsHost": "your-websocket-host",
      "paramMsg": "your-param-string",
      "name": "客厅",
      "size": 10,
      "count": 10,
      "convertToMp4": true
    }
  ]
}
```

### 2. 配置选项说明
- `convertToMp4`: `true` 启用实时MP4转换，`false` 录制HEVC格式
- 转换会在录制过程中实时进行
- 直接生成可播放的MP4文件

## 转换过程
1. 程序启动FFmpeg进程，建立管道连接
2. 接收到的HEVC视频流数据通过管道发送给FFmpeg
3. FFmpeg实时将HEVC流转换为MP4格式
4. 转换后的MP4数据直接写入文件
5. 当文件达到指定大小时，关闭当前文件并创建新文件

## 前置要求
- 必须安装FFmpeg工具
- Windows: 从 https://ffmpeg.org/download.html 下载并添加到系统PATH
- Linux: `sudo apt install ffmpeg` 或 `sudo yum install ffmpeg`
- macOS: `brew install ffmpeg`

## 注意事项
- 实时转换会消耗一定的CPU资源
- 请确保有足够的磁盘空间存储MP4文件
- 如果FFmpeg未正确安装，程序会显示警告信息
- 转换过程中如果出现错误，程序会记录错误信息

## 故障排除
1. **转换失败**: 检查FFmpeg是否正确安装并在PATH中
2. **磁盘空间不足**: 确保有足够空间存储MP4文件  
3. **权限问题**: 确保程序有读写目标目录的权限
4. **管道错误**: 检查FFmpeg版本是否支持相关参数 