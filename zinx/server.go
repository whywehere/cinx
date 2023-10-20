package zinx

import (
	"cinx/zinx/utils"
	"cinx/zinx/utils/global"
	"cinx/zinx/ziface"
	"fmt"
	"log/slog"
	"net"
	"time"
)

// Server iServer 接口实现，定义一个Server服务类
type Server struct {
	//服务器的名称
	ServerName string
	//tcp4 or other
	IPVersion string
	//服务绑定的IP地址
	IP string
	//服务绑定的端口
	Port int

	msgHandler ziface.IMsgHandle

	ConnMgr ziface.IConnManager

	OnConnStart func(conn ziface.IConnection)

	OnConnStop func(conn ziface.IConnection)
}

// Start 开启网络服务
func (s *Server) Start() {
	slog.Info("[cinx] Server start...")

	//开启一个go去做服务端listener业务
	go func() {
		s.msgHandler.StartWorkerPool()
		//1 获取一个TCP的Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			slog.Error("[cinx]  net.ResolveTCPAddr(): ", err, err)
			return
		}

		//2 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			slog.Error("[cinx] net.ListenTCP(): ", err, err)
			return
		}

		// == 已经监听成功 == //

		//3 启动server网络连接业务
		for {
			//3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				slog.Error("[cinx] net.AcceptTCP(): ", err, err)
				continue
			}

			if s.ConnMgr.Len() >= global.MaxConn {
				conn.Close()
				slog.Info("[cinx] s.ConnMgr.Len() >= global.MaxConn")
				continue
			}

			dalConn := NewConnection(s, conn, utils.RandomID(), s.msgHandler)
			go dalConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	slog.Info("[cinx] Server stop....", s.ServerName, s.ServerName)

	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	//TODO Server.Serve() 是否在启动服务的时候 还要处理其他的事情呢 可以在这里添加

	//阻塞,否则主Go退出， listener的go将会退出
	for {
		time.Sleep(10 * time.Second)
	}
}

// NewServer  创建一个服务器句柄
func NewServer() ziface.IServer {
	return &Server{
		ServerName: global.Name,
		IPVersion:  global.Version,
		IP:         global.Host,
		Port:       global.TcpPort,
		msgHandler: NewMsgHandle(),   //msgHandler 初始化
		ConnMgr:    NewConnManager(), //创建ConnManager
	}
}

// GetConnMgr 得到链接管理
func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}
func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.OnConnStart != nil {
		slog.Info("[cinx] OnConnStart is called...")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.OnConnStop != nil {
		slog.Info("[cinx] OnConnStop is called...")
		s.OnConnStop(conn)
	}
}
