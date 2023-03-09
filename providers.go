package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// AmazonS3 provider
type AmazonS3 struct {
	// alex-hughes-bucket
	BucketName string
	// golang-update-app/%s/latest/golang-update-app.zip", %runtime.GOOS
	Key string
	// us-west-2
	Region string

	tmpDir          string        // temporary directory this is used internally
	zipProvider     *provider.Zip // provider used to unzip the downloaded zip
	zipPath         string        // path to the downloaded zip (should be in tmpDir)
	latestSignature string
}

type VersionData struct {
	Version   string `json:"version"`
	Signature string `json:"signature"`
}

// getReleases gets tags of the repository
func (p *AmazonS3) getLatestVersion() (string, error) {
	versionPath := filepath.Join(filepath.Dir(p.Key), "VERSION")
	versionUrl, err := p.getS3Url(versionPath)
	if err != nil {
		return "", err
	}
	resp, err := http.Get(versionUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var versionData VersionData
	err = json.Unmarshal(data, versionData)
	if err != nil {
		return "", nil
	}
	p.latestSignature = versionData.Signature
	return versionData.Version, nil
}

func (p *AmazonS3) getS3Url(key string) (string, error) {
	if key == "" {
		key = p.Key
	}

	awsConfig, err := config.LoadDefaultConfig(
		context.Background(),
		config.WithDefaultRegion(p.Region),
	)
	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(awsConfig)
	presignClient := s3.NewPresignClient(client)

	expiration := time.Now().Add(time.Minute * 5)
	disposition := fmt.Sprintf("attachment; filename=\"%v\"", filepath.Base(key))
	getObjectArgs := s3.GetObjectInput{
		Bucket:                     aws.String(p.BucketName),
		ResponseExpires:            &expiration,
		Key:                        aws.String(key),
		ResponseContentDisposition: aws.String(disposition),
	}

	res, err := presignClient.PresignGetObject(context.Background(), &getObjectArgs)
	if err != nil {
		return "", err
	}
	return res.URL, nil
}

func (p *AmazonS3) getLatestURL() (string, error) {
	return p.getS3Url(p.Key)
}

func (p *AmazonS3) Open() error {
	zipURL, err := p.getLatestURL() // get zip url for latest version
	if err != nil {
		return err
	}
	resp, err := http.Get(zipURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	p.tmpDir, err = ioutil.TempDir("", "rocket-updater-awss3")
	if err != nil {
		return err
	}

	p.zipPath = filepath.Join(p.tmpDir, filepath.Base(p.Key))
	zipFile, err := os.Create(p.zipPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(zipFile, resp.Body)
	zipFile.Close()
	if err != nil {
		return err
	}

	// Security
	verified, err := verifyBinary(p.zipPath, p.latestSignature)
	if err != nil || verified != true {
		// We're being hacked!
		return fmt.Errorf("Could not verify signature of update: %v", err)
	}

	p.zipProvider = &provider.Zip{Path: p.zipPath}
	return p.zipProvider.Open()

}

func (p *AmazonS3) Close() error {
	if p.zipProvider != nil {
		p.zipProvider.Close()
		p.zipProvider = nil
	}

	if len(p.tmpDir) > 0 {
		os.RemoveAll(p.tmpDir)
		p.tmpDir = ""
		p.zipPath = ""
	}
	return nil
}

func (p *AmazonS3) GetLatestVersion() (string, error) {
	// Check for a VERSION file in the same folder
	version, err := p.getLatestVersion()
	if err != nil {
		return "", err
	}
	return version, nil
}

func (p *AmazonS3) Walk(walkFn provider.WalkFunc) error {
	return p.zipProvider.Walk(walkFn)
}

func (p *AmazonS3) Retrieve(src, dest string) error {
	return p.zipProvider.Retrieve(src, dest)
}

// Gitlab provider finds a zip file in the repository's releases to provide files
type Gitlab struct {
	ProjectID int
	ZipName   string // Zip name (the zip you upload for a release on gitlab), example: binaries.zip

	tmpDir      string        // temporary directory this is used internally
	zipProvider *provider.Zip // provider used to unzip the downloaded zip
	zipPath     string        // path to the downloaded zip (should be in tmpDir)
}
