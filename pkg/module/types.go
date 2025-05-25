package module

import (
	"context"
	"sync"
)

// Module接口的构造函数的抽象类型
type ModuleConstructor func(params ...interface{}) (Module, error)

// Module 接口
type Module interface {
	Name() string
	Run(ctx context.Context, wg *sync.WaitGroup)
}

// ConfigModule 基于选项模式(option)，Module实例配置函数的抽象类型
type ConfigModule func(m interface{})
