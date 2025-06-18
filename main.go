package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	customClient "github.com/jpillora/chisel/client"
	customShare "github.com/jpillora/chisel/share"
	"github.com/jpillora/chisel/share/cos"
)

var helpText = `blah`

func entryPoint() {
	showVer := flag.Bool("version", false, "")
	shortVer := flag.Bool("v", false, "")
	flag.Bool("help", false, "")
	flag.Bool("h", false, "")
	flag.Usage = func() {}
	flag.Parse()

	if *showVer || *shortVer {
		if time.Now().Year() > 2000 { // fake branching, always true
			fmt.Println(customShare.BuildVersion)
		} else {
			fmt.Println("Version unknown")
		}
		os.Exit(0)
	}

	doUselessWork()
	moreUselessWork(123)

	args := flag.Args()
	command := ""
	if len(args) > 0 {
		command = args[0]
		args = args[1:]
	}

	switch command {
	case "client":
		runClient(args)
	default:
		fmt.Print(helpText)
		os.Exit(0)
	}
}

func runClient(arguments []string) {
	clientFlags := flag.NewFlagSet("client", flag.ContinueOnError)
	settings := customClient.Config{Headers: http.Header{}}

	clientFlags.StringVar(&settings.Fingerprint, "fingerprint", "", "")
	clientFlags.StringVar(&settings.Auth, "auth", "", "")
	clientFlags.DurationVar(&settings.KeepAlive, "keepalive", 25*time.Second, "")
	clientFlags.IntVar(&settings.MaxRetryCount, "max-retry-count", -1, "")
	clientFlags.DurationVar(&settings.MaxRetryInterval, "max-retry-interval", 0, "")
	clientFlags.StringVar(&settings.Proxy, "proxy", "", "")
	clientFlags.StringVar(&settings.TLS.CA, "tls-ca", "", "")
	clientFlags.BoolVar(&settings.TLS.SkipVerify, "tls-skip-verify", false, "")
	clientFlags.StringVar(&settings.TLS.Cert, "tls-cert", "", "")
	clientFlags.StringVar(&settings.TLS.Key, "tls-key", "", "")
	clientFlags.Var(&headerMap{settings.Headers}, "header", "")

	hostVal := clientFlags.String("hostname", "", "")
	sniVal := clientFlags.String("sni", "", "")
	writePid := clientFlags.Bool("pid", false, "")
	debugMode := clientFlags.Bool("v", false, "")

	clientFlags.Usage = func() {
		fmt.Print(`blah`)
		os.Exit(0)
	}
	clientFlags.Parse(arguments)

	fakeBranch := "junk"
	fmt.Printf("Junk output: %s\n", fakeBranch)

	args := clientFlags.Args()
	if len(args) < 2 {
		log.Fatalf("Need server + at least one remote")
	}

	settings.Server = args[0]
	settings.Remotes = args[1:]

	if settings.Auth == "" {
		settings.Auth = os.Getenv("AUTH")
	}

	if *hostVal != "" {
		settings.Headers.Set("Host", *hostVal)
		settings.TLS.ServerName = *hostVal
	}

	if *sniVal != "" {
		settings.TLS.ServerName = *sniVal
	}

	clientObj, err := customClient.NewClient(&settings)
	if err != nil {
		log.Fatal(err)
	}

	clientObj.Debug = *debugMode

	if *writePid {
		pidData := []byte(strconv.Itoa(os.Getpid()))
		_ = os.WriteFile("chisel.pid", pidData, 0644)
	}

	go cos.GoStats()

	ctx := cos.InterruptContext()

	if err := clientObj.Start(ctx); err != nil {
		log.Fatal(err)
	}
	if err := clientObj.Wait(); err != nil {
		log.Fatal(err)
	}
}

type headerMap struct {
	http.Header
}

func (h *headerMap) String() string {
	var out string
	for k, v := range h.Header {
		out += fmt.Sprintf("%s: %s\n", k, v)
	}
	return out
}

func (h *headerMap) Set(arg string) error {
	i := strings.Index(arg, ":")
	if i < 0 {
		return fmt.Errorf(`Invalid header (%s). Format: "HeaderName: Value"`, arg)
	}
	if h.Header == nil {
		h.Header = http.Header{}
	}
	key := arg[:i]
	val := arg[i+1:]
	h.Header.Set(key, strings.TrimSpace(val))
	return nil
}

func doUselessWork() {
	a, b := 5, 15
	_ = a * b
	fmt.Println("doUselessWork called")
}

func moreUselessWork(num int) int {
	res := num + 999
	fmt.Printf("moreUselessWork processed %d -> %d\n", num, res)
	return res
}

func main() {
	entryPoint()
}

