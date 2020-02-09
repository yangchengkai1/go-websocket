package impl

import (
	"errors"
	"log"
	"sync"

	"github.com/gorilla/websocket"
)

type Connection struct {
	// 存放websocket连接
	wsConn *websocket.Conn
	// 用于存放数据
	inChan chan []byte
	// 用于读取数据
	outChan   chan []byte
	closeChan chan byte
	mutex     sync.Mutex
	// chan是否被关闭
	isClosed bool
}

// 读取Api
func (conn *Connection) ReadMessage() (data []byte, err error) {
	select {
	case data = <-conn.inChan:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return data, err
}

// 发送Api
func (conn *Connection) WriteMessage(data []byte) (err error) {
	select {
	case conn.outChan <- data:
	case <-conn.closeChan:
		err = errors.New("connection is closed")
	}

	return err
}

// 关闭连接的Api
func (conn *Connection) Close() {
	// 线程安全的Close，可以并发多次调用也叫做可重入的Close
	conn.wsConn.Close()
	conn.mutex.Lock()
	if !conn.isClosed {
		// 关闭chan,但是chan只能关闭一次
		close(conn.closeChan)
		conn.isClosed = true
	}

	conn.mutex.Unlock()

}

// 初始化长连接
func InitConnection(wsConn *websocket.Conn) (conn *Connection, err error) {
	conn = &Connection{
		wsConn:    wsConn,
		inChan:    make(chan []byte, 1000),
		outChan:   make(chan []byte, 1000),
		closeChan: make(chan byte, 1),
	}

	go conn.readLoop()
	go conn.writeLoop()

	return conn, err
}

// 内部实现
func (conn *Connection) readLoop() {
	var (
		data []byte
		err  error
	)

	for {
		if _, data, err = conn.wsConn.ReadMessage(); err != nil {
			log.Println("ReadMessage error" + err.Error())
			goto ERR
		}
		// 容易阻塞到这里，等待inChan有空闲的位置
		select {
		case conn.inChan <- data:
		case <-conn.closeChan: // closeChan关闭的时候执行
			goto ERR
		}
	}

ERR:
	conn.Close()
}

func (conn *Connection) writeLoop() {
	var (
		data []byte
		err  error
	)

	for {
		select {
		case data = <-conn.outChan:
		case <-conn.closeChan:
			goto ERR
		}

		data = <-conn.outChan
		if err = conn.wsConn.WriteMessage(websocket.TextMessage, data); err != nil {
			log.Println("WriteMessage error" + err.Error())
			goto ERR
		}
	}

ERR:
	conn.Close()
}
