package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// StreamManager 流管理器
type StreamManager struct {
	streams map[string]*Stream
	mutex   sync.RWMutex
}

// Stream 单个流
type Stream struct {
	ID       string
	Packets  chan []byte
	clients  map[string]*Client
	mutex    sync.RWMutex
	ctx      context.Context
	cancel   context.CancelFunc
}

// Client 客户端连接
type Client struct {
	ID       string
	Type     string // "rtsp", "rtmp", "http", "webrtc"
	Writer   http.ResponseWriter
	Packets  chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewStreamManager 创建新的流管理器
func NewStreamManager() *StreamManager {
	return &StreamManager{
		streams: make(map[string]*Stream),
	}
}

// AddStream 添加新流
func (sm *StreamManager) AddStream(id string) *Stream {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if stream, exists := sm.streams[id]; exists {
		return stream
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	stream := &Stream{
		ID:      id,
		Packets: make(chan []byte, 100),
		clients: make(map[string]*Client),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	sm.streams[id] = stream
	
	// 启动流处理协程
	go stream.process()
	
	return stream
}

// GetStream 获取流
func (sm *StreamManager) GetStream(id string) *Stream {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()
	return sm.streams[id]
}

// RemoveStream 移除流
func (sm *StreamManager) RemoveStream(id string) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()
	
	if stream, exists := sm.streams[id]; exists {
		stream.cancel()
		delete(sm.streams, id)
	}
}

// process 处理流数据
func (s *Stream) process() {
	for {
		select {
		case packet := <-s.Packets:
			s.broadcast(packet)
		case <-s.ctx.Done():
			return
		}
	}
}

// broadcast 广播数据包到所有客户端
func (s *Stream) broadcast(packet []byte) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	for _, client := range s.clients {
		select {
		case client.Packets <- packet:
		default:
			// 客户端缓冲区满，跳过
		}
	}
}

// AddClient 添加客户端
func (s *Stream) AddClient(clientType, clientID string, writer http.ResponseWriter) *Client {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		ID:      clientID,
		Type:    clientType,
		Writer:  writer,
		Packets: make(chan []byte, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	s.clients[clientID] = client
	
	// 启动客户端处理协程
	go client.process()
	
	return client
}

// RemoveClient 移除客户端
func (s *Stream) RemoveClient(clientID string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if client, exists := s.clients[clientID]; exists {
		client.cancel()
		delete(s.clients, clientID)
	}
}

// process 客户端数据处理
func (c *Client) process() {
	defer func() {
		// 清理资源
	}()
	
	for {
		select {
		case packet := <-c.Packets:
			if c.Writer != nil {
				_, err := c.Writer.Write(packet)
				if err != nil {
					log.Printf("写入数据包失败: %v", err)
					return
				}
				c.Writer.(http.Flusher).Flush()
			}
		case <-c.ctx.Done():
			return
		}
	}
}

// StreamServer 流媒体服务器
type StreamServer struct {
	manager *StreamManager
	router  *gin.Engine
	server  *http.Server
}

// NewStreamServer 创建新的流媒体服务器
func NewStreamServer() *StreamServer {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Logger(), gin.Recovery())
	
	server := &StreamServer{
		manager: NewStreamManager(),
		router:  router,
	}
	
	server.setupRoutes()
	
	return server
}

// setupRoutes 设置路由
func (s *StreamServer) setupRoutes() {
	// HTTP-FLV 流
	s.router.GET("/live/:streamID.flv", s.handleHTTPFLV)
	
	// HLS 流
	s.router.GET("/live/:streamID.m3u8", s.handleHLS)
	s.router.GET("/live/:streamID/:segment.ts", s.handleHLSSegment)
	
	// 原始流数据
	s.router.GET("/live/:streamID/raw", s.handleRawStream)
	
	// 状态接口
	s.router.GET("/status", s.handleStatus)
}

// handleHTTPFLV 处理HTTP-FLV请求
func (s *StreamServer) handleHTTPFLV(c *gin.Context) {
	streamID := c.Param("streamID")
	stream := s.manager.GetStream(streamID)
	
	if stream == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}
	
	c.Header("Content-Type", "video/x-flv")
	c.Header("Cache-Control", "no-cache")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Connection", "keep-alive")
	
	clientID := fmt.Sprintf("http-flv-%s-%d", streamID, time.Now().Unix())
	client := stream.AddClient("http-flv", clientID, c.Writer)
	
	// 等待连接关闭
	<-c.Request.Context().Done()
	stream.RemoveClient(clientID)
}

// handleRawStream 处理原始流数据
func (s *StreamServer) handleRawStream(c *gin.Context) {
	streamID := c.Param("streamID")
	stream := s.manager.GetStream(streamID)
	
	if stream == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stream not found"})
		return
	}
	
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Connection", "keep-alive")
	
	clientID := fmt.Sprintf("raw-%s-%d", streamID, time.Now().Unix())
	client := stream.AddClient("raw", clientID, c.Writer)
	
	// 等待连接关闭
	<-c.Request.Context().Done()
	stream.RemoveClient(clientID)
}

// handleHLS 处理HLS主播放列表
func (s *StreamServer) handleHLS(c *gin.Context) {
	streamID := c.Param("streamID")
	
	c.Header("Content-Type", "application/vnd.apple.mpegurl")
	c.Header("Cache-Control", "no-cache")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// 生成M3U8播放列表
	m3u8Content := fmt.Sprintf(`#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.0,
/live/%s/segment.ts
#EXT-X-ENDLIST`, streamID)
	
	c.String(http.StatusOK, m3u8Content)
}

// handleHLSSegment 处理HLS分片
func (s *StreamServer) handleHLSSegment(c *gin.Context) {
	streamID := c.Param("streamID")
	segment := c.Param("segment")
	
	c.Header("Content-Type", "video/MP2T")
	c.Header("Cache-Control", "no-cache")
	c.Header("Access-Control-Allow-Origin", "*")
	
	// 这里应该返回对应的TS分片文件
	// 简化实现，实际应该根据segment参数返回对应文件
	c.String(http.StatusOK, "")
}

// handleStatus 处理状态请求
func (s *StreamServer) handleStatus(c *gin.Context) {
	s.manager.mutex.RLock()
	defer s.manager.mutex.RUnlock()
	
	status := make(map[string]interface{})
	for id, stream := range s.manager.streams {
		stream.mutex.RLock()
		clientCount := len(stream.clients)
		stream.mutex.RUnlock()
		
		status[id] = gin.H{
			"clients": clientCount,
			"active":  true,
		}
	}
	
	c.JSON(http.StatusOK, status)
}

// Start 启动服务器
func (s *StreamServer) Start(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	
	log.Printf("流媒体服务器启动在: %s", addr)
	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *StreamServer) Stop() error {
	if s.server != nil {
		return s.server.Shutdown(context.Background())
	}
	return nil
}

// AddStreamData 添加流数据
func (s *StreamServer) AddStreamData(streamID string, packet []byte) {
	stream := s.manager.GetStream(streamID)
	if stream == nil {
		stream = s.manager.AddStream(streamID)
	}
	
	select {
	case stream.Packets <- packet:
	default:
		// 缓冲区满，丢弃数据包
	}
}
