package framework

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"pdcplet/pkg/module"
	"sync"
	"syscall"
)

type Framework interface {
	Start()
}

// Framework 组合多个功能板块
type framework struct {
	modules []module.Module
	// construtors map[string]func() module.Module
	wg sync.WaitGroup
}

// NewFramework 创建框架
func NewFramework(moduleList []string) Framework {
	f := &framework{
		modules: make([]module.Module, 0),
	}

	for _, name := range moduleList {
		moduleInstace, err := module.CreateModule(name, nil)
		if err != nil {
			slog.Error("Create module error", "name", name, "error", err)
			return nil
		}
		if moduleInstace == nil {
			slog.Error("Create module error", "name", name, "error", "module is nil")
			return nil
		}
		f.modules = append(f.modules, moduleInstace)
	}

	return f
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
