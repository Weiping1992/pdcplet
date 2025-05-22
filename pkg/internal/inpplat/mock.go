package inpplat

const (
	MOCK_ADDRESS      = "192.168.153.141"
	MOCK_PORT         = "5777"
	MOCK_API_BASE_URL = "/mock/"
)

type Rule struct{}

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

type BaseMetric struct {
	Sent    int64 `json:"sent"`
	Dropped int64 `json:"dropped"`
	Avgbps  int64 `json:"avgbps"`
	Avgpps  int64 `json:"avgpps"`
	Realbps int64 `json:"realbps"`
	Realpps int64 `json:"realpps"`
}

type NicMetric struct {
	Vid int64  `json:"vid"`
	Mac string `json:"mac"`
	BaseMetric
}

type ForwardMetrics struct {
	TaskId int         `json:"task_id"`
	Total  BaseMetric  `json:"total"`
	Nics   []NicMetric `json:"nics"`
}
