package framework

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"pdcplet/pkg/module"
	_ "pdcplet/pkg/module/vmiproxy"
	"strings"
	"sync"
	"syscall"
)

type Framework interface {
	Init()
	Start()
}

// Framework 组合多个功能板块
type framework struct {
	modules     []module.Module
	construtors map[string]func() module.Module
	wg          sync.WaitGroup
}

// NewFramework 创建框架
func NewFramework(moduleList []string) Framework {

	f := &framework{
		modules:     make([]module.Module, 0),
		construtors: make(map[string]func() module.Module, 0),
	}

	module.GetAllModules()
	registry := module.Registry
	if len(registry) == 0 {
		slog.Error("No module in Registry")
		return nil
	}
	ms := []string{}
	for name, _ := range registry {
		ms = append(ms, name)
	}
	slog.Debug("List all modules", "list", fmt.Sprintf("[%s]", strings.Join(ms, ",")))

	for _, moduleName := range moduleList {
		ModuleName := strings.ToLower(moduleName)
		if _, exists := registry[ModuleName]; exists {
			f.construtors[ModuleName] = registry[ModuleName]
		} else {
			slog.Warn("Can't find app", "ModuleName", ModuleName)
		}
	}

	return f
}

// Init 调用App的构造函数创建App
func (f *framework) Init() {
	for _, constructor := range f.construtors {
		f.modules = append(f.modules, constructor())
	}
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
		go module.Run(ctx, &f.wg)
	}

	f.wg.Wait()
	slog.Info("All modules stoped, exit...")
}
