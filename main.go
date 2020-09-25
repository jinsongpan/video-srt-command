package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"video-srt-command/app"
	"video-srt-command/app/tool"
)

//定义配置文件
const CONFIG = "config.ini"

func main() {

	//致命错误捕获
	defer func() {
		if err := recover(); err != nil {
			fmt.Println("")
			log.Printf("运行过程中出现未知错误: %v", err)

			time.Sleep(time.Second * 2)
		}
	}()

	//TODO 命令行支持的指令参数待补充
	//设置命令行接收参数
	var (
		media                string
		autoBlock            bool
		isCleanOSSTempFile   bool
		isCleanLocalTempFile bool
		inputlang            int
		otherParams          []string

		// 翻译相关
		isTranslate bool
		outputlang  int
		isBilingual bool
	)

	//命令行flag语法，三种：-flag; -flag=x; -flag x(只有非bool类型的flag可以)
	flag.StringVar(&media, "f", "", "输入需要处理的媒体文件")
	flag.BoolVar(&autoBlock, "block", true, "自动分段处理：true（开启） false（关闭）")
	flag.IntVar(&inputlang, "inlang", 0, "设定输入语言种类的整型代号，目前阿里语音支持语音：中文普通话(1)、英语(2)")
	flag.BoolVar(&isTranslate, "trans", false, "是否需要对字幕进行翻译(默认不翻译)")
	flag.IntVar(&outputlang, "outlang", 0, "设定目标翻译语言的整型代号，目前百度翻译支持27种语音：简体中文(1)、英文(2)、粤语(3)、文言文(4)、日语(5)、" +
		"韩语(6)、法语(7)、德语(8)、西班牙语(9)、泰语(10)、阿拉伯语(11)、俄语(12)、葡萄牙语(13)、意大利语(14)、希腊语(15)、荷兰语(16)、波兰语(17)、" +
		"保加利亚语(18)、丹麦语(19)、芬兰语(20)、捷克语(21)、罗马尼亚语(22)、斯洛文尼亚语(23)、瑞典语(24)、匈牙利语(25)、繁体中文(26)、越南语(27)")
	flag.BoolVar(&isBilingual, "biling", true, "是否需要生成双语字幕(默认生成双语字幕)")
	flag.BoolVar(&isCleanLocalTempFile, "cleanlocal", true, "是否删除本地的临时音频文件(默认删除)")
	flag.BoolVar(&isCleanOSSTempFile, "cleanoss", true, "是否删除OSS中的临时音频文件(默认删除)")

	// 解析flag参数
	flag.Parse()

	if media == "" {
		log.Printf("命令行中未按要求输入媒体文件!")
		os.Exit(1)
	}

	//如果出现未解析参数则报错
	otherParams = flag.Args()
	if len(otherParams) != 0 {
		log.Printf("命令行出现未解析参数：%v", otherParams)
		os.Exit(1)
	}

	//选择翻译后需要指定翻译的目标语种
	if isTranslate {
		if outputlang == 0 {
			log.Println("选择翻译后需要指定翻译的目标语种的整型代号,可供选择的语种有：中文(1),英文(2)")
			os.Exit(1)
		}
	}

	//获取应用
	newApp := app.NewApp(CONFIG)

	//初始化应用, 命令行指令优先级最高
	newApp.Init(autoBlock, inputlang, outputlang, isTranslate, isBilingual, isCleanLocalTempFile, isCleanOSSTempFile)

	//调起应用
	newApp.Run(tool.UnixDir(media))

	//延迟退出
	time.Sleep(time.Second * 1)
}
