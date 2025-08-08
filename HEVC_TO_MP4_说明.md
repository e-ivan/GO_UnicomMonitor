# HEVC到MP4转换功能说明

## 功能概述
本程序现已支持将录制的HEVC格式视频文件自动转换为MP4格式，提供更好的兼容性。

## 前置要求
- 必须安装FFmpeg工具
- Windows: 从 https://ffmpeg.org/download.html 下载并添加到系统PATH
- Linux: `sudo apt install ffmpeg` 或 `sudo yum install ffmpeg`
- macOS: `brew install ffmpeg`

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
- `convertToMp4`: `true` 启用转换，`false` 禁用转换
- 转换会在每个HEVC文件录制完成后自动开始
- 转换完成后会自动删除原HEVC文件以节省空间

## 转换过程
1. 当HEVC文件达到设定大小限制时，文件会被关闭
2. 如果启用了转换功能，程序会自动调用FFmpeg进行转换
3. 转换使用 `copy` 模式，速度快且保持原始质量
4. 转换完成后删除原HEVC文件

## 批量转换现有文件
程序还提供了批量转换功能，可以转换指定目录下的所有HEVC文件。

## 注意事项
- 转换过程会消耗一定的CPU资源
- 请确保有足够的磁盘空间进行转换
- 如果FFmpeg未正确安装，程序会显示警告信息
- 转换过程中如果出现错误，原HEVC文件会被保留

## 故障排除
1. **转换失败**: 检查FFmpeg是否正确安装并在PATH中
2. **磁盘空间不足**: 确保有足够空间存储转换后的文件  
3. **权限问题**: 确保程序有读写目标目录的权限 