package main

import (
	"fmt"
	"os"
	"strconv"

	"tuyou/pkg/config"
	"tuyou/pkg/jwt"
)

func main() {
	// 默认用户：memberID=1
	memberID := uint(1)
	username := "测试会员"

	// 支持命令行参数：go run ./cmd/gen_token [memberID] [username]
	if len(os.Args) > 1 {
		if id, err := strconv.ParseUint(os.Args[1], 10, 64); err == nil {
			memberID = uint(id)
		} else {
			fmt.Printf("警告: 无法解析 memberID=%q，使用默认值 1\n", os.Args[1])
		}
	}
	if len(os.Args) > 2 {
		username = os.Args[2]
	}

	cfg, err := config.Load("")
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}
	fmt.Println("使用的密钥:", cfg.JWT.Secret)
	fmt.Println("配置过期时间:", cfg.JWT.ExpireHours, "小时")
	fmt.Printf("生成用户: ID=%d 用户名=%s\n", memberID, username)

	token, err := jwt.GenerateToken(memberID, username)
	if err != nil {
		panic(err)
	}
	fmt.Println("\nToken:", token)
}
