package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"video-srt-command/apps"
	"video-srt-command/apps/tools"
)

//定义配置文件
const CONFIG = "config.ini"

func main() {

	//致命错误捕获
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("")
			log.Printf("main.go:19:运行过程中出现致命错误:\n%v", err)

			time.Sleep(time.Second * 3)
		}
	}()

	//初始化
	if len(os.Args) < 2 {
		os.Args = append(os.Args , "")
	}

	var video string

	//设置命令行参数
	flag.StringVar(&video, "f", "", "输入需要处理的视频文件.")
	flag.Parse()

	if video == "" && os.Args[1] != "" && os.Args[1] != "-f" {
		video = os.Args[1]
	}

	//获取应用
	app := apps.NewApp(CONFIG)

	appDir := tools.GetAppRootDir()

	//初始化应用
	app.Init(appDir)

	//调起应用
	app.Run(tools.WinDir(video))

	//延迟退出
	time.Sleep(time.Second * 1)
}
