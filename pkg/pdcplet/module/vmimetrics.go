package module

import (
	"context"
	"fmt"
	"log/slog"
	"pdcplet/pkg/config"
	"pdcplet/pkg/internal/inpplat"
	"sync"
	"time"

	"resty.dev/v3"
)

const VMI_METRICS_NAME = "VmiMetrics"

type vmiMetricsModule struct {
	name         string
	inpplatproxy inpplat.Client
	restclient   *resty.Client
	cycle        time.Duration // 采集周期
}

// var MOCK_SERVER = RestClientConfig{
// 	Addr:      "192.168.153.142",
// 	Port:      "5888",
// 	BaseUrl:   "/pdcpserver/mock/",
// 	AuthToken: "",
// }

const (
	HTTP_TIMEOUT  = 5 * time.Second
	DEFAULT_CYCLE = 5 * time.Second
)

// func NewDefaultVmiMetricsModule() Module {
// 	return NewVmiMetricsModule(&MOCK_SERVER, DEFAULT_CYCLE)
// }

func NewVmiMetricsModule(params map[string]interface{}) (Module, error) {

	if len(params) == 0 {
		slog.Error("vmiMetricsModule params is nil or empty")
		return nil, fmt.Errorf("vmiMetricsModule params is nil or empty")
	}

	vmm := &vmiMetricsModule{
		name: VMI_METRICS_NAME,
	}

	if conns, ok := params["connections"]; !ok || len(conns.([]map[string]interface{})) == 0 {
		slog.Error("vmiMetricsModule connections is nil or empty")
		return nil, fmt.Errorf("vmiMetricsModule connections is nil or empty")
	}

	for _, conn := range params["connections"].([]map[string]interface{}) {
		if conn["name"] == config.INPPLAT_CONNECTION_NAME {
			proxy, err := NewInpProxy(conn)
			if err != nil {
				slog.Error("NewInpProxy failed", "errMsg", err)
				return nil, fmt.Errorf("NewInpProxy failed: %w", err)
			}
			vmm.inpplatproxy = proxy
		}

		if conn["name"] == config.PDCPSERVER_CONNECTION_NAME {
			if restConfig, ok := conn["httpOverTcpIp"].(map[string]interface{}); ok {
				vmm.restclient = resty.New().
					SetBaseURL(fmt.Sprintf("http://%s:%d%s", restConfig["host"], restConfig["port"], restConfig["urlPrefix"])).
					SetTimeout(HTTP_TIMEOUT).
					SetHeaders(map[string]string{"Content-Type": "application/json"})
			} else {
				slog.Error("PDCPSERVER_CONNECTION_NAME config is invalid")
				return nil, fmt.Errorf("PDCPSERVER_CONNECTION_NAME config is invalid")
			}
		}
	}
	if vmm.inpplatproxy == nil {
		vmm.inpplatproxy = inpplat.NewMockClient()
	}

	var cycle time.Duration = DEFAULT_CYCLE
	if cycleParam, ok := params["retriveMetricsCycle"]; ok {
		cycle = convertToTimeDuration(cycleParam.(string), DEFAULT_CYCLE)
	}
	vmm.cycle = cycle

	return vmm, nil
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
}

func (v *vmiMetricsModule) CollectForwardMetrics() error {
	metrics, err := v.inpplatproxy.GetAllForwardMetricsGroupByTask()
	if err != nil {
		slog.Error("GetAllForwardMetricsGroupByTask failed", "errMsg", err)
		return err
	}

	slog.Debug("GetAllForwardMetricsGroupByTask successfully", "metrics_len", len(metrics))

	// TODO: 需要处理返回结果
	_, err = v.restclient.R().
		SetBody(metrics).
		SetResult(nil).
		Post("/metrics/")
	if err != nil {
		slog.Error("Post metrics failed", "errMsg", err)
	}
	return nil
}
