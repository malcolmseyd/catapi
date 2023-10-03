package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sync"
	"time"

	"github.com/jessevdk/go-flags"
)

const URL = "https://cataas.com"

type CLIOptions struct {
	OutDir     string `long:"outdir" required:"true"`
	MaxWorkers int    `long:"maxworkers" default:"10"`
	Retries    int    `long:"retries" default:"5"`
}

func main() {
	var opts CLIOptions
	_, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	http.DefaultClient.Timeout = time.Second * 15
	ids := make([]string, 0)

	for {
		resp, err := http.Get(fmt.Sprintf("%s/api/cats?limit=1000&skip=%d", URL, len(ids)))
		if err != nil {
			log.Fatalln("failed to get cat list:", err)
		}
		defer resp.Body.Close()
		var data []struct {
			Id string `json:"_id"`
		}
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			log.Fatalln("failed to parse cat list:", err)
		}
		if len(data) == 0 {
			// no more cats
			break
		}
		for _, v := range data {
			ids = append(ids, v.Id)
		}
		log.Printf("fetched %v ids (%v total)", len(data), len(ids))
	}

	idChan := make(chan string, opts.MaxWorkers)
	wg := sync.WaitGroup{}
	wg.Add(opts.MaxWorkers)
	for i := 0; i < opts.MaxWorkers; i++ {
		go worker(&opts, idChan, &wg)
	}

	for _, v := range ids {
		idChan <- v
	}
	close(idChan)
	wg.Wait()
}

func worker(opts *CLIOptions, idChan <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for id := range idChan {
		succeeded := false
		for tries := 0; !succeeded && tries < opts.Retries; tries++ {
			if tries > 0 {
				log.Printf("retrying %s\n", id)
			}

			resp, err := http.Get(URL + "/cat/" + id)
			if err != nil {
				log.Printf("error getting cat %s: %v\n", id, err)
				continue
			}
			if resp.StatusCode != 200 {
				log.Printf("returned status %s for cat %v\n", resp.Status, id)
				if resp.StatusCode == 429 {
					log.Println("we're being rate limited! lets slow down")
					log.Println("headers:", resp.Header)
					time.Sleep(time.Second)
				}
				continue
			}

			b := bytes.NewBuffer(nil)
			_, err = io.Copy(b, resp.Body)
			if err != nil {
				log.Printf("error reading from response: %v\n", err)
				continue
			}

			path := path.Join(opts.OutDir, id)
			file, err := os.Create(path)
			if err != nil {
				log.Printf("error creating image file %s: %v\n", path, err)
				continue
			}
			defer file.Close()

			_, err = io.Copy(file, b)
			if err != nil {
				log.Printf("error copying body to file %s: %v\n", path, err)
				continue
			}

			succeeded = true
			log.Printf("wrote %s to disk\n", path)
		}
	}
}
