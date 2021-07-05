package main

import (
	"bufio"
	"crypto/md5"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"

	_ "net/http/pprof"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/internal/utils"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
)

type binds []string

func (b binds) String() string {
	return strings.Join(b, ",")
}

func (b *binds) Set(v string) error {
	*b = strings.Split(v, ",")
	return nil
}

// Size is needed by the /demo/upload handler to determine the size of the uploaded file
type Size interface {
	Size() int64
}

// See https://en.wikipedia.org/wiki/Lehmer_random_number_generator
func generatePRData(l int) []byte {
	res := make([]byte, l)
	seed := uint64(1)
	for i := 0; i < l; i++ {
		seed = seed * 48271 % 2147483647
		res[i] = byte(seed)
	}
	return res
}

func setupHandler(www string) http.Handler {
	mux := http.NewServeMux()

	if len(www) > 0 {
		mux.Handle("/", http.FileServer(http.Dir(www)))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			fmt.Printf("%#v\n", r)
			const maxSize = 1 << 30 // 1 GB
			num, err := strconv.ParseInt(strings.ReplaceAll(r.RequestURI, "/", ""), 10, 64)
			if err != nil || num <= 0 || num > maxSize {
				w.WriteHeader(400)
				return
			}
			w.Write(generatePRData(int(num)))
		})
	}

	// accept file uploads and return the MD5 of the uploaded file
	// maximum accepted file size is 1 GB
	mux.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			err := r.ParseMultipartForm(1 << 30) // 1 GB
			if err == nil {
				var file multipart.File
				file, _, err = r.FormFile("uploadfile")
				if err == nil {
					var size int64
					if sizeInterface, ok := file.(Size); ok {
						size = sizeInterface.Size()
						b := make([]byte, size)
						file.Read(b)
						md5 := md5.Sum(b)
						fmt.Fprintf(w, "%x", md5)
						return
					}
					err = errors.New("couldn't get uploaded file size")
				}
			}
			utils.DefaultLogger.Infof("Error receiving upload: %#v", err)
		}
		io.WriteString(w, `<html><body><form action="/demo/upload" method="post" enctype="multipart/form-data">
				<input type="file" name="uploadfile"><br>
				<input type="submit">
			</form></body></html>`)
	})

	return mux
}

func main() {
	// defer profile.Start().Stop()
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	// runtime.SetBlockProfileRate(1)

	verbose := flag.Bool("v", false, "verbose")
	bs := binds{}
	flag.Var(&bs, "bind", "bind to")
	www := flag.String("www", "", "www data")
	tcp := flag.Bool("tcp", false, "also listen on TCP")
	enableQlog := flag.Bool("qlog", false, "output a qlog (in the same directory)")
	certFile := flag.String("cert", "", "cert path")
	keyFile := flag.String("key", "", "key path")
	flag.Parse()

	logger := utils.DefaultLogger

	if *verbose {
		logger.SetLogLevel(utils.LogLevelDebug)
	} else {
		logger.SetLogLevel(utils.LogLevelInfo)
	}
	logger.SetLogTimeFormat("")

	if *certFile == "" {
		logger.Errorf("cert argument is required\n")
		os.Exit(1)
	} else if _, err := os.Stat(*certFile); os.IsNotExist(err) {
		logger.Errorf("cert file %s does not exist\n", *certFile)
		os.Exit(1)
	}

	if *keyFile == "" {
		logger.Errorf("key argument is required\n")
		os.Exit(1)
	} else if _, err := os.Stat(*keyFile); os.IsNotExist(err) {
		logger.Errorf("key file %s does not exist\n", *keyFile)
		os.Exit(1)
	}

	if len(bs) == 0 {
		bs = binds{"localhost:6121"}
	}

	handler := setupHandler(*www)
	quicConf := &quic.Config{}
	if *enableQlog {
		quicConf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("server_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})
	}

	var wg sync.WaitGroup
	wg.Add(len(bs))
	for _, b := range bs {
		bCap := b
		go func() {
			var err error

			logger.Infof("Start server on %s\n", bCap)
			if *tcp {
				err = http3.ListenAndServe(bCap, *certFile, *keyFile, handler)
			} else {
				server := http3.Server{
					Server:     &http.Server{Handler: handler, Addr: bCap},
					QuicConfig: quicConf,
				}
				err = server.ListenAndServeTLS(*certFile, *keyFile)
			}
			if err != nil {
				fmt.Println(err)
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
