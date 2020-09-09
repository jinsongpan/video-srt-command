package tools

import (
	"os"
	"path"
	"path/filepath"
	"strings"
)

//win路径格式转换成unix
func WinDir(dir string) string {
	return strings.Replace(dir , "\\" , "/" , -1)
}

//获取文件名称（不带后缀）
func GetFileBaseName(filepath string) string {
	basefile := path.Base(filepath)
	ext := path.Ext(filepath)

	return strings.Replace(basefile , ext , "" , 1)
}

//检验目录是否存在
func DirExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}


//创建目录
func CreateDir(path string , all bool) error {
	var err error
	if all {
		err = os.Mkdir(path, os.ModePerm)
	} else {
		err = os.MkdirAll(path, os.ModePerm)
	}
	if err != nil {
		return err
	}
	return nil
}

////获取随机字符串
//func GetRandomCodeString(len int) string {
//	rand.Seed(time.Now().Unix())  //设置随机种子
//
//	seed := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
//	seedArr := strings.Split(seed , "")
//
//	result := []string{}
//	index := 0
//	for index < len {
//		s := GetIntRandomNumber(0 , 61)
//		result = append(result , seedArr[s])
//
//		index++
//	}
//
//	return strings.Join(result , "")
//}

////获取某范围的随机整数
//func GetIntRandomNumber(min int64 , max int64) int64 {
//	return rand.Int63n(max - min) + min
//}

//校验文件是否存在
func VaildVideo (video string) bool {
	_, err := os.Stat(video)  //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//获取应用根目录
func GetAppRootDir() string {
	if rootDir , err := filepath.Abs(filepath.Dir(os.Args[0])); err != nil {
		return ""
	} else {
		return WinDir(rootDir)
	}
}