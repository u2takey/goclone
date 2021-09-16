package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// HTMLExtractor ...
func HTMLExtractor(link string, projectPath string) {
	log.Println("Extracting --> ", link)

	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		log.Println("get link error", err)
		return
	}

	// Close the body once everything else is compled
	defer resp.Body.Close()

	// get the project name and path we use the path to
	var f *os.File
	for i := 1; i <= 5; i++ {
		f, err = os.OpenFile(projectPath+"/"+"index.html", os.O_RDWR|os.O_CREATE, 0777)
		if err != nil && strings.Contains(err.Error(), "too many open files") {
			time.Sleep(time.Second * time.Duration(i))
		}
		if err == nil {
			break
		}
	}
	if err != nil {
		log.Println("open file error", err)
		return
	}

	defer f.Close()
	htmlData, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err)
	}
	f.Write(htmlData)

}
