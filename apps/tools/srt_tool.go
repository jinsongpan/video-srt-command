package tools

import (
	"strconv"
	"strings"
)

//字幕时间戳转换
func SubtitleTimeMillisecond(time int64) string {
	var miao int64 = 0
	var min int64 = 0
	var hours int64 = 0
	var millisecond int64 = 0

	millisecond = (time % 1000)
	miao = (time / 1000)

	if miao > 59 {
		min = (time / 1000) / 60
		miao = miao % 60
	}
	if min > 59 {
		hours = (time / 1000) / 3600
		min = min % 60
	}

	//00:00:06,770
	var miaoText = RepeatStr(strconv.FormatInt(miao , 10) , "0" , 2 , true)
	var minText = RepeatStr(strconv.FormatInt(min , 10) , "0" , 2 , true)
	var hoursText = RepeatStr(strconv.FormatInt(hours , 10) , "0" , 2 , true)
	var millisecondText = RepeatStr(strconv.FormatInt(millisecond , 10) , "0" , 3 , true)

	return hoursText + ":" + minText + ":" + miaoText + "," + millisecondText
}

func RepeatStr(str string , s string , length int , before bool) string {
	ln := len(str)

	if ln >= length {
		return str
	}

	if before {
		return  strings.Repeat(s , (length - ln)) + str
	} else {
		return  str + strings.Repeat(s , (length - ln))
	}
}
