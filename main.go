package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
)

var (
	addr = flag.String("addr", "8080", "The address to listen on")
	path = flag.String("path", ".", "The path to watch a directory")
)

func main() {
	flag.Parse()
	if !strings.HasPrefix(*addr, ":") {
		*addr = ":" + *addr
	}

	// Join with the absolute path
	realPath, err := filepath.Abs(*path)
	if err != nil {
		log.Fatal("Error getting absolute path: ", err)
	}

	http.Handle("/", FileHandler(http.Dir(realPath), *addr))
	http.HandleFunc("/_livego/reload", ReloadHandler)

	log.Printf("Starting server on %s, serving %s\n", *addr, *path)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

func FileHandler(dir http.Dir, addr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Read file
		file, err := os.Open(filepath.Join(string(dir), r.URL.Path))
		if err != nil {
			fmt.Fprintf(w, "Error opening file: %v", err)
			return
		}
		defer file.Close()

		// Inject the script for hot reload if it's an HTML file
		var reader io.ReadSeeker = file

		if r.URL.Path == "/" {
			indexFile, err := os.Open(filepath.Join(string(dir), "index.html"))
			if err == nil {
				reader = indexFile
			}
			defer indexFile.Close()
			file.Close()
			file = indexFile
		}

		if fileInfo, _ := file.Stat(); !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".html") {
			b := make([]byte, fileInfo.Size())
			_, err := file.Read(b)
			if err != nil && err != io.EOF {
				fmt.Fprintf(w, "Error reading file: %v", err)
				return
			}

			// Inject the script
			data := AppendStrings(b, GetInjectScript(addr))
			reader = strings.NewReader(string(data))
		}

		http.ServeContent(w, r, r.URL.Path, time.Now(), reader)
	})
}

func ReloadHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Expires", "0")

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		location, err := url.Parse(r.Header.Get("Referer"))
		if err != nil {
			panic(err)
		}

		filePath, err := filepath.Abs(filepath.Join(*path, location.Path))
		if err != nil {
			panic(err)
		}

		err = WatchFile(filePath)
		if err != nil {
			panic(err)
		}

		log.Println(color.BlueString("File reloaded:"), location.Path)
		fmt.Fprintf(w, "data: reload\n\n")
		flusher.Flush()
	}
}

func WatchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}

func GetInjectScript(addr string) string {
	s := `<script type="text/javascript">var es = new EventSource("http://localhost%s/_livego/reload");es.onmessage = () => {location.reload()}</script>`
	s = fmt.Sprintf(s, addr)
	return s
}

func AppendStrings(data []byte, s ...string) []byte {
	for _, inj := range s {
		data = append(data, inj...)
	}

	return data
}
