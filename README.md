# video-srt-command
一个自动生成字幕SRT文件的开源软件工具(命令行版本)

***开源声明：修改自[video-srt](https://github.com/wxbool/video-srt).目前video-srt项目存在些问题，无法正常运行。和原作者@Viggo沟通后，
对方表示暂时没有精力对video-srt这个项目进行更新维护。我对这个项目比较感兴趣，对项目进行了修改重构，使之能够正常运行。为了便于后续迭代更新，新建了video-srt-command开源项目。***

## 功能概要
    * 支持常见格式的视频和音频的字幕生成，可指定生成字幕的格式（srt,lrc,txt）
    * 支持字幕翻译成指定语种，可选双语字幕或翻译字幕

## 语音识别接口
本项目暂时只接入了阿里云的相关接口：[OSS对象存储](https://www.aliyun.com/product/oss?spm=5176.12825654.eofdhaal5.13.e9392c4aGfj5vj&aly_as=K11FcpO8)
、[录音文件识别](https://ai.aliyun.com/nls/filetrans?spm=5176.12061031.1228726.1.47fe3cb43I34mn)的相关业务接口。

## 翻译接口
本项目暂时只接入了百度翻译的接口：[百度翻译](https://fanyi-api.baidu.com/product/113)

## 快速体验
- step0: 从github下载项目源代码
```shell
git clone https://github.com/jinsongpan/video-srt-command.git
```
- step1：配置服务接口（config.ini）
```ini
[Sys]
tempDir=results/temp/audio

#字幕选项配置
[Srt]
#输出文件编码
outputEncode=utf8
#输出文件类型
outputType=srt
;#最大处理并发数
;maxConcurrency=1


#阿里云OSS对象服务配置
#文档：https://help.aliyun.com/document_detail/31827.html?spm=a2c4g.11186623.6.582.4e7858a85Dr5pA
[AliyunOSS]
# OSS对外服务的访问域名
endpoint=your.Endpoint
# 存储空间（Bucket）名称
bucketName=your.BucketName
# 存储空间（Bucket 域名）地址
bucketDomain=your.BucketDomain
accessKeyId=your.AccessKeyId
accessKeySecret=your.AccessKeySecret

#阿里云语音识别配置
#文档：https://help.aliyun.com/document_detail/90727.html?spm=a2c4g.11186623.6.581.691af6ebYsUkd1
[AliyunCloud]
#在管控台中创建的项目Appkey，项目的唯一标识
#中文语音识别接口
appKey=your.AppKey
#英文语音识别接口
;appKey=
accessKeyId=your.AccessKeyId
accessKeySecret=your.AccessKeySecret

#百度翻译配置
#文档：http://api.fanyi.baidu.com/api/trans/product/apidoc
[BaiduTranslate]
appID=your.AppId
secretKey=your.SecretKey
;账户类型暂时仅支持"标准版"和"高级版"
accountType=your.AccountType
```
- step2: 在根目录文件夹“video-srt-command”中放入需要生成字幕的视频文件test.mp4，执行以下操作，可得字幕文件：
```shell
video-srt-command -f=test.mp4
```
- step3: 在根目录文件夹“video-srt-command”中放入需要生成字幕的视频文件test.mp4，执行以下操作，可得双语字幕文件：
```shell
video-srt-command -f=test.mp4 -trans=true -outlang=2 -biling=true
```
如果命令行窗口出现“字幕制作完成字样！”，可在“results”文件夹中找到对应的字幕文件。

## 下载安装
```shell
go get -u github.com/jinsongpan/video-srt-command
```

## 依赖相关
```shell
golang >= 1.13 
ffmpeg >= 4.2.2
alibaba-cloud-sdk-go >=  1.61.480
aliyun-oss-go-sdk >=  2.1.4
jsonparser >= 1.0.0
errors >= 0.9.1
go.uuid >= v1.2.0
mahonia
```
*注1：golang需开启go module模式*
```shell script
go env -w GO111MODULE="on"
```
*注2：[ffmpeg](http://ffmpeg.org/) 依赖，请先下载安装，并设置环境变量*

*注3：可通过如下指令查看命令支持的输入参数*
```shell script
go run main.go --help
```


## 其他生成字幕方式
放入需要生成字幕的视频文件test.mp4，在根目录下执行：
```shell
go mod init video-srt-command // 执行一次
go mod vendor // 执行一次
go run main.go test.mp4
```
如果命令行窗口出现“字幕制作完成字样！”，可在“results”文件夹中找到对应的字幕文件。



## FAQ
* 支持哪些语言？
    * 识别的语种由阿里云语音识别接口决定（目前支持中英欧常见语种）；翻译的目标语种由百度翻译接口决定（目前支持27种语音）
* 如何得到config.ini中需要的参数？
    1. 注册阿里云账号
        * 账号快速实名认证
        * 开通 `访问控制` 服务，并创建角色，设置开放 `OSS对象存储`、`智能语音交互` 的访问权限 
        * 开通 `OSS对象存储` 服务，并创建一个存储空间（Bucket）（读写权限设置为公共读）
        * 开通 `智能语音交互` 服务，并创建项目（根据使用场景选择识别语言以及偏好等）
    2. 注册百度翻译账号
        * 开通`通用翻译API`服务