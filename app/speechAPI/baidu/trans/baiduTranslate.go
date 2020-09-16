package trans

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"github.com/buger/jsonparser"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

//百度翻译支持的翻译语言选项及BaiduURL
//支持语言列表 http://api.fanyi.baidu.com/api/trans/product/apidoc#languageList
const (
	LANGUAGE_AUTO = 0  //未指定
	LANGUAGE_ZH   = 1  //中文
	LANGUAGE_EN   = 2  //英文
	LANGUAGE_YUE  = 3  //粤语
	LANGUAGE_WYW  = 4  //文言文
	LANGUAGE_JP   = 5  //日语
	LANGUAGE_KOR  = 6  //韩语
	LANGUAGE_FRA  = 7  //法语
	LANGUAGE_DE   = 8  //德语
	LANGUAGE_SPA  = 9 //西班牙语
	LANGUAGE_TH   = 10 //泰语
	LANGUAGE_ARA  = 11 //阿拉伯语
	LANGUAGE_RU   = 12 //俄语
	LANGUAGE_PT   = 13 //葡萄牙语
	LANGUAGE_IT   = 14 //意大利语
	LANGUAGE_EL   = 15 //希腊语
	LANGUAGE_NL   = 16 //荷兰语
	LANGUAGE_BUL  = 17 //波兰语
	LANGUAGE_EST  = 18 //保加利亚语
	LANGUAGE_DAN  = 19 //丹麦语
	LANGUAGE_FIN  = 20 //芬兰语
	LANGUAGE_CS   = 21 //捷克语
	LANGUAGE_ROM  = 22 //罗马尼亚语
	LANGUAGE_SLO  = 23 //斯洛文尼亚语
	LANGUAGE_SWE  = 24 //瑞典语
	LANGUAGE_HU   = 25 //匈牙利语
	LANGUAGE_CHT  = 26 //繁体中文
	LANGUAGE_VIE  = 27 //越南语

	BAIDURL string = "https://fanyi-api.baidu.com/api/trans/vip/translate"
)

type RunTranslateResult struct {
	From           string //翻译源语言
	To             string //译文语言
	TransResultSrc string //翻译结果（原文）
	TransResultDst string //翻译结果（译文）
}

//百度翻译配置
type BaiduTransConfig struct {
	AppID       string
	SecretKey   string
	AccountType string //账号认证类型"标准版","高级版"
}

//百度翻译结果集
type BaiduTransResult struct {
	TranslateText  string //翻译文本结果
	From           string //翻译源语言
	To             string //译文语言
	TransResultSrc string //翻译原文
	TransResultDst string //译文
	ErrorCode      int64  //错误码（仅当出现错误时存在）
	ErrorMsg       string //错误消息（仅当出现错误时存在）
}

//获取百度翻译引擎的语言字符标识
func GetLanguageChar(Language int) string {
	//baidutranslate引擎
	switch Language {
	case LANGUAGE_AUTO:
		return "auto"
	case LANGUAGE_ZH:
		return "zh"
	case LANGUAGE_EN:
		return "en"
	case LANGUAGE_YUE:
		return "yue"
	case LANGUAGE_WYW:
		return "wyw"
	case LANGUAGE_JP:
		return "jp"
	case LANGUAGE_KOR:
		return "kor"
	case LANGUAGE_FRA:
		return "fra"
	case LANGUAGE_DE:
		return "de"
	case LANGUAGE_SPA:
		return "spa"
	case LANGUAGE_TH:
		return "th"
	case LANGUAGE_RU:
		return "ru"
	case LANGUAGE_ARA:
		return "ara"
	case LANGUAGE_PT:
		return "pt"
	case LANGUAGE_IT:
		return "it"
	case LANGUAGE_EL:
		return "el"
	case LANGUAGE_NL:
		return "nl"
	case LANGUAGE_BUL:
		return "bul"
	case LANGUAGE_EST:
		return "est"
	case LANGUAGE_DAN:
		return "dan"
	case LANGUAGE_FIN:
		return "fin"
	case LANGUAGE_CS:
		return "cs"
	case LANGUAGE_ROM:
		return "rom"
	case LANGUAGE_SLO:
		return "slo"
	case LANGUAGE_SWE:
		return "swe"
	case LANGUAGE_HU:
		return "hu"
	case LANGUAGE_CHT:
		return "cht"
	case LANGUAGE_VIE:
		return "vie"

	}

	return ""
}

//获取某范围的随机整数
func GetIntRandNum(min int64, max int64) int64 {
	return rand.Int63n(max-min) + min
}

//文本md5
func GetMd5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

//百度api文档 http://api.fanyi.baidu.com/api/trans/product/apidoc
func (cfg BaiduTransConfig) BaiduTransAPI(strings string, from string, to string) (*BaiduTransResult, error) {

	params := &url.Values{}

	params.Add("q", strings)
	params.Add("appid", cfg.AppID)
	params.Add("salt", strconv.FormatInt(GetIntRandNum(32768, 65536), 10))
	params.Add("from", from)
	params.Add("to", to)
	params.Add("sign", cfg.BuildSign(strings, params.Get("salt")))

	return cfg.CallRequest(params)
}

//生成加密sign
func (cfg BaiduTransConfig) BuildSign(strings string, salt string) string {
	str := cfg.AppID + strings + salt + cfg.SecretKey
	return GetMd5(str)
}

//向百度翻译API发起请求
func (cfg BaiduTransConfig) CallRequest(params *url.Values) (*BaiduTransResult, error) {
	transURL := BAIDURL + "?" + params.Encode()

	request, err := http.NewRequest(http.MethodGet, transURL, nil)
	if err != nil {
		return nil, err
	}
	http.DefaultClient.Timeout = 60 * time.Second
	//do request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	//content
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	//解析数据
	errorCode, _ := jsonparser.GetString(content, "error_code")
	errorMsg, _ := jsonparser.GetString(content, "error_msg")
	from, _ := jsonparser.GetString(content, "from")
	to, _ := jsonparser.GetString(content, "to")

	errorCodeInt, _ := strconv.Atoi(errorCode)

	result := &BaiduTransResult{
		ErrorCode: int64(errorCodeInt),
		ErrorMsg:  errorMsg,
		From:      from,
		To:        to,
	}

	_, _ = jsonparser.ArrayEach(content, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		result.TransResultSrc, _ = jsonparser.GetString(value, "src")
		result.TransResultDst, _ = jsonparser.GetString(value, "dst")
	}, "trans_result")

	//翻译错误校验
	if result.ErrorCode != 0 {
		return nil, errors.New(result.ErrorMsg)
	}
	//log.Println("result", result)
	return result, nil
}

//翻译模块
func (cfg BaiduTransConfig) RunTranslate(text string, mediaName string, inputLang int, outputLang int) (*RunTranslateResult, error) {

	translateResult := new(RunTranslateResult)

	if cfg.AccountType == "标准版" {
		//百度翻译标准版休眠1000毫秒
		time.Sleep(time.Millisecond * 1000)
	} else {
		//休眠200毫秒
		time.Sleep(time.Millisecond * 200)
	}

	//转换语言字符标识
	from := GetLanguageChar(inputLang)
	to := GetLanguageChar(outputLang)

	//发起翻译请求
	links := 0
	baiduResult, err := cfg.BaiduTransAPI(text, from, to)
	for err != nil && links <= 5 {
		links++
		log.Println("百度翻译请求失败，重试第" + strconv.Itoa(links) + "次 ...")
		time.Sleep(time.Second * time.Duration(links))
		//重试
		baiduResult, err = cfg.BaiduTransAPI(text, from, to)
	}
	if err != nil {
		return translateResult, errors.New("翻译失败！错误信息：" + err.Error())
	}

	translateResult.TransResultDst = baiduResult.TransResultDst
	translateResult.TransResultSrc = baiduResult.TransResultSrc
	translateResult.From = baiduResult.From
	translateResult.To = baiduResult.To

	//log.Println("tanslateResult", translateResult)
	return translateResult, nil

}
