package main

import (
	"errors"
	"fmt"
	"github.com/mouuff/go-rocket-update/pkg/provider"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestProviderAmazonS3(t *testing.T) {
	p := &AmazonS3{
		BucketName: "alex-hughes-bucket",
		Key:        fmt.Sprintf("golang-update-app/%v/latest/golang-update-app.zip", runtime.GOOS),
		Region:     "us-west-2",
	}

	if err := p.Open(); err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	err := ProviderTestWalkAndRetrieve(p)
	if err != nil {
		t.Fatal(err)
	}
}

// ProviderTestWalkAndRetrieve tests the expected behavior of a provider
func ProviderTestWalkAndRetrieve(p provider.AccessProvider) error {
	version, err := p.GetLatestVersion()
	if err != nil {
		return err
	}
	if len(version) < 1 { // TODO idea check version format?
		return errors.New("Bad version: " + version)
	}
	tmpDir, err := ioutil.TempDir("", "rocket-updater-awss3")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tmpDest := filepath.Join(tmpDir, "tmpDest")
	err = p.Retrieve("thisfiledoesnotexists", tmpDest)
	if err == nil {
		return errors.New("provider.Retrieve() should return an error when source file does not exists")
	}
	if FileExists(tmpDest) {
		return errors.New("provider.Retrieve() should not create destination file when source file does not exists")
	}

	filesCount := 0
	err = p.Walk(func(info *provider.FileInfo) error {
		destPath := filepath.Join(tmpDir, info.Path)
		if info.Mode.IsDir() {
			os.MkdirAll(destPath, os.ModePerm)
		} else {
			if strings.Contains(info.Path, SignatureRelPath) {
				return nil
			}
			filesCount += 1
			os.MkdirAll(filepath.Dir(destPath), os.ModePerm)
			err = p.Retrieve(info.Path, destPath)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	if filesCount <= 0 {
		return fmt.Errorf("filesCount <= 0")
	}

	err = p.Walk(func(info *provider.FileInfo) error {
		destPath := filepath.Join(tmpDir, info.Path)
		if !FileExists(destPath) && !strings.Contains(info.Path, SignatureRelPath) {
			return fmt.Errorf("File %s should exists", destPath)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Test to make sure the walk stops when walkFunc returns an error
	count := 0
	err = p.Walk(func(info *provider.FileInfo) error {
		count += 1
		return errors.New("Walk cancelled")
	})
	if err == nil {
		return errors.New("Walk should return the error of walkFunc")
	}
	if count > 1 {
		return errors.New("Walk should have stopped on error")
	}
	return nil
}

const SignatureRelPath = "signatures.json"

func FileExists(src string) bool {
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return false
	}
	return true
}
