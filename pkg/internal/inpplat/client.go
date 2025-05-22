package inpplat

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/mitchellh/mapstructure"
	"resty.dev/v3"
)

type Client interface {
	CreateTask(map[string]string) (int, error)
	CloseTask(int) error
	SendHeartbeat(int) error
	BindRules([]Rule) error
	UnbindRules([]Rule) error
}

const (
	CREATETASKROUTER  = "/api/task/create"
	CLOSETASKROUTER   = "/api/task/close"
	HEARTBEATROUTER   = "/api/task/heartbeat"
	BINDRULESROUTER   = "/api/rules/bind"
	UNBINDRULESROUTER = "/api/rules/unbind"
)

// TODO: 约定传参
type CreateTaskParams struct {
	Name string `mapstructure:"name"`
	VID  string `mapstructure:"vid"`
}

// TODO: 约定返回接口
type CreateTaskResult struct {
	Name string `json:"name"`
	Id   int    `json:"id"`
}

func CompleteTaskParams(taskParams map[string]string) CreateTaskParams {
	var params CreateTaskParams
	decoder, _ := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &params,
		TagName: "mapstructure",
	})
	if err := decoder.Decode(taskParams); err != nil {
		slog.Error("Failed to complete TaskParams from map", "errorMsg", err)
	}
	return params
}

const (
	HTTP_TIMEOUT = 5 //seconds
)

type restProxyClient struct {
	client *resty.Client
}

func NewClient(addr, port, baseUrl, authToken string) Client {
	if !strings.HasPrefix(baseUrl, "/") {
		baseUrl = "/" + baseUrl
	}
	baseFullUrl := "http://" + addr + ":" + port + baseUrl

	client := resty.New()
	client.SetBaseURL(baseFullUrl).
		SetTimeout(HTTP_TIMEOUT * time.Second).
		SetHeaders(map[string]string{"Content-Type": "application/json"})

	return &restProxyClient{
		client: client,
	}
}

func NewMockClient() Client {
	return NewClient(MOCK_ADDRESS, MOCK_PORT, MOCK_API_BASE_URL, "")
}

func (p *restProxyClient) CreateTask(taskParams map[string]string) (int, error) {

	var result CreateTaskResult

	resp, err := p.client.R().
		SetBody(CompleteTaskParams(taskParams)).
		SetResult(&result).
		Post(CREATETASKROUTER)
	if err != nil {
		return -1, err
	}

	// TODO: 需要约定HTTP方法的响应码和响应消息
	if resp.StatusCode() != http.StatusOK {
		slog.Error("CreateTask failed, Recvied: %s", "Response Message", resp.String())
		return -1, err
	}

	return result.Id, err
}

func (p *restProxyClient) CloseTask(id int) error {

	resp, err := p.client.R().
		SetBody(map[string]int{"id": id}).
		Post(CLOSETASKROUTER)
	if err != nil {
		return err
	}
	// TODO: 需要约定HTTP方法的响应码和响应消息
	if resp.StatusCode() != http.StatusOK {
		slog.Error("CloseTask failed, Recvied: %s", "Response Message", resp.String())
		return err
	}

	return err
}

func (p *restProxyClient) SendHeartbeat(id int) error {
	resp, err := p.client.R().
		SetBody(map[string]int{"id": id}).
		Post(HEARTBEATROUTER)
	if err != nil {
		return err
	}
	// TODO: 需要约定HTTP方法的响应码和响应消息
	if resp.StatusCode() != http.StatusOK {
		slog.Error("SendHeartbeat failed, Recvied: %s", "Response Message", resp.String())
		return err
	}

	return err
}

func (p *restProxyClient) BindRules(rules []Rule) error {
	resp, err := p.client.R().
		SetBody(rules).
		Post(BINDRULESROUTER)
	if err != nil {
		return err
	}
	// TODO: 需要约定HTTP方法的响应码和响应消息
	if resp.StatusCode() != http.StatusOK {
		slog.Error("BindRules failed, Recvied: %s", "Response Message", resp.String())
		return err
	}

	return err
}

func (p *restProxyClient) UnbindRules(rules []Rule) error {
	resp, err := p.client.R().
		SetBody(rules).
		Post(UNBINDRULESROUTER)
	if err != nil {
		return err
	}
	// TODO: 需要约定HTTP方法的响应码和响应消息
	if resp.StatusCode() != http.StatusOK {
		slog.Error("UnbindRules failed, Recvied: %s", "Response Message", resp.String())
		return err
	}

	return err
}
