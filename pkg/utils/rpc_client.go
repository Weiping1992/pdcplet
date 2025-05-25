package utils

type RpcClient interface {
	// Connect 连接到RPC服务器
	Connect() error
	// Disconnect 断开与RPC服务器的连接
	Disconnect() error
	// Call 调用RPC方法
	Call(method string, args interface{}, reply interface{}) error
	// IsConnected 检查是否连接到RPC服务器
	IsConnected() bool
	// GetClient 获取RPC客户端
	GetClient() interface{}
	// GetClientType 获取RPC客户端类型
	GetClientType() string
	// GetClientName 获取RPC客户端名称
	GetClientName() string
	// GetClientVersion 获取RPC客户端版本
	GetClientVersion() string
	// GetClientID 获取RPC客户端ID
}
