// mock-platform: 模拟地面健康管理平台的接收端
// 监听 8188 端口，接受 ground-reporter 的三类 POST 请求，
// 将收到的 JSON 格式化打印到控制台，并返回成功响应。
//
// 用法:
//   go run main.go
//   go run main.go -port 8188
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var knownPaths = map[string]string{
	"/gate/METRO-PHM/api/faultRecordsSubsystem/saveRecord":           "6.1 预警/报警写入",
	"/gate/METRO-SELFCHECK-SUBSYSTEM/api/faultRecordsSubsystem/saveStatus": "6.6 运行状态心跳",
	"/gate/METRO-PHM/api/devices/status/train/saveOrUpdate":          "6.7 寿命状态写入",
}

func main() {
	port := flag.Int("port", 8188, "监听端口")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", handleRequest)

	addr := fmt.Sprintf("0.0.0.0:%d", *port)
	log.Printf("mock-platform 启动，监听 %s", addr)
	log.Printf("已注册接口：")
	for path, name := range knownPaths {
		log.Printf("  POST %s  →  %s", path, name)
	}

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("启动失败: %v", err)
	}
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	ts := time.Now().Format("15:04:05.000")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("[%s] 读取 body 失败: %v", ts, err)
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	name, known := knownPaths[r.URL.Path]
	if !known {
		name = "未知接口"
	}

	// 格式化 JSON
	var pretty bytes.Buffer
	if err := json.Indent(&pretty, body, "", "  "); err != nil {
		pretty.Write(body) // 格式化失败就原样输出
	}

	apiKey := r.Header.Get("X-Api-Key")
	if apiKey == "" {
		apiKey = "(未设置)"
	}

	sep := "─────────────────────────────────────────────────"
	fmt.Printf("\n%s\n", sep)
	fmt.Printf("[%s]  %s\n", ts, name)
	fmt.Printf("路径: %s %s\n", r.Method, r.URL.Path)
	fmt.Printf("X-Api-Key: %s\n", apiKey)
	fmt.Printf("Body:\n%s\n", pretty.String())
	fmt.Printf("%s\n", sep)

	// 返回平台标准成功响应
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(`{"code":200,"message":"操作成功","entity":null}`))
}
