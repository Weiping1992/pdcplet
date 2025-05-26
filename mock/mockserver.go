package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"time"
)

// 自定义响应结构体（网页5方案增强）
type CreateResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	Code    int    `json:"code"`
	Name    string `json:"name"`
	Id      int    `json:"id"`
}

// TODO: 约定传参
type CreateTaskParams struct {
	Name string `mapstructure:"name"`
	VID  string `mapstructure:"vid"`
}

var randGen *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func main() {
	// 注册全局请求处理器（网页1/8方案结合）
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 打印完整请求信息（网页2方案优化）
		dump, _ := httputil.DumpRequest(r, true)
		fmt.Printf("\n=== 请求日志 ===\n%s\n", dump)

		// 特殊路由处理（网页3/5方案结合）
		if r.URL.Path == "/mock/api/task/create" {
			handleCreate(w, r)
			return
		} else if r.URL.Path == "/mock/api/metrics/" {
			handleMetrics(w, r)
			return
		}

		// 默认响应（网页4方案）
		w.Write([]byte("请求已接收"))
	})

	// 启动服务（网页3/4最佳实践）
	addr := "0.0.0.0:5777"
	fmt.Printf("服务端启动，监听地址：%s\n", addr)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}

// 特殊路由处理器（网页5/8方案扩展）
func handleCreate(w http.ResponseWriter, r *http.Request) {

	var req CreateTaskParams
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		fmt.Println("解析请求结构体失败, body")
		return
	}

	response := CreateResponse{
		Status:  "success",
		Message: "资源创建成功",
		Code:    200,
		Name:    req.Name,
		Id:      randGen.Intn(50) + 1,
	}
	fmt.Printf("Response: taskId %d\n", response.Id)
	// 设置响应头（网页5关键实践）
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// 序列化JSON（网页5错误处理优化）
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "内部服务器错误", http.StatusInternalServerError)
	}
}

func handleMetrics(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received metrics request")
	return
}
