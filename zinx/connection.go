package zinx

import (
	"cinx/zinx/utils/global"
	"cinx/zinx/ziface"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
)

type Connection struct {
	ziface.IConnection
	//当前Conn属于哪个Server
	TcpServer ziface.IServer //当前conn属于哪个server，在conn初始化的时候添加即可
	//当前连接的socket TCP套接字
	Conn *net.TCPConn
	//当前连接的ID 也可以称作为SessionID，ID全局唯一
	ConnID uint32
	//当前连接的关闭状态
	isClosed *sync.Once

	//消息管理MsgId和对应处理方法的消息管理模块
	MsgHandler ziface.IMsgHandle

	//告知该链接已经退出/停止的channel
	ExitBuffChan chan struct{}

	msgChan chan []byte
	//有关冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte //定义channel成员
	// ================================
	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
	// ================================
}

// NewConnection 创建连接的方法
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TcpServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     nil,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan struct{}),
		msgChan:      make(chan []byte),                       //msgChan初始化
		msgBuffChan:  make(chan []byte, global.MaxMsgChanLen), //不要忘记初始化
		property:     make(map[string]interface{}),            //对链接属性map初始化

	}
	//将新创建的Conn添加到链接管理中
	c.TcpServer.GetConnMgr().Add(c) //将当前新创建的连接添加到ConnManager中
	return c
}

// StartReader 处理conn读数据的Goroutine
func (c *Connection) StartReader() {
	slog.Info("[cinx] Server Reader Goroutine is running...")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit!")

	defer c.Stop()

	for {
		//进行拆包，对消息进行拆解
		dp := NewDataPack()

		//获得message的头部消息
		headData := make([]byte, dp.GetHeadLen())
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			slog.Error("[cinx] read message ", "error", err)
			c.ExitBuffChan <- struct{}{}
			continue
		}

		msg, err := dp.Unpack(headData)
		if err != nil {
			slog.Error("[cinx] unpack ", "error", err)
			c.ExitBuffChan <- struct{}{}
			continue
		}

		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				slog.Error("[cinx] read message ", "error", err)
				c.ExitBuffChan <- struct{}{}
				continue
			}
		}
		msg.SetData(data)

		//得到当前客户端请求的Request数据
		req := Request{
			conn: c,
			msg:  msg, //将之前的buf 改成 msg
		}
		if global.WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(&req)
		} else {
			//从路由Routers 中找到注册绑定Conn的对应Handle
			go c.MsgHandler.DoMsgHandler(&req)
		}

	}
}

// Start 启动连接，让当前连接开始工作
func (c *Connection) Start() {

	//开启处理该链接读取到客户端数据之后的请求业务
	go c.StartReader()
	go c.StartWriter()
	c.TcpServer.CallOnConnStart(c)
	for {
		select {
		case <-c.ExitBuffChan: //得到退出消息，不再阻塞
			return
		}
	}
}

// Stop 停止连接，结束当前连接状态M
func (c *Connection) Stop() {
	//1. 如果当前链接已经关闭
	c.isClosed.Do(
		func() {
			c.TcpServer.CallOnConnStop(c)
			// 关闭socket链接
			c.Conn.Close()
			//通知从缓冲队列读数据的业务，该链接已经关闭
			c.ExitBuffChan <- struct{}{}
			c.TcpServer.GetConnMgr().Remove(c) //删除conn从ConnManager中
			//关闭该链接全部管道
			close(c.ExitBuffChan)
			close(c.msgChan)
		})
}

// GetTCPConnection 从当前连接获取原始的socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 获取当前连接ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// SendMsg  直接将Message数据发送数据给远程的TCP客户端
func (c *Connection) SendMsg(msgId uint32, data []byte) error {

	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		slog.Error("[cinx] Pack ", "error", err)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgChan <- msg

	return nil
}

/*
写消息Goroutine， 用户将数据发送给客户端
*/
func (c *Connection) StartWriter() {

	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			//有数据要写给客户端
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:, ", err, " Conn Writer exit")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				//有数据要写给客户端
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:, ", err, " Conn Writer exit")
					return
				}
			} else {
				break
				fmt.Println("msgBuffChan is Closed")
			}
		case <-c.ExitBuffChan:
			//conn已经关闭
			return
		}
	}
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {

	//将data封包，并且发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		slog.Error("[cinx] Pack ", "error", err)
		return errors.New("Pack error msg ")
	}

	//写回客户端
	c.msgBuffChan <- msg

	return nil
}

// SetProperty 设置链接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// GetProperty 获取链接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

// RemoveProperty 移除链接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
