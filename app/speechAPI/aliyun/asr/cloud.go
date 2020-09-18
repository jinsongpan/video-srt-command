package asr

import (
	"encoding/json"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk"
	"github.com/aliyun/alibaba-cloud-sdk-go/sdk/requests"
	"github.com/axgle/mahonia"
	"github.com/buger/jsonparser"
	"github.com/pkg/errors"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"video-srt-command/app/tool"
)

type AliyunCloud struct {
	AccessKeyId     string
	AccessKeySecret string
	AppKey          string
}

//阿里云录音文件识别结果集
type AliyunAudioRecResult struct {
	Text            string //文本结果
	TranslateText   string //翻译文本结果
	ChannelId       int64  //音轨ID
	BeginTime       int64  //该句的起始时间偏移，单位为毫秒
	EndTime         int64  //该句的结束时间偏移，单位为毫秒
	SilenceDuration int64  //本句与上一句之间的静音时长，单位为秒
	SpeechRate      int64  //本句的平均语速，单位为每分钟字数
	EmotionValue    int64  //情绪能量值1-10，值越高情绪越强烈
}

//阿里云识别词语数据集
type AliyunAudioWord struct {
	Word      string
	ChannelId int64
	BeginTime int64
	EndTime   int64
}

// Go SDK site: https://help.aliyun.com/document_detail/94072.html?spm=a2c4g.11186623.6.598.1f261f8dtL9QXW
// 地域ID，常量内容，请勿改变
const REGION_ID string = "cn-shanghai"
const ENDPOINT_NAME string = "cn-shanghai"
const PRODUCT string = "nls-filetrans"
const DOMAIN string = "filetrans.cn-shanghai.aliyuncs.com"
const API_VERSION string = "2018-08-17"
const POST_REQUEST_ACTION string = "SubmitTask"
const GET_REQUEST_ACTION string = "GetTaskResult"

// 请求参数key
const KEY_APP_KEY string = "appkey"
const KEY_FILE_LINK string = "file_link"
const KEY_VERSION string = "version"
const KEY_ENABLE_WORDS string = "enable_words"
const KEY_ENABLE_INVERSE_TEXT_NORMALIZATION string = "enable_inverse_text_normalization"

// 响应参数key
const KEY_TASK string = "Task"
const KEY_TASK_ID string = "TaskId"
const KEY_STATUS_TEXT string = "StatusText"
const KEY_RESULT string = "Result"

// 状态值
const STATUS_SUCCESS string = "SUCCESS"
const STATUS_RUNNING string = "RUNNING"
const STATUS_QUEUEING string = "QUEUEING"

//录音文件识别
//接口文档 https://help.aliyun.com/document_detail/90727.html?spm=a2c4g.11186623.6.581.691af6ebYsUkd1
func (c AliyunCloud) NewAudioFile(fileLink string) (string, *sdk.Client, error) {
	client, err := sdk.NewClientWithAccessKey(REGION_ID, c.AccessKeyId, c.AccessKeySecret)

	if err != nil {
		return "", client, err
	}

	postRequest := requests.NewCommonRequest()
	postRequest.Domain = DOMAIN
	postRequest.Version = API_VERSION
	postRequest.Product = PRODUCT
	postRequest.ApiName = POST_REQUEST_ACTION
	postRequest.Method = "POST"

	mapTask := make(map[string]string)
	mapTask[KEY_APP_KEY] = c.AppKey
	mapTask[KEY_FILE_LINK] = fileLink
	// 新接入请使用4.0版本，已接入(默认2.0)如需维持现状，请注释掉该参数设置
	mapTask[KEY_VERSION] = "4.0"
	// 设置是否输出词信息，默认为false，开启时需要设置version为4.0
	mapTask[KEY_ENABLE_WORDS] = "true"
	// 打开ITN，中文数字将转为阿拉伯数字输出，默认值为false，开启时需要设置version为“4.0”，
	mapTask[KEY_ENABLE_INVERSE_TEXT_NORMALIZATION] = "true"
	// to json
	task, err := json.Marshal(mapTask)

	if err != nil {
		return "", client, errors.New("cloud.go: to json error .")
	}
	postRequest.FormParams[KEY_TASK] = string(task)
	// 发起请求
	postResponse, err := client.ProcessCommonRequest(postRequest)
	if err != nil {
		return "", client, err
	}
	postResponseContent := postResponse.GetHttpContentString()
	//校验请求
	if postResponse.GetHttpStatus() != 200 {
		return "", client, errors.New("cloud.go:录音文件识别请求失败 , Http错误码 : " + strconv.Itoa(postResponse.GetHttpStatus()))
	}
	//解析数据
	var postMapResult map[string]interface{}
	err = json.Unmarshal([]byte(postResponseContent), &postMapResult)
	if err != nil {
		return "", client, errors.New("cloud.go:to map struct error .")
	}

	var taskId = ""
	var statusText = ""
	statusText = postMapResult[KEY_STATUS_TEXT].(string)

	//fmt.Println("cloud.go:postMapResult：")
	//for k, v := range postMapResult {
	//	fmt.Printf("	keyname: %v valname: %v\n", k, v)
	//}

	//检验结果
	if statusText == STATUS_SUCCESS {
		taskId = postMapResult[KEY_TASK_ID].(string)
		return taskId, client, nil
	}

	return "", client, errors.New("cloud.go:录音文件识别请求失败!")
}

//获取录音文件识别结果
//接口文档 https://help.aliyun.com/document_detail/90727.html?spm=a2c4g.11186623.6.581.691af6ebYsUkd1
func (c AliyunCloud) GetAudioFileResult(taskId string, client *sdk.Client, callback func(result []byte)) error {
	getRequest := requests.NewCommonRequest()
	getRequest.Domain = DOMAIN
	getRequest.Version = API_VERSION
	getRequest.Product = PRODUCT
	getRequest.ApiName = GET_REQUEST_ACTION
	getRequest.Method = "GET"
	getRequest.QueryParams[KEY_TASK_ID] = taskId
	statusText := ""

	//遍历获取识别结果
	for true {
		getResponse, err := client.ProcessCommonRequest(getRequest)
		if err != nil {
			return err
		}
		getResponseContent := getResponse.GetHttpContentString()

		if getResponse.GetHttpStatus() != 200 {
			return errors.New("cloud.go:识别结果查询请求失败 , Http错误码 : " + strconv.Itoa(getResponse.GetHttpStatus()))
		}

		var getMapResult map[string]interface{}
		err = json.Unmarshal([]byte(getResponseContent), &getMapResult)
		if err != nil {
			return err
		}

		//调用回调函数
		callback(getResponse.GetHttpContentBytes())

		//校验遍历条件
		statusText = getMapResult[KEY_STATUS_TEXT].(string)
		if statusText == STATUS_RUNNING || statusText == STATUS_QUEUEING {
			time.Sleep(3 * time.Second)
		} else {
			break
		}
	}

	if statusText != STATUS_SUCCESS {
		return errors.New("cloud.go:174:录音文件识别失败!")
	}

	return nil
}

//阿里云录音文件识别
func AliyunAudioRecognition(filelink string, engine AliyunCloud, autoBlock bool) (AudioResult map[int64][]*AliyunAudioRecResult) {
	//创建识别请求
	taskid, client, err := engine.NewAudioFile(filelink)
	if err != nil {
		panic(err)
	}

	AudioResult = make(map[int64][]*AliyunAudioRecResult)

	//遍历获取识别结果
	if err = engine.GetAudioFileResult(taskid, client, func(result []byte) {

		//结果处理
		statusText, _ := jsonparser.GetString(result, "StatusText") //结果状态
		if statusText == STATUS_SUCCESS {

			//自动分段
			if autoBlock {
				AliyunAudioResultWordHandle(result, func(vresult *AliyunAudioRecResult) {
					channelId := vresult.ChannelId

					_, isPresent := AudioResult[channelId]
					if isPresent {
						//追加
						AudioResult[channelId] = append(AudioResult[channelId], vresult)
					} else {
						//初始
						AudioResult[channelId] = []*AliyunAudioRecResult{}
						AudioResult[channelId] = append(AudioResult[channelId], vresult)
					}
				})
				return
			}

			_, err := jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
				text, _ := jsonparser.GetString(value, "Text")
				channelId, _ := jsonparser.GetInt(value, "ChannelId")
				beginTime, _ := jsonparser.GetInt(value, "BeginTime")
				endTime, _ := jsonparser.GetInt(value, "EndTime")
				silenceDuration, _ := jsonparser.GetInt(value, "SilenceDuration")
				speechRate, _ := jsonparser.GetInt(value, "SpeechRate")
				emotionValue, _ := jsonparser.GetInt(value, "EmotionValue")

				vresult := &AliyunAudioRecResult{
					Text:            text,
					ChannelId:       channelId,
					BeginTime:       beginTime,
					EndTime:         endTime,
					SilenceDuration: silenceDuration,
					SpeechRate:      speechRate,
					EmotionValue:    emotionValue,
				}

				_, isPresent := AudioResult[channelId]
				if isPresent {
					//追加
					AudioResult[channelId] = append(AudioResult[channelId], vresult)
				} else {
					//初始
					AudioResult[channelId] = []*AliyunAudioRecResult{}
					AudioResult[channelId] = append(AudioResult[channelId], vresult)
				}
			}, "Result", "Sentences")
			if err != nil {
				panic(err)
			}
		}
	}); err != nil {
		panic(err)
	}

	return
}

//使用阿里云录音识别结果集来生成指定类型的字幕文件
func AliyunRecResultGenText(media string, AudioResult map[int64][]*AliyunAudioRecResult, outputEncode string,
	outputType string, isTranslate bool, isBilingual bool) {

	mediaName := strings.TrimSuffix(path.Base(media), path.Ext(media)) // 获取文件名称
	textFile := "results" + "/" + mediaName + "." + outputType
	enc := mahonia.NewEncoder(outputEncode)

	file, err := os.Create(textFile)
	if err != nil {
		panic(err)
	}

	defer file.Close() //defer

	// 输出字幕文件
	if outputType == "srt" {
		for _, result := range AudioResult {
			index := 0
			for _, data := range result {

				if data.Text == "" {
					continue
				}
				textLine := enc.ConvertString(tool.GenSubtitle(index, data.BeginTime, data.EndTime, data.Text,
					data.TranslateText, isTranslate, isBilingual)) //将字幕文本写入textLine
				if _, err := file.WriteString(textLine); err != nil {
					panic(err)
				}
				index++
			}
		}
	} else if outputType == "lrc" {
		for _, result := range AudioResult {
			index := 0
			for _, data := range result {

				if data.Text == "" {
					continue
				}
				textLine := enc.ConvertString(tool.GenLyric(data.BeginTime, data.Text, data.TranslateText, isTranslate, isBilingual)) //将歌词文本写入textLine
				if _, err := file.WriteString(textLine); err != nil {
					panic(err)
				}
				index++
			}
		}
	} else if outputType == "txt" {
		for _, result := range AudioResult {
			index := 0
			for _, data := range result {

				if data.Text == "" {
					continue
				}
				textLine := enc.ConvertString(tool.GenText(data.Text, data.TranslateText, isTranslate, isBilingual)) //将文本写入textLine
				if _, err := file.WriteString(textLine); err != nil {
					panic(err)
				}
				index++
			}
		}
	} else {
		log.Println("输出文件类型 outputType 不是“srt, lrc, txt”中的一种，系统默认按srt格式保存内容...")
		for _, result := range AudioResult {
			index := 0
			for _, data := range result {

				if data.Text == "" {
					continue
				}
				textLine := enc.ConvertString(tool.GenSubtitle(index, data.BeginTime, data.EndTime, data.Text,
					data.TranslateText, isTranslate, isBilingual)) //将字幕文本写入textLine
				if _, err := file.WriteString(textLine); err != nil {
					panic(err)
				}
				index++
			}
		}
	}
}
