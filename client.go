package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/urfave/cli"
)

var config struct {
	command      string
	onmodify     bool
	remove       bool
	url          string
	verbose      bool
	workdir      string
	writeTimeout int
}

var lastWrite map[string]time.Time
var mutex = &sync.Mutex{}

func main() {
	app := cli.NewApp()
	app.Name = "SCut client"
	app.Usage = "watch the workdir for new image files and upload them to url"
	app.Version = "1.1.0"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Stanislav Vetlovskiy",
			Email: "mrerliz@gmail.com",
		},
	}
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "url, u",
			Value:       "http://localhost:3000/",
			Usage:       "`URL` where to upload image",
			EnvVar:      "APP_URl",
			Destination: &config.url,
		},
		cli.StringFlag{
			Name:        "workdir, w",
			Value:       "~/Desktop/",
			Usage:       "`DIR` to whatch for new images",
			EnvVar:      "APP_WORKDIR",
			Destination: &config.workdir,
		},
		cli.StringFlag{
			Name:        "command, c",
			Usage:       "command that will be execute with response in first argument after successfully upload e.x. 'open'",
			Destination: &config.command,
		},
		cli.BoolFlag{
			Name:        "delete, r",
			Usage:       "will remove image file after upload",
			Destination: &config.remove,
		},
		cli.BoolFlag{
			Name:        "verbose",
			Usage:       "verbose logs",
			Destination: &config.verbose,
		},
		cli.BoolFlag{
			Name:        "onmodify, m",
			Usage:       "need for virtual env, cause there is no create event, just chmod",
			Destination: &config.onmodify,
		},
		cli.IntFlag{
			Name:        "timeout, t",
			Value:       300,
			Usage:       "upload timeout in ms after last write to file",
			Destination: &config.writeTimeout,
		},
	}
	app.Action = func(c *cli.Context) error {
		if !config.verbose {
			log.SetOutput(ioutil.Discard)
		}
		lastWrite = make(map[string]time.Time)
		serve()
		return nil
	}

	app.Run(os.Args)
}

func updateFileWriteTime(filePath string) {
	mutex.Lock()
	lastWrite[filePath] = time.Now()
	mutex.Unlock()
}

func getFileWriteTime(filePath string) time.Time {
	mutex.Lock()
	writeTime := lastWrite[filePath]
	mutex.Unlock()

	return writeTime
}

func serve() {
	watcher, err := fsnotify.NewWatcher()

	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				if event.Op == fsnotify.Create {
					updateFileWriteTime(event.Name)
					go onCreateHandler(event)
				} else if config.onmodify && event.Op == fsnotify.Chmod {
					updateFileWriteTime(event.Name)
					go onCreateHandler(event)
				} else if event.Op == fsnotify.Write {
					updateFileWriteTime(event.Name)
				} else {
					log.Print("Detect not handled event: ", event)
				}
			case err := <-watcher.Errors:
				log.Fatal("Error: ", err)
			}
		}
	}()

	err = watcher.Add(config.workdir)
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Wathcing for new files in dir: ", config.workdir)
	<-done
}

func onCreateHandler(e fsnotify.Event) {
	log.Print("Find new file: ", e.Name)
	for {
		if (time.Since(getFileWriteTime(e.Name)) / time.Millisecond) > time.Duration(config.writeTimeout) {
			break
		} else {
			log.Print("Waiting till write done for: ", e.Name)
			time.Sleep(time.Duration(config.writeTimeout) * time.Millisecond)
		}
	}

	filePath := e.Name
	fileExtension := strings.ToLower(filepath.Ext(filePath))
	switch fileExtension {
	case ".gif":
		fallthrough
	case ".png":
		fallthrough
	case ".jpeg":
		fallthrough
	case ".jpg":
		onImageCreateHandler(filePath)
	default:
		log.Print("Ignoring new file with extension: ", fileExtension)
	}
}

func onImageCreateHandler(filePath string) {
	log.Print("Find new image file: ", filePath)
	if response, ok := upload(filePath); ok {
		log.Printf("Image file successfully proceed: %s with response: '%s'", filePath, response)

		if len(config.command) > 0 {
			err := exec.Command(config.command, response).Run()
			if err != nil {
				log.Fatal(err)
			} else {
				log.Printf("Successfully execute: %s %s", config.command, response)
			}
		} else {
			fmt.Println(response)
		}

		if config.remove {
			err := os.Remove(filePath)
			if err != nil {
				log.Fatal(err)
			} else {
				log.Print("Image file successfully removed: ", filePath)
			}
		}
	}
}

func upload(filePath string) (response string, ok bool) {
	url := config.url + filepath.Base(filePath)
	fileHandler, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer fileHandler.Close()

	req, err := http.NewRequest("PUT", url, fileHandler)
	if err != nil {
		log.Fatal(err)
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	isSuccess := resp.StatusCode == 200
	if err != nil {
		log.Fatal(err)
	}
	bodyBites, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBites[:])
	return bodyString, isSuccess
}
