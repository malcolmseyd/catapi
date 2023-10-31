package main

import (
	"encoding/json"
	"errors"
	"fmt"
	_ "image/png"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gin-gonic/gin"
	"github.com/malcolmseyd/catapi/imageproc"
)

func must[T any](value T, err error) T {
	if err != nil {
		log.Fatalln("fatal error:", err)
	}
	return value
}

const catImagePath = "img"

var catImageIds []string = make([]string, 0)
var listenHost string = os.Getenv("LISTEN_HOST")
var listenPort string = os.Getenv("LISTEN_PORT")

func init() {
	for _, entry := range must(os.ReadDir(catImagePath)) {
		if entry.Type().IsRegular() {
			catImageIds = append(catImageIds, entry.Name())
		}
	}
	if listenPort == "" {
		listenPort = "8080"
	}
}

func main() {
	router := gin.Default()
	router.GET("/cat", func(c *gin.Context) {
		id := c.Query("id")
		if id == "" {
			id = catImageIds[rand.Int()%len(catImageIds)]
		}
		log.Println("using image with id", id)
		img, err := getCatImage(id)
		if errors.Is(err, os.ErrNotExist) {
			log.Println("cat image 404:", err)
			c.AbortWithStatus(404)
			return
		} else if err != nil {
			c.AbortWithError(500, err)
			return
		}
		if memeText := c.Query("text"); memeText != "" {
			img, err = imageproc.MakeMeme(img, memeText)
			if err != nil {
				c.AbortWithError(500, err)
				return
			}
		}

		isGithubBot := strings.Contains(c.Request.UserAgent(), "github-camo")
		if isGithubBot {
			c.Header("Cache-Control", "no-cache")
		}

		c.Data(200, mimetype.Detect(img).String(), img)

		if isGithubBot {
			time.Sleep(time.Millisecond * 200)
			purgeSelf()
		}
	})
	router.Run(net.JoinHostPort(listenHost, listenPort))
}

func getCatImage(id string) ([]byte, error) {
	sanitizedId := path.Join("/", id)
	return os.ReadFile(path.Join(catImagePath, sanitizedId))
}

func purgeSelf() {
	client := http.Client{Timeout: 10 * time.Second}

	selfURL, err := getSelfURL(&client)
	if err != nil {
		log.Println("failed to get self url:", err)
		return
	}

	purgeReq, err := http.NewRequest("PURGE", selfURL, nil)
	if err != nil {
		log.Println("bad url in purge request:", err)
		return
	}

	_, err = client.Do(purgeReq)
	if err != nil {
		log.Println("failed to purge self:", err)
		return
	}
	log.Println("successfully purged!")
}

var selfURLPattern = regexp.MustCompile(`<img[^>]+alt="cat"[^>]+ src="(https:\/\/camo[^"]*)"[^>]*>`)

func getSelfURL(client *http.Client) (string, error) {
	req, _ := http.NewRequest("GET", "https://github.com/malcolmseyd/malcolmseyd/blob/main/README.md", nil)
	req.Header.Add("Accept", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to request readme: %w", err)
	}
	defer resp.Body.Close()

	var body struct {
		Payload struct {
			Blob struct {
				RichText string `json:"richText"`
			} `json:"blob"`
		} `json:"payload"`
	}
	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		return "", fmt.Errorf("failed to decode json body: %w", err)
	}

	matches := selfURLPattern.FindStringSubmatch(body.Payload.Blob.RichText)
	if len(matches) < 2 {
		return "", fmt.Errorf("no match in readme")
	}
	return matches[1], nil
}
