package global

import (
	"cinx/zinx/ziface"
)

var TcpServer ziface.IServer //当前Zinx的全局Server对象
var Host string              //当前服务器主机IP
var TcpPort int              //当前服务器主机监听端口号
var Name string              //当前服务器名称
var Version string           //当前Zinx版本号
var MaxPacketSize uint32     //都需数据包的最大值
var MaxConn int              //当前服务器主机允许的最大链接个数
var WorkerPoolSize uint32    //业务工作Worker池的数量
var MaxWorkerTaskLen uint32  //业务工作Worker对应负责的任务队列最大任务存储数量
var MaxMsgChanLen uint32

func init() {
	Name = "ZinxServerApp"
	Version = "tcp4"
	TcpPort = 7777
	Host = "0.0.0.0"
	MaxConn = 12000
	MaxPacketSize = 409600
	WorkerPoolSize = 10
	MaxWorkerTaskLen = 1024
	MaxMsgChanLen = 1024
}
