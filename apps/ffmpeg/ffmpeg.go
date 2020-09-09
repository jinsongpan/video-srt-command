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
func ExtractAudio (video string , tmpAudio string) (error) {
	ts := exec.Command("ffmpeg", "-version")
	if _, err := ts.CombinedOutput(); err != nil {
		return errors.New("请先安装 ffmpeg 依赖 ，并设置环境变量")
	}

	cmd := exec.Command("ffmpeg", "-i", video, "-ac", "1", "-ar", "16000", tmpAudio)
	//fmt.Println("ffmpeg.go:21:cmd", cmd)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + string(output))
	}
	return nil
}
