package main

// httpdev is a reverse proxy that checks if your binary
// has changed since the program started, and if it has,
// the system will restart it

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"time"
)

type config struct {
	UserAddress   string
	ServerAddress string
	ProgramPath   string
	Arguments     []string
	LastRestart   os.FileInfo
	backendUrl    *url.URL
	cmd           *exec.Cmd
}

func loadConfig(file string) (*config, error) {
	cfg, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(cfg)
	serverConfig := &config{}
	err = dec.Decode(serverConfig)
	if err != nil {
		return serverConfig, err
	}
	serverConfig.backendUrl, err = url.Parse(serverConfig.ServerAddress)
	return serverConfig, err
}

func startBackend(cfg *config) error {
	fi, err := os.Stat(cfg.ProgramPath)
	if err != nil {
		return err
	}
	cfg.cmd = exec.Command(cfg.ProgramPath, cfg.Arguments...)
	if err != nil {
		return err
	}
	cfg.LastRestart = fi
	go func() {
		log.Printf("starting process...")
		cfg.cmd.Stderr = os.Stderr
		cfg.cmd.Stdout = os.Stdout
		if err := cfg.cmd.Run(); err != nil {
			log.Printf("process returned error: %v", err)
		}
	}()
	time.Sleep(5 * time.Second)
	return nil
}

func wasChanged(cfg *config) bool {
	fi, err := os.Stat(cfg.ProgramPath)
	if err != nil {
		return false
	}
	return fi.ModTime().After(cfg.LastRestart.ModTime())
}

func killBackend(cfg *config) error {
	defer func() {
		// prevent any panic here
		_ = recover()
	}()
	return cfg.cmd.Process.Kill()
}

func setupReverseProxy(cfg *config) chan *checkUpdate {
	ret := make(chan *checkUpdate, 0)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		check := &checkUpdate{done: make(chan signal)}
		ret <- check
		<-check.done
		proxy := httputil.NewSingleHostReverseProxy(cfg.backendUrl)
		proxy.ServeHTTP(w, req)
	})
	return ret
}

type signal struct{}

var (
	doneSignal = signal{}
)

type checkUpdate struct {
	done chan signal
}

func updateBackend(cfg *config, checkReq <-chan *checkUpdate) {
	for c := range checkReq {
		if wasChanged(cfg) {
			err := killBackend(cfg)
			if err != nil {
				log.Fatalf("error killing bakcend. %v", err)
			}
			err = startBackend(cfg)
			if err != nil {
				log.Fatalf("error restarting backend. %v", err)
			}
		}
		c.done <- doneSignal
	}
}

func main() {
	if len(os.Args) == 1 {
		log.Fatalf("you must pass the config file")
	}
	cfg, err := loadConfig(os.Args[1])
	if err != nil {
		log.Fatalf("unable to start. error: %v", err)
	}
	log.Printf("config used: \n%v", cfg)
	err = startBackend(cfg)
	if err != nil {
		log.Fatalf("erro starting backend: error: %v", err)
	}
	checkForUpdate := setupReverseProxy(cfg)
	go updateBackend(cfg, checkForUpdate)
	log.Printf("starting server at %v", cfg.UserAddress)
	err = http.ListenAndServe(cfg.UserAddress, nil)
	if err != nil {
		log.Fatalf("error starting reverse proxy: %v", err)
	}
}
