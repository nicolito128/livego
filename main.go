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

	http.Handle("/", FileHandler(realPath, *addr))
	http.HandleFunc("/_livego/reload", ReloadHandler)

	log.Printf("Starting server on %s, serving %s\n", *addr, *path)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("Error starting server: ", err)
	}
}

func FileHandler(dir, addr string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var file *os.File
		var err error
		var reader io.ReadSeeker

		if r.URL.Path == "/" {
			_, err = os.Stat(filepath.Join(string(dir), "index.html"))
			if err == nil {
				http.Redirect(w, r, "/index.html", http.StatusFound)
				return
			}
		}

		file, err = os.Open(filepath.Join(string(dir), r.URL.Path))
		if err != nil {
			fmt.Fprintf(w, "Error opening file: %v", err)
			return
		}
		defer file.Close()
		reader = file

		if fileInfo, _ := file.Stat(); !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".html") {
			b := make([]byte, fileInfo.Size())
			_, err = file.Read(b)
			if err != nil && err != io.EOF {
				fmt.Fprintf(w, "Error reading file: %v", err)
				return
			}
			// Inject the script
			data := strings.ReplaceAll(string(b), "</body>", GetInjectScript(addr)+"</body>")
			reader = strings.NewReader(data)
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
			fmt.Fprint(w, "Error parsing referer:", err)
			return
		}

		filePath, err := filepath.Abs(filepath.Join(*path, location.Path))
		if err != nil {
			fmt.Fprint(w, "Error trying to get file path:", err)
			return
		}

		err = WatchFile(filePath)
		if err != nil {
			fmt.Fprint(w, "Error watching file:", err)
			return
		}

		log.Println(color.BlueString("File reloaded:"), color.GreenString(location.Path))
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
