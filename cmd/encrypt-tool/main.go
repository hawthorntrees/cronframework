package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/hawthorntrees/cronframework/framework/config"
	"github.com/hawthorntrees/cronframework/framework/utils"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

func main() {
	configPath := flag.String("conf", "config/app.yml", "配置文件路径")
	databases_default := flag.String("databases:default", "", "全局默认明文密码")
	databases_lists := flag.String("databases:lists", "", "指定库明文密码")

	flag.Parse()

	absConfigPath, err := filepath.Abs(*configPath)
	if err != nil {
		fmt.Printf("获取配置文件绝对路径失败: %v\n", err)
		os.Exit(1)
	}

	data, err := os.ReadFile(absConfigPath)
	if err != nil {
		fmt.Printf("读取配置文件失败: %v\n", err)
		os.Exit(1)
	}

	var cfg config.Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		os.Exit(1)
	}

	if len(cfg.Databases.DefaultConfig.SM4Key) != 16 {
		fmt.Println("错误: SM4密钥必须为16字节")
		os.Exit(1)
	}

	if *databases_default == "" && *databases_lists == "" {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Print("请选择要修改目标: 1-databases_default，2-databases_lists :")
		scanner.Scan()
		databases_type := scanner.Text()
		if databases_type == "1" {
			fmt.Print("请输入明文密码 :")
			scanner.Scan()
			*databases_default = scanner.Text()
			if *databases_default == "" {
				fmt.Println("密码不能为空")
				os.Exit(1)
			}
			encrypted, err := utils.SM4Encrypt([]byte(cfg.Databases.DefaultConfig.SM4Key), *databases_default)
			if err != nil {
				fmt.Printf("加密失败: %v\n", err)
				os.Exit(1)
			}
			cfg.Databases.DefaultConfig.Password = encrypted

			fmt.Printf("加密成功: %s\n", encrypted)

		} else if databases_type == "2" {
			fmt.Print("请输入要修改的数据源名称 :")
			scanner.Scan()
			list_name := scanner.Text()
			db, ok := cfg.Databases.ListsConfig[list_name]
			if !ok {
				fmt.Println("不存在该数据源名称")
				os.Exit(1)
			} else {
				fmt.Print("请输入要修改的数据源的类型，1-default,2-primary,3-standby :")
				scanner.Scan()
				t := scanner.Text()
				if t != "1" && t != "2" && t != "3" {
					fmt.Println("请按照提示正确输入")
					os.Exit(1)
				}
				fmt.Print("请输入明文密码:")
				scanner.Scan()
				p := scanner.Text()
				if p == "" {
					fmt.Println("密码不能为空")
					os.Exit(1)
				}
				encrypted, err := utils.SM4Encrypt([]byte(cfg.Databases.DefaultConfig.SM4Key), p)
				if err != nil {
					fmt.Printf("加密失败: %v\n", err)
					os.Exit(1)
				}
				switch t {
				case "1":
					db.DefaultConfig.Password = encrypted
				case "2":
					db.PrimaryConfig.Password = encrypted
				case "3":
					db.StandbyConfig.Password = encrypted
				}
				fmt.Printf("加密成功: %s\n", encrypted)

			}

		} else {
			fmt.Println("请按照提示正确输入")
			os.Exit(1)
		}
	}

	newData, err := yaml.Marshal(cfg)
	if err != nil {
		fmt.Printf("序列化配置文件失败: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(absConfigPath, newData, 0644); err != nil {
		fmt.Printf("更新配置文件失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("配置文件已更新: %s\n", absConfigPath)
}
