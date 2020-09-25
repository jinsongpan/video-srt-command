package asr

import (
	"bufio"
	"github.com/buger/jsonparser"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

type AliyunAudioRecResultBlock struct {
	AliyunAudioRecResult
	Blocks []int
}

//阿里云录音录音文件识别 自动分段处理
func AliyunAudioResultWordHandle(result []byte, callback func(tmpResult *AliyunAudioRecResult)) {
	var (
		SentenceResult = make(map[int64][]*AliyunAudioRecResultBlock)
		wordResult  = make(map[int64][]*AliyunAudioWord)
		err         error
	)

	////test
	//AliyunAudioResult, err := jsonparser.GetUnsafeString(result)
	//if err != nil {
	//	panic(err)
	//}
	//log.Println("result", AliyunAudioResult)

	//获取整句识别数据集
	_, err = jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		text, _ := jsonparser.GetString(value, "Text")
		channelId, _ := jsonparser.GetInt(value, "ChannelId")
		beginTime, _ := jsonparser.GetInt(value, "BeginTime")
		endTime, _ := jsonparser.GetInt(value, "EndTime")
		silenceDuration, _ := jsonparser.GetInt(value, "SilenceDuration")
		speechRate, _ := jsonparser.GetInt(value, "SpeechRate")
		emotionValue, _ := jsonparser.GetInt(value, "EmotionValue")

		tmpResult := &AliyunAudioRecResultBlock{}
		tmpResult.Text = text
		tmpResult.ChannelId = channelId
		tmpResult.BeginTime = beginTime
		tmpResult.EndTime = endTime
		tmpResult.SilenceDuration = silenceDuration
		tmpResult.SpeechRate = speechRate
		tmpResult.EmotionValue = emotionValue

		log.Println(" tmpResult:", tmpResult)
		// isExist判断SentenceResult中是否有内容存在，如果为空，则需先绑定AliyunAudioRecResultBlock
		_, isExist := SentenceResult[channelId]
		log.Println("SentenceResult", SentenceResult)
		bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 1
		if isExist {
			//追加
			SentenceResult[channelId] = append(SentenceResult[channelId], tmpResult)
		} else {
			//初始
			SentenceResult[channelId] = []*AliyunAudioRecResultBlock{}
			SentenceResult[channelId] = append(SentenceResult[channelId], tmpResult)
		}
	}, "Result", "Sentences")
	if err != nil {
		panic(err)
	}

	//获取词语识别数据集
	_, err = jsonparser.ArrayEach(result, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		word, _ := jsonparser.GetString(value, "Word")
		channelId, _ := jsonparser.GetInt(value, "ChannelId")
		beginTime, _ := jsonparser.GetInt(value, "BeginTime")
		endTime, _ := jsonparser.GetInt(value, "EndTime")
		tmpResult := &AliyunAudioWord{
			Word:      word,
			ChannelId: channelId,
			BeginTime: beginTime,
			EndTime:   endTime,
		}
		_, isExist := wordResult[channelId]
		if isExist {
			//追加
			wordResult[channelId] = append(wordResult[channelId], tmpResult)
		} else {
			//初始
			wordResult[channelId] = []*AliyunAudioWord{}
			wordResult[channelId] = append(wordResult[channelId], tmpResult)
		}
	}, "Result", "Words")
	if err != nil {
		panic(err)
	}

	// 对识别的数据集进行处理
	puncStr := []string{"？", "。", "，", "！", "；", "、", "?", ".", ",", "!"}

	//句子数据集处理
	for _, value := range SentenceResult {
		for _, sentences := range value {
			sentences.Blocks = GetTextBlock(sentences.Text, puncStr)
			sentences.Text = ReplaceStrs(sentences.Text, puncStr, "")
			log.Println("sentences", sentences)
			//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 2
		}
	}

	//词语数据集处理
	for _, value := range wordResult {

		var (
			block     string = ""
			blockRune int    = 0
			lastBlock int    = 0
			beginTime int64  = 0
			blockBool        = false
			ischinese        = IsChineseWords(value) //校验中文
		)

		for i, word := range value {
			if blockBool || i == 0 {
				beginTime = word.BeginTime
				blockBool = false
			}

			if ischinese {
				block += word.Word
			} else {
				block += CompleSpace(word.Word) //补全空格
			}
			log.Println("block", block)
			blockRune = utf8.RuneCountInString(block)

			for channel, p := range SentenceResult {
				if word.ChannelId != channel {
					continue
				}
				for windex, w := range p {
					if word.BeginTime >= w.BeginTime && word.EndTime <= w.EndTime {
						flag := false
						early := false

						for t, B := range w.Blocks {
							if (blockRune >= B) && B != -1 {
								flag = true

								log.Println("block", block)
								log.Println("w.Text", w.Text)
								log.Println("w.Blocks", w.Blocks)
								log.Println(B, word.Word)

								//bufio.NewReader(os.Stdin).ReadBytes('\n') //断点 3

								var thisText = ""
								//容错机制
								if t == (len(w.Blocks) - 1) {
									thisText = SubString(w.Text, lastBlock, 10000)
								} else {
									//下个词提前结束
									if i < len(value)-1 && value[i+1].BeginTime >= w.EndTime {
										thisText = SubString(w.Text, lastBlock, 10000)
										early = true
									} else {
										thisText = SubString(w.Text, lastBlock, B-lastBlock)
									}
								}

								lastBlock = B
								if early == true {
									//全部设置为-1
									for vt, vb := range w.Blocks {
										if vb != -1 {
											w.Blocks[vt] = -1
										}
									}
								} else {
									w.Blocks[t] = -1
								}

								tmpResult := &AliyunAudioRecResult{
									Text:            thisText,
									ChannelId:       channel,
									BeginTime:       beginTime,
									EndTime:         word.EndTime,
									SilenceDuration: w.SilenceDuration,
									SpeechRate:      w.SpeechRate,
									EmotionValue:    w.EmotionValue,
								}
								callback(tmpResult) //回调传参

								blockBool = true
								break
							}
						}

						//fmt.Println("word.Word:" , word.Word)
						//fmt.Println(block)

						if FindSliceIntCount(w.Blocks, -1) == len(w.Blocks) {
							//全部截取完成
							block = ""
							lastBlock = 0
						}

						//容错机制
						if FindSliceIntCount(w.Blocks, -1) == (len(w.Blocks)-1) && flag == false {
							var thisText = SubString(w.Text, lastBlock, 10000)

							w.Blocks[len(w.Blocks)-1] = -1
							//tmpResult
							tmpResult := &AliyunAudioRecResult{
								Text:            thisText,
								ChannelId:       channel,
								BeginTime:       beginTime,
								EndTime:         w.EndTime,
								SilenceDuration: w.SilenceDuration,
								SpeechRate:      w.SpeechRate,
								EmotionValue:    w.EmotionValue,
							}

							//fmt.Println(  thisText )
							//fmt.Println(  block )
							//fmt.Println(  word.Word , beginTime, w.EndTime , flag  , word.EndTime  )

							callback(tmpResult) //回调传参

							//覆盖下一段落的时间戳
							if windex < (len(p) - 1) {
								beginTime = p[windex+1].BeginTime
							} else {
								beginTime = w.EndTime
							}

							//清除参数
							block = ""
							lastBlock = 0
						}
					}
				}
			}
		}
	}
}

func FindSliceIntCount(slice []int, target int) int {
	c := 0
	for _, v := range slice {
		if target == v {
			c++
		}
	}
	return c
}

func ReplaceStrs(strs string, olds []string, s string) string {
	for _, word := range olds {
		strs = strings.Replace(strs, word, s, -1)
	}
	return strs
}


//补全右边空格
func CompleSpace(s string) string {
	s = strings.TrimLeft(s, " ")
	s = strings.TrimRight(s, " ")
	return s + " "
}

func IsChineseWords(words []*AliyunAudioWord) bool {
	for _, v := range words {
		if IsChineseChar(v.Word) {
			return true
		}
	}
	return false
}

func IsChineseChar(str string) bool {
	for _, r := range str {
		if unicode.Is(unicode.Scripts["Han"], r) || (regexp.MustCompile("[\u3002\uff1b\uff0c\uff1a\u201c\u201d\uff08\uff09\u3001\uff1f\u300a\u300b]").MatchString(string(r))) {
			return true
		}
	}
	return false
}


func IsContain(items []rune, item rune) bool {
	for _, eachItem := range items {
		if eachItem == item {
			return true
		}
	}
	return false
}

func GetTextBlock(strs string, puncStr []string) []int {


	log.Println("strs", strs)

	puncsStr := strings.Join(puncStr, "")
	//获得标点和字符串对应的utf8编码值
	puncsRune := []rune(puncsStr)
	strsRune := []rune(strs)
	//log.Println("puncRune", puncsRune)
	//log.Println("strsRune", strsRune)

	//切分
	index := 0
	puncIndex := []int{}
	blocks := []int{}
	blocks = append(blocks, 0)

	for i, strRune := range strsRune {
		if IsContain(puncsRune, strRune) {
			puncIndex = append(puncIndex, i)
			for {
				if i - blocks[index] < 25 {
					blocks = append(blocks, i)
					index = len(blocks) - 1 //更新索引指向blocks数组中的最后一个元素
					sort.Ints(blocks)  //调整切块索引顺序，从小到大
					break
				}
				blocks = append(blocks, i)

				//TODO 改进切分算法，避免把词分开
				i -= 15
			}
		}
	}

	//除去开头植入的0元素
	blocks = blocks[1:]

	//消除标点带来的索引位移
	for i, block := range blocks {
		for _, punc := range puncIndex {
			if block <= punc {
				break
			}
			blocks[i] = blocks[i] - 1
		}
	}

	//log.Println("blocks", blocks)

	return blocks
}

func SubString(str string, begin int, length int) (substr string) {
	// 将字符串的转换成[]rune
	rs := []rune(str)
	lth := len(rs)

	// 简单的越界判断
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length
	if end > lth {
		end = lth
	}
	// 返回子串
	return string(rs[begin:end])
}
