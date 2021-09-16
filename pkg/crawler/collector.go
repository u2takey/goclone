package crawler

import (
	"context"
	"log"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"sync"

	"github.com/gocolly/colly"
	"github.com/imthaghost/goclone/pkg/file"
	"github.com/imthaghost/goclone/pkg/html"
)

var (
	lockedMap = sync.Map{}
)

// Collector searches for css, js, and images within a given link
// TODO improve for better performance
func Collector(ctx context.Context, url string, projectPath string, collyOpts ...func(*colly.Collector)) (err error) {
	// create a new collector
	c := colly.NewCollector(append(collyOpts, colly.Async(true),
		colly.CacheDir("/tmp/colly"))...)
	c.WithTransport(cancelableTransport{ctx: ctx, transport: http.DefaultTransport})
	_ = c.Limit(&colly.LimitRule{DomainGlob: "*", Parallelism: 1})

	if err = collect(c, url, projectPath, url, projectPath); err != nil {
		return
	}

	c.Wait()

	lockedMap.Range(func(key, value interface{}) bool {
		projectPath := value.(string)
		// Restructure html
		if err := html.LinkRestructure(projectPath); err != nil {
			log.Println("LinkRestructure failed: ", projectPath, err)
		}
		return true
	})

	return nil
}

func collect(c *colly.Collector, rootUri, rootPath, uri, path string) (err error) {

	if _, ok := lockedMap.LoadOrStore(uri, path); ok {
		return nil
	} else {
		log.Println("creating project", path)
		file.CreateProject(path)
	}

	c.OnHTML("a[href]", func(e *colly.HTMLElement) {
		// hyperlink reference
		link := e.Attr("href")
		if strings.HasPrefix(link, "/") && !strings.Contains(link, "?") && strings.Contains(link, "/c/") {
			subUri := strings.TrimRight(rootUri, "/") + link
			subPath := strings.TrimRight(rootPath, "/") + link

			log.Println("Found subProject: ", subUri, subPath)
			err := collect(c, rootUri, rootPath, subUri, subPath)
			if err != nil {
				log.Println("collect project failed", subPath, err)
			}

		}
	})

	// search for all link tags that have a rel attribute that is equal to stylesheet - CSS
	c.OnHTML("link[rel='stylesheet']", func(e *colly.HTMLElement) {
		// hyperlink reference
		link := e.Attr("href")
		// print css file was found
		log.Println("Css found", "-->", link)
		// extraction
		Extractor(e.Request.AbsoluteURL(link), path)
	})

	// search for all script tags with src attribute -- JS
	c.OnHTML("script[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		// Print link
		log.Println("Js found", "-->", link)
		// extraction
		Extractor(e.Request.AbsoluteURL(link), path)
	})

	// serach for all img tags with src attribute -- Images
	c.OnHTML("img[src]", func(e *colly.HTMLElement) {
		// src attribute
		link := e.Attr("src")
		// Print link
		log.Println("Img found", "-->", link)
		// extraction
		Extractor(e.Request.AbsoluteURL(link), path)
	})

	//Before making a request
	c.OnRequest(func(r *colly.Request) {
		link := r.URL.String()
		log.Println("Request for link: ", link)
		if uri == link {
			HTMLExtractor(link, path)
		}
	})

	// Visit each url and wait for stuff to load :)
	return c.Visit(uri)
}

// SetCookieJar returns a colly.Collector option that sets the cookie jar to the specified.
func SetCookieJar(jar *cookiejar.Jar) func(*colly.Collector) {
	return func(c *colly.Collector) { c.SetCookieJar(jar) }
}

type cancelableTransport struct {
	ctx       context.Context
	transport http.RoundTripper
}

func (t cancelableTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.ctx.Err(); err != nil {
		return nil, err
	}
	return t.transport.RoundTrip(req.WithContext(t.ctx))
}
