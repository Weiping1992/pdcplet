package vmimetrics

import (
	"context"
	"log/slog"
	"pdcplet/pkg/internal/inpplat"
	"pdcplet/pkg/module"
	"strings"
	"sync"
	"time"

	"resty.dev/v3"
)

const NAME = "VmiMetrics"

type vmiMetricsModule struct {
	name       string
	inpclient  inpplat.Client
	restclient *resty.Client
	cycle      time.Duration // 采集周期
}

func init() {
	module.RegisterInit(func() {
		module.RegisterConstructor(NAME, NewDefaultVmiMetricsModule)
	})
}

type RestClientConfig struct {
	Addr      string
	Port      string
	BaseUrl   string
	AuthToken string
}

var MOCK_SERVER = RestClientConfig{
	Addr:      "192.168.153.142",
	Port:      "5888",
	BaseUrl:   "/pdcpserver/mock/",
	AuthToken: "",
}

const (
	HTTP_TIMEOUT  = 5 * time.Second
	DEFAULT_CYCLE = 5 * time.Second
)

func NewDefaultVmiMetricsModule() module.Module {
	return NewVmiMetricsModule(&MOCK_SERVER, DEFAULT_CYCLE)
}

func NewVmiMetricsModule(config *RestClientConfig, cycle time.Duration) module.Module {
	// 1: restclient
	if config == nil {
		config = &MOCK_SERVER
	}
	if !strings.HasPrefix(config.BaseUrl, "/") {
		config.BaseUrl = "/" + config.BaseUrl
	}
	baseFullUrl := "http://" + config.Addr + ":" + config.Port + config.BaseUrl
	client := resty.New()
	client.SetBaseURL(baseFullUrl).
		SetTimeout(HTTP_TIMEOUT).
		SetHeaders(map[string]string{"Content-Type": "application/json"})

	// 2: inpclient
	inpclient := inpplat.NewMockClient()

	if cycle == 0 {
		cycle = DEFAULT_CYCLE
	}

	return &vmiMetricsModule{
		name:       NAME,
		inpclient:  inpclient,
		restclient: client,
		cycle:      cycle,
	}
}

func (v *vmiMetricsModule) Name() string {
	return v.name
}

func (v *vmiMetricsModule) Run(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(v.cycle)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			slog.Debug("vmiMetricsModule doJob when tick happens", "cycle", v.cycle)
			v.doJob()
		case <-ctx.Done():
			return
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

func (v *vmiMetricsModule) doJob() {
	err := v.CollectForwardMetrics()
	if err != nil {
		slog.Error("CollectForwardMetrics failed", "errMsg", err)
	}
	return
}

func (v *vmiMetricsModule) CollectForwardMetrics() error {
	metrics, err := v.inpclient.GetAllForwardMetricsGroupByTask()
	if err != nil {
		slog.Error("GetAllForwardMetricsGroupByTask failed", "errMsg", err)
		return err
	}

	slog.Debug("GetAllForwardMetricsGroupByTask successfully", "metrics_len", len(metrics))

	v.restclient.R().
		SetBody(metrics).
		SetResult(nil).
		Post("/metrics/")
	if err != nil {
		slog.Error("Post metrics failed", "errMsg", err)
	}
	return nil
}
