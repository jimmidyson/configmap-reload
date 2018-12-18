package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	fsnotify "github.com/fsnotify/fsnotify"
)

var volumeDirs volumeDirsFlag
var webhook = flag.String("webhook-url", "", "the url to send a request to when the specified config map volume directory has been updated")
var webhookMethod = flag.String("webhook-method", "POST", "the HTTP method url to use to send the webhook")
var webhookStatusCode = flag.Int("webhook-status-code", 200, "the HTTP status code indicating successful triggering of reload")

func main() {
	flag.Var(&volumeDirs, "volume-dir", "the config map volume directory to watch for updates; may be used multiple times")
	flag.Parse()

	if len(volumeDirs) < 1 {
		log.Println("Missing volume-dir")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}
	if *webhook == "" {
		log.Println("Missing webhook")
		log.Println()
		flag.Usage()
		os.Exit(1)
	}

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
				if event.Op&fsnotify.Create == fsnotify.Create {
					if filepath.Base(event.Name) == "..data" {
						log.Println("config map updated")
						req, err := http.NewRequest(*webhookMethod, *webhook, nil)
						if err != nil {
							log.Println("error:", err)
							continue
						}
						resp, err := http.DefaultClient.Do(req)
						if err != nil {
							log.Println("error:", err)
							continue
						}
						resp.Body.Close()
						if resp.StatusCode != *webhookStatusCode {
							log.Println("error:", "Received response code", resp.StatusCode, ", expected", *webhookStatusCode)
							continue
						}
						log.Println("successfully triggered reload")
					}
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	for _, d := range volumeDirs {
		log.Printf("Watching directory: %q", d)
		err = watcher.Add(d)
		if err != nil {
			log.Fatal(err)
		}
	}
	<-done
}

type volumeDirsFlag []string

func (v *volumeDirsFlag) Set(value string) error {
	*v = append(*v, value)
	return nil
}

func (v *volumeDirsFlag) String() string {
	return fmt.Sprint(*v)
}
