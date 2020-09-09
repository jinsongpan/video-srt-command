# video-srt-command
一个自动生成字幕SRT文件的开源软件工具(命令行版本)

***开源声明：修改自[video-srt](https://github.com/wxbool/video-srt).目前video-srt项目存在些问题，无法正常运行。和原作者@Viggo沟通后，
表示暂时没有精力进行这个项目的更新迭代。我对这个项目比较感兴趣，对项目进行了修改重构，使之能够正常运行。为了便于后续迭代更新，新建了video-srt-command开源项目。***

## 语音识别接口
本项目暂时只接入了阿里云的相关接口：[OSS对象存储](https://www.aliyun.com/product/oss?spm=5176.12825654.eofdhaal5.13.e9392c4aGfj5vj&aly_as=K11FcpO8)
、[录音文件识别](https://ai.aliyun.com/nls/filetrans?spm=5176.12061031.1228726.1.47fe3cb43I34mn)的相关业务接口。

## 快速体验
- step1：配置服务接口（config.ini）
```ini
#字幕相关设置
[srt]
#智能分段处理：true（开启） false（关闭）
intelligent_block=true

#阿里云Oss对象服务配置
#文档：https://help.aliyun.com/document_detail/31827.html?spm=a2c4g.11186623.6.582.4e7858a85Dr5pA
[aliyunOss]
# OSS 对外服务的访问域名
endpoint=your.Endpoint
# 存储空间（Bucket）名称
bucketName=your.BucketName
# 存储空间（Bucket 域名）地址
bucketDomain=your.BucketDomain
accessKeyId=your.AccessKeyId
accessKeySecret=your.AccessKeySecret

#阿里云语音识别配置
#文档：
[aliyunCloud]
# 在管控台中创建的项目Appkey，项目的唯一标识
appKey=your.AppKey
accessKeyId=your.AccessKeyId
accessKeySecret=your.AccessKeySecret
```
- step2: 在根目录文件夹“video-srt-command”中放入需要生成字幕的视频文件test.mp4，执行以下操作：
```shell
video-srt-command test.mp4
```
如果命令行窗口出现“字幕制作完成字样！”，可在“srt-results”文件夹中找到对应的字幕文件。

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
```
*注1：golang需开启go module模式*
```shell script
go env -w GO111MODULE="on"
```
*注2：[ffmpeg](http://ffmpeg.org/) 依赖，请先下载安装，并设置环境变量.*


## 其他生成字幕方式
放入需要生成字幕的视频文件test.mp4，在根目录下执行：
```shell
go mod init video-srt-command // 执行一次
go mod vendor // 执行一次
go run main.go test.mp4
```
如果命令行窗口出现“字幕制作完成字样！”，可在“srt-results”文件夹中找到对应的字幕文件。



## FAQ
* 支持哪些语言？
    * 视频字幕文本识别的核心服务是由阿里云`录音文件识别`业务提供的接口进行的，支持汉语普通话、方言、欧美英语等语言
* 如何得到config.ini中需要的参数？
    * 注册阿里云账号
    * 账号快速实名认证
    * 开通 `访问控制` 服务，并创建角色，设置开放 `OSS对象存储`、`智能语音交互` 的访问权限 
    * 开通 `OSS对象存储` 服务，并创建一个存储空间（Bucket）（读写权限设置为公共读）
    * 开通 `智能语音交互` 服务，并创建项目（根据使用场景选择识别语言以及偏好等）
