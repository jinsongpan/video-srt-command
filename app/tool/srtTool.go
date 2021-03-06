package tool

import (
	"bytes"
	"strconv"
	"strings"
)

// 时间戳转换成“hour:minute:second,millisecond”或“hour:minute:second”srt格式
func StampToTimeline(time int64, hasMillisecond bool) string {
	var second int64 = 0
	var minute int64 = 0
	var hour int64 = 0
	var millisecond int64 = 0

	millisecond = time % 1000
	second = time / 1000

	if second > 59 {
		minute = (time / 1000) / 60
		second = second % 60
	}
	if minute > 59 {
		hour = (time / 1000) / 3600
		minute = minute % 60
	}

	//00:00:06,770
	var secondText = PaddingStr(strconv.FormatInt(second, 10), 2)
	var minuteText = PaddingStr(strconv.FormatInt(minute, 10), 2)
	var hourText = PaddingStr(strconv.FormatInt(hour, 10), 2)
	var millisecondText = PaddingStr(strconv.FormatInt(millisecond, 10), 3)

	if hasMillisecond {
		return hourText + ":" + minuteText + ":" + secondText + "," + millisecondText
	}
	return hourText + ":" + minuteText + ":" + secondText
}

// 时间戳转换成“[minute:second,millisecond]lrc格式
func StampToLrcTime(time int64, hasMillisecond bool) string {
	var second int64 = 0
	var minute int64 = 0
	var millisecond int64 = 0

	millisecond = time % 1000
	second = time / 1000

	if second > 59 {
		minute = (time / 1000) / 60
		second = second % 60
	}

	//00:00:06,770
	var secondText = PaddingStr(strconv.FormatInt(second, 10), 2)
	var minuteText = PaddingStr(strconv.FormatInt(minute, 10), 2)
	var millisecondText = PaddingStr(strconv.FormatInt(millisecond, 10), 3)

	if hasMillisecond {
		return "[" + minuteText + ":" + secondText + "," + millisecondText + "]"
	}
	return "[" + minuteText + ":" + secondText + "]"
}

// 在时间轴中填充"0"字符串，使格式统一
func PaddingStr(time string, Num int) string {
	len_time := len(time)

	if len_time >= Num {
		return time
	}

	return strings.Repeat("0", Num-len_time) + time
}

//拼接字符串,生成字幕
func GenSubtitle(index int, startTime int64, endTime int64, text string, translateText string, isTranslate bool, isBilingual bool) string {
	var content bytes.Buffer
	content.WriteString(strconv.Itoa(index))
	// 换行符：Linux "\n"；Windows "\r\n"；Mac "\r"
	content.WriteString("\r\n")
	content.WriteString(StampToTimeline(startTime, true))
	content.WriteString(" --> ")
	content.WriteString(StampToTimeline(endTime, true))
	content.WriteString("\r\n")

	if isTranslate {
		if isBilingual {
			content.WriteString(translateText)
			content.WriteString("\r\n")
			content.WriteString(text)

		} else {
			content.WriteString(translateText)
		}
	} else {
		content.WriteString(text)
	}

	content.WriteString("\r\n")
	content.WriteString("\r\n")

	return content.String()
}

//拼接字符串,生成歌词
func GenLyric(startTime int64, text string, translateText string, isTranslate bool, isBilingual bool) string {
	var content bytes.Buffer
	//换行符：Linux "\n"；Windows "\r\n"；Mac "\r"
	content.WriteString(StampToLrcTime(startTime, true))
	if isTranslate {
		if isBilingual {
			content.WriteString(translateText)
			content.WriteString("\r\n")
			content.WriteString(StampToLrcTime(startTime, true))
			content.WriteString(text)

		} else {
			content.WriteString(translateText)
		}
	} else {
		content.WriteString(text)
	}

	content.WriteString("\r\n")
	content.WriteString("\r\n")

	return content.String()
}

//拼接字符串,生成文本
func GenText(text string, translateText string, isTranslate bool, isBilingual bool) string {
	var content bytes.Buffer

	if isTranslate {
		if isBilingual {
			content.WriteString(translateText)
			content.WriteString("\r\n")
			content.WriteString(text)

		} else {
			content.WriteString(translateText)
		}
	} else {
		content.WriteString(text)
	}

	content.WriteString("\r\n")
	content.WriteString("\r\n")

	return content.String()
}
