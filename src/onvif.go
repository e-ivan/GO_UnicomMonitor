package main

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ONVIF服务管理器
type OnvifManager struct {
	streams map[string]*VideoStream
	mutex   sync.RWMutex
}

// 视频流结构
type VideoStream struct {
	Name     string
	Data     []byte
	LastSeen time.Time
	mutex    sync.RWMutex
}

// 全局ONVIF管理器
var onvifManager = &OnvifManager{
	streams: make(map[string]*VideoStream),
}

// 全局配置引用
var globalConfig *Config

// 初始化ONVIF服务
func InitOnvifService(config *Config) {
	globalConfig = config
	// 启动ONVIF服务
	go func() {
		FmtPrint("启动ONVIF服务，端口：", config.OnvifPort)
		http.HandleFunc("/onvif/device_service", handleOnvifRequest)
		http.HandleFunc("/onvif/stream/", handleStreamRequest)
		http.ListenAndServe(":"+strconv.Itoa(config.OnvifPort), nil)
	}()

	// 启动流清理协程
	go func() {
		for {
			time.Sleep(30 * time.Second)
			onvifManager.cleanupOldStreams()
		}
	}()
}

// 处理ONVIF请求
func handleOnvifRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 检查认证
	if !checkOnvifAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="ONVIF Device"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 读取请求体
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	// 解析SOAP请求
	soapAction := r.Header.Get("SOAPAction")
	response := handleSoapRequest(string(body), soapAction)

	// 设置响应头
	w.Header().Set("Content-Type", "application/soap+xml; charset=utf-8")
	w.Header().Set("Content-Length", strconv.Itoa(len(response)))
	w.Write([]byte(response))
}

// 处理SOAP请求
func handleSoapRequest(body, soapAction string) string {
	// 解析XML
	var envelope struct {
		XMLName xml.Name `xml:"Envelope"`
		Body    struct {
			XMLName xml.Name `xml:"Body"`
			Content []byte   `xml:",innerxml"`
		} `xml:"Body"`
	}

	err := xml.Unmarshal([]byte(body), &envelope)
	if err != nil {
		return createSoapFault("Failed to parse SOAP request")
	}

	// 根据SOAP Action处理不同的请求
	switch {
	case strings.Contains(soapAction, "GetDeviceInformation"):
		return handleGetDeviceInformation()
	case strings.Contains(soapAction, "GetCapabilities"):
		return handleGetCapabilities()
	case strings.Contains(soapAction, "GetProfiles"):
		return handleGetProfiles()
	case strings.Contains(soapAction, "GetStreamUri"):
		return handleGetStreamUri()
	default:
		return createSoapFault("Unsupported operation")
	}
}

// 处理获取设备信息
func handleGetDeviceInformation() string {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetDeviceInformationResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Manufacturer>GO_UnicomMonitor</tds:Manufacturer>
      <tds:Model>Video Stream Server</tds:Model>
      <tds:FirmwareVersion>1.0.0</tds:FirmwareVersion>
      <tds:SerialNumber>UNICOM001</tds:SerialNumber>
      <tds:HardwareId>HW001</tds:HardwareId>
    </tds:GetDeviceInformationResponse>
  </soap:Body>
</soap:Envelope>`
	return response
}

// 处理获取设备能力
func handleGetCapabilities() string {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <tds:GetCapabilitiesResponse xmlns:tds="http://www.onvif.org/ver10/device/wsdl">
      <tds:Capabilities>
        <tds:Device>
          <tt:XAddr>http://localhost:8080/onvif/device_service</tt:XAddr>
        </tds:Device>
        <tds:Media>
          <tt:XAddr>http://localhost:8080/onvif/device_service</tt:XAddr>
          <tt:StreamingCapabilities>
            <tt:RTPMulticast>false</tt:RTPMulticast>
            <tt:RTP_TCP>true</tt:RTP_TCP>
            <tt:RTP_RTSP_TCP>true</tt:RTP_RTSP_TCP>
          </tt:StreamingCapabilities>
        </tds:Media>
      </tds:Capabilities>
    </tds:GetCapabilitiesResponse>
  </soap:Body>
</soap:Envelope>`
	return response
}

// 处理获取配置文件
func handleGetProfiles() string {
	if globalConfig == nil {
		return createSoapFault("Configuration not available")
	}

	// 生成配置文件XML
	profilesXML := ""
	for i, video := range globalConfig.Video {
		profileToken := fmt.Sprintf("Profile_%d", i+1)
		videoSourceToken := fmt.Sprintf("VideoSource_%d", i+1)
		videoEncoderToken := fmt.Sprintf("VideoEncoder_%d", i+1)
		
		profileXML := fmt.Sprintf(`        <tt:Profile token="%s">
          <tt:Name>%s</tt:Name>
          <tt:VideoSourceConfiguration token="%s">
            <tt:Name>%s Source</tt:Name>
            <tt:UseCount>1</tt:UseCount>
            <tt:SourceToken>%s</tt:SourceToken>
            <tt:Bounds x="0" y="0" width="1920" height="1080"/>
          </tt:VideoSourceConfiguration>
          <tt:VideoEncoderConfiguration token="%s">
            <tt:Name>%s Encoder</tt:Name>
            <tt:UseCount>1</tt:UseCount>
            <tt:Encoding>H264</tt:Encoding>
            <tt:Resolution>
              <tt:Width>1920</tt:Width>
              <tt:Height>1080</tt:Height>
            </tt:Resolution>
            <tt:Quality>5</tt:Quality>
            <tt:RateControl>
              <tt:FrameRateLimit>30</tt:FrameRateLimit>
              <tt:BitrateLimit>2048</tt:BitrateLimit>
            </tt:RateControl>
          </tt:VideoEncoderConfiguration>
        </tt:Profile>`, profileToken, video.Name, videoSourceToken, video.Name, videoSourceToken, videoEncoderToken, video.Name)
		
		profilesXML += profileXML
	}

	response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <trt:GetProfilesResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:Profiles>
%s
      </trt:Profiles>
    </trt:GetProfilesResponse>
  </soap:Body>
</soap:Envelope>`, profilesXML)
	
	return response
}

// 处理获取流URI
func handleGetStreamUri() string {
	if globalConfig == nil {
		return createSoapFault("Configuration not available")
	}

	// 默认返回第一个设备的流URI
	deviceName := "客厅"
	if len(globalConfig.Video) > 0 {
		deviceName = globalConfig.Video[0].Name
	}

	response := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <trt:GetStreamUriResponse xmlns:trt="http://www.onvif.org/ver10/media/wsdl">
      <trt:MediaUri>
        <tt:Uri>http://localhost:%d/onvif/stream/%s</tt:Uri>
        <tt:InvalidAfterConnect>false</tt:InvalidAfterConnect>
        <tt:InvalidAfterReboot>false</tt:InvalidAfterReboot>
        <tt:Timeout>PT60S</tt:Timeout>
      </trt:MediaUri>
    </trt:GetStreamUriResponse>
  </soap:Body>
</soap:Envelope>`, globalConfig.OnvifPort, deviceName)
	
	return response
}

// 创建SOAP错误响应
func createSoapFault(faultString string) string {
	response := `<?xml version="1.0" encoding="UTF-8"?>
<soap:Envelope xmlns:soap="http://www.w3.org/2003/05/soap-envelope">
  <soap:Body>
    <soap:Fault>
      <soap:Code>
        <soap:Value>soap:Sender</soap:Value>
      </soap:Code>
      <soap:Reason>
        <soap:Text>` + faultString + `</soap:Text>
      </soap:Reason>
    </soap:Fault>
  </soap:Body>
</soap:Envelope>`
	return response
}

// 处理流请求
func handleStreamRequest(w http.ResponseWriter, r *http.Request) {
	// 检查认证
	if !checkOnvifAuth(r) {
		w.Header().Set("WWW-Authenticate", `Basic realm="ONVIF Stream"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// 从URL路径中提取流名称
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid stream path", http.StatusBadRequest)
		return
	}
	streamName := pathParts[len(pathParts)-1]

	// 设置响应头
	w.Header().Set("Content-Type", "video/mp4")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 获取流数据
	stream := onvifManager.getStream(streamName)
	if stream == nil {
		http.Error(w, "Stream not found", http.StatusNotFound)
		return
	}

	// 发送流数据
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()
	
	if len(stream.Data) > 0 {
		w.Write(stream.Data)
	} else {
		// 如果没有数据，发送一个简单的MP4头
		w.Write([]byte{0x00, 0x00, 0x00, 0x20, 0x66, 0x74, 0x79, 0x70, 0x6D, 0x70, 0x34, 0x32})
	}
}

// 更新视频流数据
func UpdateVideoStream(name string, data []byte) {
	onvifManager.updateStream(name, data)
}

// 获取流
func (om *OnvifManager) getStream(name string) *VideoStream {
	om.mutex.RLock()
	defer om.mutex.RUnlock()
	return om.streams[name]
}

// 更新流
func (om *OnvifManager) updateStream(name string, data []byte) {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	stream, exists := om.streams[name]
	if !exists {
		stream = &VideoStream{
			Name:     name,
			LastSeen: time.Now(),
		}
		om.streams[name] = stream
	}

	stream.mutex.Lock()
	stream.Data = data
	stream.LastSeen = time.Now()
	stream.mutex.Unlock()
}

// 清理旧流
func (om *OnvifManager) cleanupOldStreams() {
	om.mutex.Lock()
	defer om.mutex.Unlock()

	now := time.Now()
	for name, stream := range om.streams {
		if now.Sub(stream.LastSeen) > 5*time.Minute {
			delete(om.streams, name)
			FmtPrint("清理过期流：", name)
		}
	}
}

// 检查ONVIF认证
func checkOnvifAuth(r *http.Request) bool {
	if globalConfig == nil {
		return false
	}

	// 如果禁用了ONVIF认证，直接返回true
	if !globalConfig.OnvifAuth {
		return true
	}

	// 解析用户名和密码
	userParts := strings.Split(globalConfig.User, ":")
	if len(userParts) != 2 {
		return false
	}
	expectedUser := userParts[0]
	expectedPass := userParts[1]

	// 获取Basic认证信息
	user, pass, ok := r.BasicAuth()
	if !ok {
		return false
	}

	// 验证用户名和密码
	return user == expectedUser && pass == expectedPass
}
