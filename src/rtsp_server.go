package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

// RTSP服务器
type RTSPServer struct {
	port     int
	listener net.Listener
	streams  map[string]*RTSPStream
	mutex    sync.RWMutex
	server   *StreamServer
}

// RTSP流
type RTSPStream struct {
	ID       string
	clients  map[string]*RTSPClient
	mutex    sync.RWMutex
	packets  chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
}

// RTSP客户端
type RTSPClient struct {
	ID       string
	conn     net.Conn
	stream   *RTSPStream
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRTSPServer 创建RTSP服务器
func NewRTSPServer(port int, server *StreamServer) *RTSPServer {
	return &RTSPServer{
		port:    port,
		streams: make(map[string]*RTSPStream),
		server:  server,
	}
}

// Start 启动RTSP服务器
func (rs *RTSPServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rs.port))
	if err != nil {
		return err
	}
	rs.listener = listener
	
	log.Printf("RTSP服务器启动在端口: %d", rs.port)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}
		
		go rs.handleConnection(conn)
	}
}

// handleConnection 处理RTSP连接
func (rs *RTSPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// 简单的RTSP处理
	// 这里应该实现完整的RTSP协议
	// 简化实现，只处理基本的DESCRIBE和PLAY请求
	
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return
	}
	
	request := string(buffer[:n])
	log.Printf("收到RTSP请求: %s", request)
	
	// 发送简单的响应
	response := "RTSP/1.0 200 OK\r\n" +
		"CSeq: 1\r\n" +
		"Content-Type: application/sdp\r\n" +
		"Content-Length: 0\r\n" +
		"\r\n"
	
	conn.Write([]byte(response))
}

// AddStream 添加RTSP流
func (rs *RTSPServer) AddStream(streamID string) *RTSPStream {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	if stream, exists := rs.streams[streamID]; exists {
		return stream
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	stream := &RTSPStream{
		ID:      streamID,
		clients: make(map[string]*RTSPClient),
		packets: make(chan []byte, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	rs.streams[streamID] = stream
	
	go stream.process()
	
	return stream
}

// process 处理流数据
func (stream *RTSPStream) process() {
	for {
		select {
		case packet := <-stream.packets:
			stream.broadcast(packet)
		case <-stream.ctx.Done():
			return
		}
	}
}

// broadcast 广播数据到所有客户端
func (stream *RTSPStream) broadcast(packet []byte) {
	stream.mutex.RLock()
	defer stream.mutex.RUnlock()
	
	for _, client := range stream.clients {
		select {
		case client.conn.Write(packet):
		default:
			// 写入失败，跳过
		}
	}
}

// Stop 停止RTSP服务器
func (rs *RTSPServer) Stop() error {
	if rs.listener != nil {
		return rs.listener.Close()
	}
	return nil
}
