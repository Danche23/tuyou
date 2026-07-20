package main

import (
	"flag"
	"fmt"

	"tuyou/internal/app"
)

func main() {
	// 定义命令行参数
	configPath := flag.String("c", "", "配置文件路径")
	flag.Parse()

	// 创建应用实例
	application := app.NewApp()

	// 初始化应用
	if err := application.Initialize(*configPath); err != nil {
		panic(fmt.Sprintf("应用初始化失败: %v", err))
	}

	// 运行应用
	application.Run()
}
