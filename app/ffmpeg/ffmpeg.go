package ffmpeg

import (
	"fmt"
	"github.com/pkg/errors"
	"os/exec"
)

type Ffmpeg struct {
	Os string //ffmpeg 文件目录
}

//提取视频音频
func ExtractAudio(video string, audio string) error {
	ts := exec.Command("ffmpeg", "-version")
	if _, err := ts.CombinedOutput(); err != nil {
		return errors.New("请先安装 ffmpeg 依赖 ，并设置环境变量")
	}

	// 抽取视频中的音频信息，并将其转换成16khz的单通道音频文件
	cmd := exec.Command("ffmpeg", "-i", video, "-ac", "1", "-ar", "16000", audio)
	//fmt.Println("ffmpeg.go:21:cmd", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return nil
}

// 转换音频格式
func AudioToMP3(Audio string, tmpAudio string) error {
	ts := exec.Command("ffmpeg", "-version")
	if _, err := ts.CombinedOutput(); err != nil {
		return errors.New("请先安装 ffmpeg 依赖 ，并设置环境变量")
	}
	// 抽取视频中的音频信息，并将其转换成16khz的单通道音频文件
	cmd := exec.Command("ffmpeg", "-i", Audio, "-acodec", "libmp3lame", "-ac", "1", "-ar", "16000", tmpAudio)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return nil
}
