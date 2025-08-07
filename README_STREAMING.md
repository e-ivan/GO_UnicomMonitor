# GO_UnicomMonitor 流媒体服务器版本

## 功能概述

本项目已升级为多协议流媒体服务器，支持以下功能：

1. **从网络获取视频流** - 通过WebSocket连接获取视频流数据
2. **多协议输出** - 支持HTTP-FLV、RTSP、RTMP等多种协议
3. **实时流媒体服务** - 将获取的视频流实时转发给录像机或其他客户端

## 支持的协议

### HTTP-FLV
- 地址格式：`http://localhost:8080/live/{流名称}.flv`
- 适用于网页播放器

### 原始流数据
- 地址格式：`http://localhost:8080/live/{流名称}/raw`
- 适用于自定义客户端

### RTSP
- 地址格式：`rtsp://localhost:8554/{流名称}`
- 适用于专业录像设备

### RTMP
- 地址格式：`rtmp://localhost:1935/live/{流名称}`
- 适用于直播平台

## 配置文件

```json
{
  "host": ":25678",
  "stream_host": ":8080",
  "rtsp_port": 8554,
  "rtmp_port": 1935,
  "user": "root:root",
  "path": "./videos/",
  "sleep": 60,
  "video": [
    {
      "wsHost": "vd-file-hnzz2-wcloud.wojiazongguan.cn:50443",
      "paramMsg": "MT1234567890==",
      "name": "客厅",
      "size": 10,
      "count": 10
    }
  ]
}
```

### 配置说明

- `host`: 网站服务端口（可选）
- `stream_host`: HTTP流媒体服务器端口
- `rtsp_port`: RTSP服务器端口
- `rtmp_port`: RTMP服务器端口
- `video`: 视频源配置列表
  - `name`: 流名称，用于生成访问地址
  - `wsHost`: WebSocket服务器地址
  - `paramMsg`: 连接参数

## 使用方法

### 1. 启动服务器

```bash
cd src
go run .
```

### 2. 访问流媒体

根据配置的流名称，可以通过以下地址访问：

- **客厅摄像头**：
  - HTTP-FLV: `http://localhost:8080/live/客厅.flv`
  - 原始流: `http://localhost:8080/live/客厅/raw`
  - RTSP: `rtsp://localhost:8554/客厅`
  - RTMP: `rtmp://localhost:1935/live/客厅`

### 3. 网页监控

访问 `http://localhost:25678/stream.html` 查看流媒体状态页面。

## 录像机集成

### 支持RTSP的录像机
使用RTSP地址：`rtsp://服务器IP:8554/客厅`

### 支持RTMP的录像机
使用RTMP地址：`rtmp://服务器IP:1935/live/客厅`

### 支持HTTP-FLV的录像机
使用HTTP-FLV地址：`http://服务器IP:8080/live/客厅.flv`

## 技术特点

1. **低延迟** - 直接从WebSocket获取数据并转发
2. **多协议支持** - 同时支持HTTP、RTSP、RTMP协议
3. **实时性** - 无需保存文件，直接流式传输
4. **可扩展** - 支持多个视频源同时流媒体
5. **稳定性** - 自动重连机制，确保服务稳定

## 注意事项

1. 确保防火墙开放相应端口
2. RTSP和RTMP协议需要专业播放器支持
3. 建议在局域网内使用，避免公网暴露
4. 可以根据需要调整缓冲区大小和连接数限制

## 版本历史

- v20250325_001: 原始录制版本
- v20250325_002: 流媒体服务器版本
- v20250325_003: 多协议流媒体服务器版本（当前版本）
