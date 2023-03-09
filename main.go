package main

import (
	"flag"
	"fmt"
	"log"
	"runtime"
	"sync"

	"github.com/mouuff/go-rocket-update/pkg/updater"
)

func main() {
	u := &updater.Updater{
		Provider: &AmazonS3{
			BucketName: "alex-hughes-bucket",
			Key:        fmt.Sprintf("golang-update-app/%v/latest/golang-update-app.zip", runtime.GOOS),
			Region:     "us-west-2",
		},
		ExecutableName: fmt.Sprintf("golang-update-app"),
		Version:        "v0.2.1", // You can change this value to trigger an update
	}

	versionFlag := false
	flag.BoolVar(&versionFlag, "version", false, "prints the version and exit")
	flag.Parse()

	if versionFlag {
		// we use this flag to verify the installation for this example:
		// https://github.com/mouuff/go-rocket-update/blob/master/examples/github-rollback/main.go
		fmt.Println(u.Version)
		return
	}

	log.Println("Current version: " + u.Version)
	log.Println("Looking for updates...")
	var wg sync.WaitGroup
	wg.Add(1)
	// For the example we run the update in the background
	// but you could directly call u.Update()
	var updateStatus updater.UpdateStatus
	var updateErr error
	go func() {
		updateStatus, updateErr = u.Update()
		wg.Done()
	}()

	// Here you can execute your program

	wg.Wait() // Waiting for the update process to finish before exiting
	if updateErr != nil {
		log.Println(updateErr)
	}
	if updateStatus == updater.Updated {
		log.Println("Updated!")
	}
}
