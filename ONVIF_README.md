# GO_UnicomMonitor ONVIF功能说明

## 新增功能

本项目新增了ONVIF协议支持，可以将实时视频流转换为ONVIF协议服务，供录像机等设备连接读取。

## 功能特性

1. **保留原有功能**：原有的视频录制功能完全保留
2. **新增ONVIF服务**：提供ONVIF协议接口，支持录像机连接
3. **流模式支持**：可选择仅提供流服务，不保存文件
4. **多设备支持**：支持多个设备同时提供ONVIF流服务

## 配置说明

### 配置文件 (config.json)

```json
{
  "host": ":25678",
  "user": "root:root",
  "path": "./videos/",
  "sleep": 60,
  "onvifPort": 8080,
  "video": [
    {
      "wsHost": "vd-file-hnzz2-wcloud.wojiazongguan.cn:50443",
      "paramMsg": "MT1234567890==",
      "name": "客厅",
      "size": 10,
      "count": 10,
      "streamOnly": false
    }
  ]
}
```

### 新增配置项

- `onvifPort`: ONVIF服务端口，默认8080
- `onvifAuth`: ONVIF认证开关，默认true，设置为false时无需认证
- `streamOnly`: 仅流模式，设置为true时不保存文件，只提供ONVIF流服务

## 使用方法

### 1. 混合模式（推荐）
```json
{
  "streamOnly": false
}
```
- 同时保存文件到本地
- 提供ONVIF流服务
- 录像机可以通过ONVIF协议连接获取实时流

### 2. 仅流模式
```json
{
  "streamOnly": true
}
```
- 不保存文件到本地
- 仅提供ONVIF流服务
- 节省存储空间，适合纯流传输场景

## ONVIF服务接口

### 设备发现
- URL: `http://localhost:8080/onvif/device_service`
- 支持标准ONVIF设备发现协议

### 流地址
- 格式: `http://localhost:8080/onvif/stream/{设备名称}`
- 示例: `http://localhost:8080/onvif/stream/客厅`

## 录像机连接

1. **添加设备**：在录像机中添加ONVIF设备
2. **设备地址**：输入 `http://localhost:8080/onvif/device_service`
3. **用户名密码**：
   - 如果启用认证（默认）：使用配置文件中的用户名和密码
   - 如果禁用认证：无需填写用户名密码
4. **选择流**：选择对应的设备流进行录制

## 支持的ONVIF操作

- `GetDeviceInformation`: 获取设备信息
- `GetCapabilities`: 获取设备能力
- `GetProfiles`: 获取配置文件列表
- `GetStreamUri`: 获取流地址

## 注意事项

1. **端口配置**：确保ONVIF端口(8080)未被占用
2. **网络访问**：录像机需要能够访问到ONVIF服务端口
3. **设备名称**：设备名称将作为流标识，请使用中文或英文名称
4. **实时性**：ONVIF流为实时传输，延迟较低
5. **安全认证**：默认启用Basic认证，建议在生产环境中保持启用状态

## 故障排除

### 1. ONVIF服务无法启动
- 检查端口是否被占用
- 确认配置文件中的onvifPort设置正确

### 2. 录像机无法连接
- 检查网络连通性
- 确认防火墙设置
- 验证设备地址和端口配置

### 3. 流数据异常
- 检查原始视频流是否正常
- 确认设备配置正确
- 查看程序日志输出

## 版本信息

- 版本：20250325_002
- 新增：ONVIF协议支持
- 作者：zgcwkj
- 开源：https://github.com/zgcwkjOpenProject/GO_UnicomMonitor
