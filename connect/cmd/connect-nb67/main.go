package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redpanda-data/connect/v4/public/service"
)

// 初始化函数 - 注册自定义处理器
func init() {
	// 注册 nb67_parser 处理器
	err := service.RegisterProcessor(
		"nb67_parser",
		service.NewConfigSpec().
			Summary("NB67 二进制协议框架解析器").
			Description(
				"使用Kaitai Struct生成的Go代码解析NB67二进制浮车空调数据帧。\n"+
					"输入：原始二进制消息\n"+
					"输出：完整JSON对象，包含180+字段、新增车站信息、故障标志等\n"+
					"包含字段：头部信息、时间戳、温度/湿度/压力传感器、故障诊断、新增车站信息(452-460偏移)\n",
			).
			Field(
				service.NewInt64Field("log_sample_every").
					Description("每处理N条消息输出日志采样（0=不采样）").
					Default(100),
			),
		func(ctx context.Context, conf *service.ParsedConfig) (service.Processor, error) {
			return NewNB67Processor(conf)
		},
	)
	if err != nil {
		panic(err)
	}
}

func main() {
	// 使用Redpanda Connect官方的CLI启动器
	// 这个会处理：
	// - 命令行参数解析
	// - 配置文件加载
	// - 优雅关闭信号处理
	// - Prometheus metrics导出
	// - 日志配置
	if err := service.Run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
