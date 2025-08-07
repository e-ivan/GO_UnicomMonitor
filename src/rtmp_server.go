package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
)

// RTMP服务器
type RTMPServer struct {
	port     int
	listener net.Listener
	streams  map[string]*RTMPStream
	mutex    sync.RWMutex
	server   *StreamServer
}

// RTMP流
type RTMPStream struct {
	ID       string
	clients  map[string]*RTMPClient
	mutex    sync.RWMutex
	packets  chan []byte
	ctx      context.Context
	cancel   context.CancelFunc
}

// RTMP客户端
type RTMPClient struct {
	ID       string
	conn     net.Conn
	stream   *RTMPStream
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewRTMPServer 创建RTMP服务器
func NewRTMPServer(port int, server *StreamServer) *RTMPServer {
	return &RTMPServer{
		port:    port,
		streams: make(map[string]*RTMPStream),
		server:  server,
	}
}

// Start 启动RTMP服务器
func (rs *RTMPServer) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", rs.port))
	if err != nil {
		return err
	}
	rs.listener = listener
	
	log.Printf("RTMP服务器启动在端口: %d", rs.port)
	
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("接受连接失败: %v", err)
			continue
		}
		
		go rs.handleConnection(conn)
	}
}

// handleConnection 处理RTMP连接
func (rs *RTMPServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	
	// 简单的RTMP处理
	// 这里应该实现完整的RTMP协议
	// 简化实现，只处理基本的握手和流数据
	
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		return
	}
	
	log.Printf("收到RTMP数据: %d bytes", n)
	
	// 发送简单的响应
	response := []byte{0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	conn.Write(response)
}

// AddStream 添加RTMP流
func (rs *RTMPServer) AddStream(streamID string) *RTMPStream {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	if stream, exists := rs.streams[streamID]; exists {
		return stream
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	stream := &RTMPStream{
		ID:      streamID,
		clients: make(map[string]*RTMPClient),
		packets: make(chan []byte, 100),
		ctx:     ctx,
		cancel:  cancel,
	}
	
	rs.streams[streamID] = stream
	
	go stream.process()
	
	return stream
}

// process 处理流数据
func (stream *RTMPStream) process() {
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
func (stream *RTMPStream) broadcast(packet []byte) {
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

// Stop 停止RTMP服务器
func (rs *RTMPServer) Stop() error {
	if rs.listener != nil {
		return rs.listener.Close()
	}
	return nil
}
