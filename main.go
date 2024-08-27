package main

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/fatih/color"
)

var absolutePath string

func main() {
	var cmdPath, port string
	flag.StringVar(&cmdPath, "path", ".", "Set the path to watch files")
	flag.StringVar(&port, "port", ":5500", "Set the port to listen and serve")

	flag.Parse()
	if !strings.HasPrefix(port, ":") {
		port = ":" + (port)
	}

	absPath, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	// Clean path
	rx := regexp.MustCompile(`\\|\/`)
	sep := string(os.PathSeparator)
	res := rx.ReplaceAllLiteralString(cmdPath, sep)
	// composing path
	parts := strings.Split(res, sep)
	parts = append([]string{absPath}, parts...)

	absolutePath = filepath.Join(parts...)

	http.HandleFunc("/", readDir(absolutePath, port))
	http.HandleFunc("/_livego/reload", reloadHandler)

	startMsg := fmt.Sprintf("Server running at http://localhost%s/ - Press CTRL+C to exit", port)
	fmt.Println(color.YellowString(startMsg))
	http.ListenAndServe(port, nil)
}

func readDir(dir, port string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(dir, r.URL.Path)

		file, err := os.Open(filePath)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		data, err := io.ReadAll(file)
		if err != nil {
			fmt.Fprintf(w, err.Error())
			return
		}

		switch filepath.Ext(file.Name()) {
		case ".html":
			w.Header().Set("Content-Type", "text/html")
			data = injectString(data, injectScript(port))

		case ".json":
			w.Header().Set("Content-Type", "application/json")

		case ".txt", ".conf", ".md", ".yml", ".toml":
			w.Header().Set("Content-Type", "text/html")
			// escape all the html code
			data = []byte(html.EscapeString(string(data)))
			// injections
			data = injectString(data, injectScript(port), injectTxtBodyStyles())

		default:
			w.Header().Set("Content-Type", "text/plain")
		}

		log.Println(color.BlueString("File reloaded:"), color.GreenString(file.Name()))
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	}
}

func reloadHandler(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	for {
		location, err := url.Parse(r.Header.Get("Referer"))
		if err != nil {
			panic(err)
		}

		filePath := filepath.Join(absolutePath, location.Path)
		err = watchFile(filePath)
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(w, "data: reload\n\n")
		flusher.Flush()
	}
}

func watchFile(filePath string) error {
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

func injectString(data []byte, s ...string) []byte {
	for _, inj := range s {
		data = append(data, inj...)
	}

	return data
}

func injectScript(port string) string {
	s := `<script>var es = new EventSource("http://localhost%s/_livego/reload");es.onmessage = () => {location.reload()}</script>`
	s = fmt.Sprintf(s, port)
	return s
}

func injectTxtBodyStyles() string {
	styles := `<style>body {background: #111;color: #fff;white-space: pre-wrap;}</style>`
	return styles
}
