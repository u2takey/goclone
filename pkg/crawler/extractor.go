package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/imthaghost/goclone/pkg/parser"
)

// file extension map for directing files to their proper directory in O(1) time
var (
	extensionDir = map[string]string{
		".css":  "css",
		".js":   "js",
		".jpg":  "imgs",
		".jpeg": "imgs",
		".gif":  "imgs",
		".png":  "imgs",
		".svg":  "imgs",
	}
)

// Extractor visits a link determines if its a page or sublink
// downloads the contents to a correct directory in project folder
// TODO add functionality for determining if page or sublink
func Extractor(link string, projectPath string) {
	log.Println("Extracting --> ", link)

	// get the html body
	resp, err := http.Get(link)
	if err != nil {
		log.Println("get link error", err)
		return
	}

	// Closure
	defer resp.Body.Close()
	// file base
	base := parser.URLFilename(link)
	// store the old ext, in special cases the ext is weird ".css?a134fv"
	oldExt := filepath.Ext(base)
	// new file extension
	ext := parser.URLExtension(link)

	// checks if there was a valid extension
	if ext != "" {
		// checks if that extension has a directory path name associated with it
		// from the extensionDir map
		dirPath := extensionDir[ext]
		if dirPath != "" {
			// If extension and path are valid pass to writeFileToPath
			writeFileToPath(projectPath, base, oldExt, ext, dirPath, resp)
		}
	}
}

func writeFileToPath(projectPath, base, oldFileExt, newFileExt, fileDir string, resp *http.Response) {
	var name = base[0 : len(base)-len(oldFileExt)]
	document := name + newFileExt

	// get the project name and path we use the path to

	var f *os.File
	var err error
	for i := 1; i <= 5; i++ {
		f, err = os.OpenFile(projectPath+"/"+fileDir+"/"+document, os.O_RDWR|os.O_CREATE, 0777)
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
