package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/joho/godotenv"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

var awsBucketName, filename, filepath string

const s3Timeout = 5 * time.Second
const syncInterval = 30 * time.Second

func main() {
	// dev variables
	err := godotenv.Load(".env")
	if err != nil {
		// no worries
	}

	variables := []string{
		"AWS_ACCESS_KEY",
		"AWS_SECRET_KEY",
		"AWS_REGION",
		"AWS_BUCKET_NAME",
		"FILE_NAME",
	}

	for _, variable := range variables {
		if isEnvVariableEmpty(variable) {
			logError(fmt.Errorf("%s env variable is empty", variable))
			return
		}
	}

	filename = os.Getenv("FILE_NAME")
	awsBucketName = os.Getenv("AWS_BUCKET_NAME")
	filepath = fmt.Sprintf("data/%s", filename)

	var lastModCache int64 = 0

	for {
		info, err := os.Stat(filepath)
		if err != nil || info == nil {
			logError(fmt.Errorf("could not open file. file does not exist. %w", err))
		}
		lastMod := info.ModTime().Unix()
		if lastModCache != lastMod {
			lastModCache = lastMod
			uploadFile()
		}

		time.Sleep(syncInterval)
	}

}

func uploadFile() {
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		logError(fmt.Errorf("could not read file: %w", err))
		return
	}

	sess := session.Must(session.NewSession())
	svc := s3.New(sess)
	ctx := context.Background()
	var cancelFn func()
	ctx, cancelFn = context.WithTimeout(ctx, s3Timeout)

	if cancelFn != nil {
		defer cancelFn()
	}

	key := time.Now().Format("2006-01-02-15-04-05") + "-" + filename

	_, err = svc.PutObjectWithContext(ctx, &s3.PutObjectInput{
		Bucket: &awsBucketName,
		Key:    aws.String(key),
		Body:   strings.NewReader(string(data)),
	})
	if err != nil {
		logError(fmt.Errorf("failed to upload object, %w", err))
	}
}

func isEnvVariableEmpty(name string) bool {
	v := os.Getenv(name)
	if v == "" {
		return true
	}

	return false
}

func logError(msg error) {
	_, _ = fmt.Fprintf(os.Stderr, msg.Error())
}