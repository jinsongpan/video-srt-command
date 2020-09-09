package apps

import (
	"bytes"
	"github.com/buger/jsonparser"
	"os"
	"strconv"
	"strings"
	"video-srt-command/apps/SpeechAPIs/aliyun/asr"
	"video-srt-command/apps/SpeechAPIs/aliyun/oss"
	"video-srt-command/apps/configs"
	"video-srt-command/apps/ffmpeg"
	"video-srt-command/apps/logs"
	"video-srt-command/apps/tools"
)


//主应用
type VideoSrt struct {
	Ffmpeg ffmpeg.Ffmpeg
	AliyunOss oss.AliyunOss
	AliyunCloud asr.AliyunCloud //阿里语音识别引擎

	IntelligentBlock bool //智能分段处理
	TempDir string //临时文件目录
	AppDir string //应用根目录
}


//获取应用
func NewApp(cfg string) *VideoSrt {
	app := ReadConfig(cfg)

	return app
}


//读取配置
func ReadConfig (cfg string) *VideoSrt {
	if file, e := configs.LoadConfigFile(cfg , ".");e != nil  {
		panic(e);
	} else {
		appconfig := &VideoSrt{}
				
		//AliyunOSS
		appconfig.AliyunOss.Endpoint = file.GetMust("aliyunOss.endpoint" , "")
		appconfig.AliyunOss.AccessKeyId = file.GetMust("aliyunOss.accessKeyId" , "")
		appconfig.AliyunOss.AccessKeySecret = file.GetMust("aliyunOss.accessKeySecret" , "")
		appconfig.AliyunOss.BucketName = file.GetMust("aliyunOss.bucketName" , "")
		appconfig.AliyunOss.BucketDomain = file.GetMust("aliyunOss.bucketDomain" , "")

		//AliyunCloud
		appconfig.AliyunCloud.AccessKeyId = file.GetMust("aliyunCloud.accessKeyId" , "")
		appconfig.AliyunCloud.AccessKeySecret = file.GetMust("aliyunCloud.accessKeySecret" , "")
		appconfig.AliyunCloud.AppKey = file.GetMust("aliyunCloud.appKey" , "")


		appconfig.IntelligentBlock = file.GetBoolMust("srt.intelligent_block" , false)
		appconfig.TempDir = "srt-results/temp/audio"

		return appconfig
	}
}


//应用初始化
func (app *VideoSrt) Init(appDir string) {
	app.AppDir = appDir
}

//应用运行
func (app *VideoSrt) Run(video string) {
	if video == "" {
		panic("app.go:74: 输入需要识别的视频文件.")
	}

	//校验视频
	if tools.VaildVideo(video) != true {
		panic("app.go:74: 视频文件不存在.")
	}

	if !tools.DirExists(app.TempDir) {
		//创建目录
		if err := tools.CreateDir(app.TempDir , false); err != nil {
			panic(err)
		}
	}

	//tmpAudioFile := GetRandomCodeString(15) + ".mp3"
	tmpAudioFile := strings.Split(video, ".")[0] + ".mp3"

	tmpAudio := app.TempDir + "/" + tmpAudioFile
	logs.Log("提取音频文件 ...")

	//分离出视频中的音频文件
	ExtractVideoAudio(video , tmpAudio)

	logs.Log("上传音频文件 ...")
	//bufio.NewReader(os.Stdin).ReadBytes('\n')  //断点 0

	//上传音频至OSS
	filelink := UploadAudioToClound(app.AliyunOss , tmpAudio)

	//获取完整链接
	filelink = app.AliyunOss.GetObjectFileUrl(filelink)
	//fmt.Println("app.go:109:filelink", filelink)

	//filelink := "http://asr-test-srt.oss-cn-shenzhen.aliyuncs.com/2020/9/9/test.mp3"

	logs.Log("上传文件成功 , 识别中 ...")
	//bufio.NewReader(os.Stdin).ReadBytes('\n')  //断点 1

	//阿里云录音文件识别
	AudioResult := AliyunAudioRecognition(app.AliyunCloud, filelink , app.IntelligentBlock)

	logs.Log("文件识别成功 , 字幕处理中 ...")

	//输出字幕文件
	AliyunAudioResultMakeSubtitleFile(video , AudioResult)

	logs.Log("字幕制作完成！")

	//删除临时音频文件
	if remove := os.Remove(tmpAudio); remove != nil {
		panic(remove)
	}
}


//提取视频音频文件
func ExtractVideoAudio(video string , tmpAudio string) {
	if err := ffmpeg.ExtractAudio(video , tmpAudio); err != nil {
		panic(err)
	}
}


//上传音频至OSS
func UploadAudioToClound(target oss.AliyunOss , audioFile string) string {
	name := ""
	//提取文件名称
	if fileInfo, e := os.Stat(audioFile);e != nil {
		panic(e)
	} else {
		name = fileInfo.Name()
	}

	//上传
	if file , e := target.UploadFile(audioFile , name); e != nil {
		panic(e)
	} else {
		return file
	}
}


//阿里云录音文件识别
func AliyunAudioRecognition(engine asr.AliyunCloud , filelink string , intelligent_block bool) (AudioResult map[int64][] *asr.AliyunAudioRecognitionResult) {
	//创建识别请求
	taskid, client, e := engine.NewAudioFile(filelink)
	if e != nil {
		panic(e)
	}

	AudioResult = make(map[int64][] *asr.AliyunAudioRecognitionResult)

	//遍历获取识别结果
	engine.GetAudioFileResult(taskid , client , func(result []byte) {

		//结果处理
		statusText, _ := jsonparser.GetString(result, "StatusText") //结果状态
		if statusText == asr.STATUS_SUCCESS {

			//智能分段
			if intelligent_block {
				 asr.AliyunAudioResultWordHandle(result , func(vresult *asr.AliyunAudioRecognitionResult) {
					channelId := vresult.ChannelId

					_ , isPresent  := AudioResult[channelId]
					if isPresent {
						//追加
						AudioResult[channelId] = append(AudioResult[channelId] , vresult)
					} else {
						//初始
						AudioResult[channelId] = []*asr.AliyunAudioRecognitionResult{}
						AudioResult[channelId] = append(AudioResult[channelId] , vresult)
					}
				})
				return
			}

			_, err := jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				text , _ := jsonparser.GetString(value, "Text")
				channelId , _ := jsonparser.GetInt(value, "ChannelId")
				beginTime , _ := jsonparser.GetInt(value, "BeginTime")
				endTime , _ := jsonparser.GetInt(value, "EndTime")
				silenceDuration , _ := jsonparser.GetInt(value, "SilenceDuration")
				speechRate , _ := jsonparser.GetInt(value, "SpeechRate")
				emotionValue , _ := jsonparser.GetInt(value, "EmotionValue")

				vresult := &asr.AliyunAudioRecognitionResult {
					Text:text,
					ChannelId:channelId,
					BeginTime:beginTime,
					EndTime:endTime,
					SilenceDuration:silenceDuration,
					SpeechRate:speechRate,
					EmotionValue:emotionValue,
				}

				_ , isPresent  := AudioResult[channelId]
				if isPresent {
					//追加
					AudioResult[channelId] = append(AudioResult[channelId] , vresult)
				} else {
					//初始
					AudioResult[channelId] = []*asr.AliyunAudioRecognitionResult{}
					AudioResult[channelId] = append(AudioResult[channelId] , vresult)
				}
			} , "Result", "Sentences")
			if err != nil {
				panic(err)
			}
		}
	})

	return
}


//阿里云录音识别结果集生成字幕文件
func AliyunAudioResultMakeSubtitleFile (video string , AudioResult map[int64][] *asr.AliyunAudioRecognitionResult)  {
	subfile := tools.GetFileBaseName(video)

	for _,result := range AudioResult {
		srtfile := "srt-results" + "/" + subfile + ".srt"
		//输出字幕文件rt
		//println("字幕文件位置:", srtfile)
		file, e := os.Create(srtfile)
		if e != nil {
			panic(e)
		}

		defer file.Close() //defer

		index := 0
		for _ , data := range result {
			linestr := MakeSubtitleText(index , data.BeginTime , data.EndTime , data.Text)

			file.WriteString(linestr)

			index++
		}
	}
}


//拼接字幕字符串
func MakeSubtitleText(index int , startTime int64 , endTime int64 , text string) string {
	var content bytes.Buffer
	content.WriteString(strconv.Itoa(index))
	content.WriteString("\n")
	content.WriteString(tools.SubtitleTimeMillisecond(startTime))
	content.WriteString(" --> ")
	content.WriteString(tools.SubtitleTimeMillisecond(endTime))
	content.WriteString("\n")
	content.WriteString(text)
	content.WriteString("\n")
	content.WriteString("\n")
	return content.String()
}