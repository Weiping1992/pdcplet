package module

import (
	"context"
	"strings"
	"sync"
)

// Registry App的注册表，记录所有App的Name和构造函数
var (
	InitFunc = make([]func(), 0)
	Registry = make(map[string]func() Module)
	mu       sync.RWMutex
)

// Module 接口
type Module interface {
	Name() string
	Run(ctx context.Context, wg *sync.WaitGroup)
}

func RegisterConstructor(name string, constructor func() Module) {
	mu.Lock()
	defer mu.Unlock()
	Registry[strings.ToLower(name)] = constructor
}

func RegisterInit(init func()) {
	mu.Lock()
	defer mu.Unlock()
	InitFunc = append(InitFunc, init)
}

func GetAllModules() {
	for _, init := range InitFunc {
		init()
	}
}
