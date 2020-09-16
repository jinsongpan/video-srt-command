package oss

import (
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"os"
	"strconv"
	"strings"
	"time"
)

//Aliyun OSS SDK for Go
//https://github.com/aliyun/aliyun-oss-go-sdk/blob/master/README-CN.md

type AliyunOSS struct {
	Endpoint        string  //OSS对外服务的访问域名
	AccessKeyId     string
	AccessKeySecret string
	BucketName      string  //存储空间（Bucket）名称
	BucketDomain    string  //Bucket 域名
}

//获取存储空间列表（List Bucket）
func (c AliyunOSS) GetListBuckets() ([]string, error) {
	client, err := oss.New(c.Endpoint, c.AccessKeyId, c.AccessKeySecret)
	if err != nil {
		return nil, err
	}

	lsRes, err := client.ListBuckets()
	if err != nil {
		return nil, err
	}

	result := []string{}  // 存储空间列表
	for _, bucket := range lsRes.Buckets {
		result = append(result, bucket.Name)
	}

	return result, nil
}

//上传本地文件
//localFileName:本地文件
//objectName:OSS文件名称
func (c AliyunOSS) UploadFile(localFileName string, objectName string) (string, error) {
	// 创建OssClient实例
	client, err := oss.New(c.Endpoint, c.AccessKeyId, c.AccessKeySecret)
	if err != nil {
		return "", err
	}
	// 获取存储空间
	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return "", err
	}

	//分日期存储
	date := time.Now()
	year := date.Year()
	month := date.Month()
	day := date.Day()
	objectName = strconv.Itoa(year) + "/" + strconv.Itoa(int(month)) + "/" + strconv.Itoa(day) + "/" + objectName

	// 上传文件
	err = bucket.PutObjectFromFile(objectName, localFileName)
	if err != nil {
		return "", err
	}

	return objectName, nil
}

//上传音频至AliyunOSS
func UploadAudioToCloud(target AliyunOSS, audioFile string) string {
	name := ""
	//提取文件名称
	if fileInfo, err := os.Stat(audioFile); err != nil {
		panic(err)
	} else {
		name = fileInfo.Name()
	}

	//开始上传
	if file, err := target.UploadFile(audioFile, name); err != nil {
		panic(err)
	} else {
		return file
	}
}

//删除oss文件
func (c AliyunOSS) DelFile(objectName string) error {
	// 创建OSSClient实例
	client, err := oss.New(c.Endpoint, c.AccessKeyId, c.AccessKeySecret)
	if err != nil {
		return err
	}
	// 获取存储空间
	bucket, err := client.Bucket(c.BucketName)
	if err != nil {
		return err
	}

	// objectName表示删除OSS文件时需要指定包含文件后缀在内的完整路径，例如abc/efg/123.jpg。
	// 如需删除文件夹，请将objectName设置为对应的文件夹名称。如果文件夹非空，则需要将文件夹下的所有object删除后才能删除该文件夹。
	err = bucket.DeleteObject(objectName)
	if err != nil {
		return err
	}
	return nil
}

func DelOSSTempFile(target AliyunOSS, OSSFile string) error {
	if err := target.DelFile(OSSFile); err != nil {
		return err
	}
	return nil
}

//获取文件 file link
func (c AliyunOSS) GetObjectFileUrl(objectFile string) string {
	if strings.Index(c.BucketDomain, "http://") == -1 && strings.Index(c.BucketDomain, "https://") == -1 {
		return "http://" + c.BucketDomain + "/" + objectFile
	} else {
		return c.BucketDomain + "/" + objectFile
	}
}
