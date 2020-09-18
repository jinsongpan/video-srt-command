package app

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"video-srt-command/app/config"
	"video-srt-command/app/ffmpeg"
	"video-srt-command/app/speechAPI/aliyun/asr"
	"video-srt-command/app/speechAPI/aliyun/oss"
	"video-srt-command/app/speechAPI/baidu/trans"
	"video-srt-command/app/tool"
)

//主应用
type VideoSrt struct {
	Ffmpeg      ffmpeg.Ffmpeg
	AliyunOSS   oss.AliyunOSS          //阿里OSS
	AliyunCloud asr.AliyunCloud        //阿里语音识别引擎
	TransConfig trans.BaiduTransConfig //百度翻译配置

	//AppDir         string //应用根目录
	IsCleanLocalTempFile bool
	IsCleanOSSTempFile   bool
	AutoBlock            bool   //自动分段处理
	InputLang            int    //输入语言种类
	OutputLang           int    //输出语言种类
	IsTranslate          bool   //是否翻译
	IsBilingual          bool   //输出双语字幕，默认双语字幕
	OutputType           string //输出文件类型
	OutputEncode         string //输出文件编码
	//MaxConcurrency       int    // 最大并发数
	TempDir string //临时文件目录

}

//获取应用配置
func NewApp(cfg string) *VideoSrt {
	app := ReadConfig(cfg)

	return app
}

//读取config.ini配置
func ReadConfig(cfg string) *VideoSrt {
	if file, err := config.LoadConfigFile(cfg, "."); err != nil {
		panic(err)
	} else {
		appconfig := &VideoSrt{}

		//AliyunOSS
		appconfig.AliyunOSS.Endpoint = file.GetMust("Aliyun.endpoint", "")
		appconfig.AliyunOSS.AccessKeyId = file.GetMust("Aliyun.accessKeyId", "")
		appconfig.AliyunOSS.AccessKeySecret = file.GetMust("Aliyun.accessKeySecret", "")
		appconfig.AliyunOSS.BucketName = file.GetMust("Aliyun.bucketName", "")
		appconfig.AliyunOSS.BucketDomain = file.GetMust("Aliyun.bucketDomain", "")

		//AliyunCloud
		appconfig.AliyunCloud.AppKey = file.GetMust("Aliyun.appKey", "")
		appconfig.AliyunCloud.AccessKeyId = file.GetMust("Aliyun.accessKeyId", "")
		appconfig.AliyunCloud.AccessKeySecret = file.GetMust("Aliyun.accessKeySecret", "")

		//BaiduTranslate
		appconfig.TransConfig.AppID = file.GetMust("BaiduTranslate.appID", "")
		appconfig.TransConfig.SecretKey = file.GetMust("BaiduTranslate.secretKey", "")
		appconfig.TransConfig.AccountType = file.GetMust("BaiduTranslate.accountType", "")

		//srt
		appconfig.TempDir = file.GetMust("Sys.tempDir", "")
		appconfig.OutputEncode = file.GetMust("Srt.outputEncode", "")
		appconfig.OutputType = file.GetMust("Srt.outputType", "")
		//appconfig.MaxConcurrency = file.GetIntMust("Srt.maxConcurrency", 1)

		return appconfig
	}
}

//应用初始化
func (app *VideoSrt) Init(autoBlock bool, inputlang int, outputlang int, isTranslate bool,
	isBilingual bool, isCleanLocalTempFile bool, isCleanOSSTempFile bool) {
	app.AutoBlock = autoBlock
	app.InputLang = inputlang
	app.OutputLang = outputlang
	app.IsTranslate = isTranslate
	app.IsBilingual = isBilingual
	app.IsCleanLocalTempFile = isCleanLocalTempFile
	app.IsCleanOSSTempFile = isCleanOSSTempFile

}

//应用运行
func (app *VideoSrt) Run(media string) {
	//支持的媒体文件类型
	videoTypes := []string{".mp4", ".mpeg", ".mkv", ".wmv", ".avi", ".m4v", ".mov", ".flv", ".rmvb", ".3gp", ".f4v"}
	audioTypes := []string{".mp3", ".wav", ".aac", ".wma", ".flac", ".m4a"}
	var tmpAudio string                                                // 待上传的音频文件
	mediaName := strings.TrimSuffix(path.Base(media), path.Ext(media)) // 获取文件名称
	tmpAudioFile := mediaName + ".mp3"

	if !tool.DirExists(app.TempDir) {
		//创建目录
		if err := tool.CreateDir(app.TempDir, false); err != nil {
			panic(err)
		}
	}
	tmpAudio = app.TempDir + "/" + tmpAudioFile

	if media == "" {
		panic("app.go: 输入需要识别的媒体文件!")
	}

	//校验媒体文件
	if tool.VaildFile(media) != true {
		panic("app.go: 视频文件无效!")
	}
	//校验临时文件是否冲突
	if tool.VaildFile(tmpAudio) != false {
		panic("tempDir目录下存在与输入媒体文件同名称的音频文件!")
	}

	//判断媒体文件是视频还是音频
	mediaType := path.Ext(media)
	log.Println("step1：根据输入媒体文件获得对应的音频文件...")

	if tool.IsContain(videoTypes, mediaType) {
		//分离出视频中的音频文件
		tool.ExtractVideoAudio(media, tmpAudio)
	} else if tool.IsContain(audioTypes, mediaType) {
		if err := ffmpeg.AudioToMP3(media, tmpAudio); err != nil {
			panic(err)
		}
	} else {
		panic("输入媒体文件格式不支持；支持格式：.mp4 , .mpeg , .mkv , .wmv , .avi , .m4v , .mov , .flv , .rmvb , .3gp , .f4v，" +
			".mp3 , .wav , .aac , .wma , .flac , .m4a")
	}

	//删除本地临时音频文件
	if app.IsCleanLocalTempFile {
		defer func() {
			if remove := os.Remove(tmpAudio); remove != nil {
				log.Println("临时音频清理失败，建议手动删除", tmpAudio)
				panic(remove)
			} else {
				log.Println("本地临时音频文件清理成功！")
			}
		}()
	}

	log.Printf("step2: 上传规范化后的音频文件:%v...", tmpAudio)
	//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 0

	//上传音频至OSS
	OSSTempFile := oss.UploadAudioToCloud(app.AliyunOSS, tmpAudio)
	//fmt.Println("app.go:157:OSSTempFile", OSSTempFile)

	//获取完整OSS链接
	filelink := app.AliyunOSS.GetObjectFileUrl(OSSTempFile)
	//fmt.Println("app.go:160:filelink", filelink)
	//filelink := "http://asr-test-srt.oss-cn-shenzhen.aliyuncs.com/2020/9/9/test.mp3"

	log.Printf("step3: %v 上传成功 , 开始识别 ...", tmpAudio)
	//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 1

	//清理OSS临时音频文件
	if app.IsCleanOSSTempFile {
		defer func() {
			if err := oss.DelOSSTempFile(app.AliyunOSS, OSSTempFile); err != nil {
				log.Println("AliyunOSS临时音频清理失败，建议手动删除", filelink)
			} else {
				log.Println("AliyunOSS临时音频清理成功！")
			}
		}()
	}

	//阿里云录音文件识别
	AudioResult := asr.AliyunAudioRecognition(filelink, app.AliyunCloud, app.AutoBlock)

	log.Printf("step4: %v 识别成功 , 开始制作字幕 ...", tmpAudio)

	//字幕翻译
	if app.IsTranslate {
		log.Println("step4-5: 开始翻译字幕...")
		//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 2
		if err := app.AliyunTransResult(media, AudioResult, app.InputLang, app.OutputLang); err != nil {
			log.Printf("%v 字幕翻译失败："+err.Error(), media)
			log.Printf("已强制关闭翻译，仅输出原始 %v 字幕文件", media)
			app.IsTranslate = false
		}
	}

	//输出字幕文件
	log.Printf("step5：%v 字幕文件生成中...", media)
	//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 3
	asr.AliyunRecResultGenText(media, AudioResult, app.OutputEncode, app.OutputType, app.IsTranslate, app.IsBilingual)

	log.Println("字幕制作完成！")

}

//翻译阿里云识别文本
func (app *VideoSrt) AliyunTransResult(mediaName string, AudioResult map[int64][]*asr.AliyunAudioRecResult, InputLang int,
	outputLang int) error {

	//计算总任务行数
	var (
		textLines            = 0 //总处理行数
		transProcess float64 = 0 //翻译进度（%）
		index                = 0 //原文文本序号
	)

	for _, result := range AudioResult {
		textLines += len(result)
	}

	//执行翻译任务
	for _, result := range AudioResult {

		for _, data := range result {
			transResult, err := app.TransConfig.RunTranslate(data.Text, mediaName, InputLang, outputLang)
			if err != nil {
				return err
				//panic(err) //终止翻译
			}
			data.TranslateText = strings.TrimSpace(transResult.TransResultDst) //译文文本

			index++
			//翻译进程
			tempProcess := (float64(index) / float64(textLines)) * 100
			if (tempProcess - transProcess) > 25 {
				//输出比例
				log.Println("字幕翻译已处理："+fmt.Sprintf("%.2f", tempProcess)+"%", mediaName)
				transProcess = tempProcess
			}
		}
	}

	return nil
}
