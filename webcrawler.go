package main

import (
    "fmt"
    "sync"
)

type Fetcher interface {
    // Fetch returns the body of URL and
    // a slice of URLs found on that page.
    Fetch(url string) (body string, urls []string, err error)
}

var fetched = make(map[string]bool)

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
func Crawl(url string, depth int, fetcher Fetcher, startCrawlFn func(string) bool, crawlComplete chan string, wg *sync.WaitGroup) {
    // TODO: Fetch URLs in parallel.
    // TODO: Don't fetch the same URL twice.
    // This implementation doesn't do either:
    wg.Add(1)
    defer wg.Done()

    fmt.Println("Crawling ", url)
    if depth <= 0 {
        fmt.Println("Depth 0, end crawl")
        crawlComplete <- url
        return
    }

    // fmt.Printf("Can crawl %v ? %v\n", url, canCrawl)
    
    body, urls, err := fetcher.Fetch(url)
    if err != nil {
        fmt.Println(err)
        crawlComplete <- url
        return
    }

    fmt.Printf("found: %s %q %s\n", url, body, urls)
    for _, u := range urls {
        fmt.Println("Check if can crawl", u)
        if startCrawlFn(u) {
            fmt.Println("Spawning child", u)
            go Crawl(u, depth-1, fetcher, startCrawlFn, crawlComplete, wg)
        }
    }

    fmt.Println("Send crawl complete signal", url)
    crawlComplete <- url

    return
}

type CanStartCrawlUrl struct {
    Url string
    ReplyChan chan bool
}

func initCrawler(startCrawl chan CanStartCrawlUrl, crawlComplete chan string) {
    for {
        select {
        case startCrawlData := <-startCrawl:
            _, otherCrawling := fetched[startCrawlData.Url]
            fmt.Printf("Can crawl %s ? %v\n", startCrawlData.Url, !otherCrawling)

            if !otherCrawling {
                fetched[startCrawlData.Url] = false
            } 

            startCrawlData.ReplyChan <- !otherCrawling
        case urlCrawled := <-crawlComplete:
            fmt.Println("Crawl ended", urlCrawled)
            fetched[urlCrawled] = true

            // for url, _ := range fetched {
            //     fmt.Println(url)
            // }

            // allChecked := true
            // for _, checkComplete := range fetched {
            //     allChecked = allChecked && checkComplete
            // }
            // if allChecked {
            //     v := make([]string, 0, len(fetched))

            //     for  url, _ := range fetched {
            //        v = append(v, url)
            //     }

            //     done <- v
            //     return
            // }
        }
    }

}

func main() {
    startCrawl := make(chan CanStartCrawlUrl)
    crawlComplete := make(chan string)
    //done := make(chan []string)
    wg := new(sync.WaitGroup)

    go initCrawler(startCrawl, crawlComplete)

    startCrawlFn := func (url string) bool {
        replychan := make(chan bool)
        startCrawl <- CanStartCrawlUrl{url, replychan}
        return <-replychan
    }

    Crawl("http://golang.org/", 4, fetcher, startCrawlFn, crawlComplete, wg)

    wg.Wait()
    v2 := make(map[string]bool)

    for  url, _ := range fetcher {
        v2[url] = false

        for _, childUrl := range fetcher[url].urls {
            v2[childUrl] = false
        }
    }

    v := make([]string, 0, len(fetched))

    for  url, _ := range fetched {
       v = append(v, url)
    }

    if(len(v) == len(v2)) {
        fmt.Println("PASSED")
    } else {
        fmt.Println("FAILED")
    }
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
    body string
    urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
    if res, ok := f[url]; ok {
        return res.body, res.urls, nil
    }
    return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
    "http://golang.org/": &fakeResult{
        "The Go Programming Language",
        []string{
            "http://golang.org/pkg/",
            "http://golang.org/cmd/",
        },
    },
    "http://golang.org/pkg/": &fakeResult{
        "Packages",
        []string{
            "http://golang.org/",
            "http://golang.org/cmd/",
            "http://golang.org/pkg/fmt/",
            "http://golang.org/pkg/os/",
        },
    },
    "http://golang.org/pkg/fmt/": &fakeResult{
        "Package fmt",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
    "http://golang.org/pkg/os/": &fakeResult{
        "Package os",
        []string{
            "http://golang.org/",
            "http://golang.org/pkg/",
        },
    },
}