package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"time"

	"suah.dev/protect"
)

//go:embed Logseq-linux-x64/resources/app/*
var content embed.FS
var rootFS = "Logseq-linux-x64/resources/app/"

var (
	dump   bool
	listen string
)

func init() {
	flag.StringVar(&listen, "http", "127.0.0.1:8080", "Listen on")
	flag.BoolVar(&dump, "dump", false, "Dump Logseq assets to disk (./logseq).")
	flag.Parse()

	_ = protect.Unveil(path.Join(os.Getenv("HOME"), "Notes"), "rwc")
	_ = protect.Unveil(os.TempDir(), "rwc")
	_ = protect.Pledge("stdio wpath rpath cpath inet dns")
}

func httpLog(r *http.Request) {
	n := time.Now()
	fmt.Printf("%s (%s) [%s] \"%s %s\" %03d\n",
		r.RemoteAddr,
		n.Format(time.RFC822Z),
		r.Method,
		r.URL.Path,
		r.Proto,
		r.ContentLength,
	)
}

func dumpFS(dest string, entries []fs.DirEntry) {
	var localRoot = rootFS

	err := os.Mkdir(path.Join(dest, localRoot), 0750)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for _, e := range entries {
		fp := path.Join(localRoot, e.Name())

		if e.IsDir() {
			rootFS = fp

			newDir, err := content.ReadDir(rootFS)
			if err != nil {
				log.Fatalln(err)
			}
			dumpFS(dest, newDir)
		} else {
			fh, err := content.Open(fp)
			if err != nil {
				log.Fatalln(err)
			}

			defer fh.Close()

			nfp := path.Join(dest, fp)
			nfh, err := os.Create(filepath.Clean(nfp))
			if err != nil {
				log.Fatalln(err)
			}

			_, err = nfh.ReadFrom(fh)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Printf("\t%s\n", nfp)
			err = nfh.Close()
			if err != nil {
				log.Fatalln(err)
			}
		}
	}
}

func main() {
	if dump {
		tmpDir, err := os.MkdirTemp("", "logseq")
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Printf("Dumping files to %q\n", tmpDir)
		dir, err := content.ReadDir(rootFS)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		dumpFS(tmpDir, dir)
		os.Exit(0)
	}

	fileServer := http.FileServer(http.FS(content))

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		r.URL.Path = fmt.Sprintf("/%s%s", rootFS, r.URL.Path)

		httpLog(r)
		fileServer.ServeHTTP(w, r)
	})

	s := http.Server{
		Handler: mux,
	}

	lis, err := net.Listen("tcp", listen)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Logseq can be reached at: http://%s", listen)
	log.Panic(s.Serve(lis))
}
