package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com\yangchengkai1\go-websocket\server-push\impl"
)

var (
	upgrader = websocket.Upgrader{
		// 读取存储空间大小
		ReadBufferSize: 1024,
		// 写入存储空间大小
		WriteBufferSize: 1024,
		// 允许跨域
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func wsHandler(w http.ResponseWriter, r *http.Request) {
	var (
		wsConn *websocket.Conn
		err    error
		// data []byte
		conn *impl.Connection
		data []byte
	)
	// 完成http应答，在httpheader中放下如下参数
	if wsConn, err = upgrader.Upgrade(w, r, nil); err != nil {
		log.Println("connection filed" + err.Error())
		return // 获取连接失败直接返回
	}

	if conn, err = impl.InitConnection(wsConn); err != nil {
		log.Println("init connection error" + err.Error())
		goto ERR
	}

	go func() {
		var err error

		for {
			// 每隔一秒发送一次心跳
			if err = conn.WriteMessage([]byte("heartbeat")); err != nil {
				log.Println(err)
				return
			}
			time.Sleep(1 * time.Second)
		}

	}()

	for {
		if data, err = conn.ReadMessage(); err != nil {
			log.Println(err)
			goto ERR
		}
		if err = conn.WriteMessage(data); err != nil {
			log.Println(err)
			goto ERR
		}
	}

ERR:
	// 关闭当前连接

}

func main() {
	// 当有请求访问ws时，执行此回调方法
	http.HandleFunc("/ws", wsHandler)
	// 监听127.0.0.1:7777
	err := http.ListenAndServe("0.0.0.0:7777", nil)
	if err != nil {
		log.Fatal("ListenAndServe", err.Error())
	}
}
