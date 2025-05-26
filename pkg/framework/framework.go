package framework

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"pdcplet/pkg/config"
	"pdcplet/pkg/module"
	"sync"
	"syscall"
)

type Framework interface {
	Start()
	AddModule(name string, params map[string]interface{}, connections []config.Connection) error
}

// Framework 组合多个功能板块
type framework struct {
	modules []module.Module
	// construtors map[string]func() module.Module
	wg sync.WaitGroup
}

// NewFramework 创建框架
func NewFramework() Framework {
	f := &framework{
		modules: make([]module.Module, 0),
	}
	return f
}

func (f *framework) AddModule(name string, params map[string]interface{}, connections []config.Connection) error {

	params["connections"] = make([]map[string]interface{}, 0, len(connections))
	for _, conn := range connections {
		params["connections"] = append(params["connections"].([]map[string]interface{}), conn.ConvertToMap())
	}

	moduleInstace, err := module.CreateModule(name, params)
	if err != nil {
		slog.Error("Create module error", "name", name, "error", err)
		return err
	}
	if moduleInstace == nil {
		slog.Error("Create module error", "name", name, "error", "module is nil")
		return fmt.Errorf("Create module %s error, err: module is ni", name)
	}
	f.modules = append(f.modules, moduleInstace)
	slog.Info("Module added", "name", name)

	return nil
}

func (f *framework) Start() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 信号处理
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// 等待终止信号
	go func() {
		// 等待终止信号
		sig := <-signalChan
		slog.Info("Recvied Stop Signal", "signal", sig)
		cancel()
	}()

	// 启动App
	for _, module := range f.modules {
		f.wg.Add(1)
		slog.Info("Starting Module", "ModuleName", module.Name())
		go module.Run(ctx, &f.wg)
	}

	f.wg.Wait()
	slog.Info("All modules stoped, exit...")
}
