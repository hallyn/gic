package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"github.com/apex/log"
	"github.com/pkg/errors"
)

type Config struct {
	Server struct {
		Port     int    `yaml:"port"`
		Host     string `yaml:"host"` // the server's hostname
		Name     string `yaml:"name"` // a servername for our purposes
		SSL      bool   `yaml:"ssl"`
		Password string `yaml:"password"`
		Nick     string `yaml:"nick"`
	} `yaml:"server"`
	Config struct {
		OutPath string `yaml:"output"`
		Input struct {
			InType string `yaml:"type"`
			InPath string `yaml:"path"`
		} `yaml:"input"`
	} `yaml:"config"`
}

func usage(me string) {
	fmt.Printf("Usage: %s [-f configfile]\n", me)
	fmt.Printf("    configfile is ~/.config/gic/config")
}

var inFile = os.Stdin
var outFile = os.Stdout // All server output goes here, whether duplicated elsewhere or not

func fromKeyring(k string) (string, error) {
	cmd1 := exec.Command("/usr/bin/keyctl", "request", "user", k)
	res1, err := cmd1.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed requesting user key")
	}
	cmd2 := exec.Command("/usr/bin/keyctl", "print", strings.TrimSpace(string(res1)))
	v, err := cmd2.CombinedOutput()
	if err != nil {
		return "", errors.Wrap(err, "Failed printing user key")
	}
	return string(v), nil
}

func readConn(conn net.Conn, ch chan string) {
	for {
		b := make([]byte, 4096)
		_, err := conn.Read(b)
		if err != nil {
			if err == io.EOF {
				log.Infof("Read EOF from server, exiting")
				os.Exit(0)
			}
			log.Infof("Server error: %v", err)
			os.Exit(1)
		}
		ch<-string(b)
	}
}

func readFile(f *os.File, ch chan string) {
	for {
		b := make([]byte, 4096)
		_, err := f.Read(b)
		if err != nil {
			if err == io.EOF {
				log.Infof("Read EOF from user, exiting")
				os.Exit(0)
			}
			log.Infof("user input error: %v", err)
			os.Exit(1)
		}
		ch<-string(b)
	}
}

func serve(cfg Config) {
	fmt.Printf("serving %v\n", cfg)
	password := ""
	serverName := cfg.Server.Name
	if serverName == "" {
		serverName = cfg.Server.Host
	}
	var err error
	if cfg.Server.Password != "" {
		l := len("keyring ")
		log.Warnf("pass is %s 0:l is %s", cfg.Server.Password, cfg.Server.Password[0:l])
		if cfg.Server.Password[0:l] == "keyring " {
			k := cfg.Server.Password[l:]
			password, err = fromKeyring(k)
			if err != nil {
				log.Fatalf("Failed retrieving password %s: %v", k, err)
			}
		} else {
			password = cfg.Server.Password
		}
	}
	if cfg.Server.Port < 1 {
		cfg.Server.Port = 6667
		if cfg.Server.SSL {
			cfg.Server.Port = 6669
		}
	}
	var conn net.Conn
	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	if cfg.Server.SSL {
		conn, err = tls.Dial("tcp", address, &tls.Config{})
	} else {
		conn, err = net.Dial("tcp", address)
	}
	if err != nil {
		log.Fatalf("Error connecting to %s: %v", address, err)
	}
	defer conn.Close()
	log.Infof("Connected to %s", address)
	if password != "" {
		_, err = conn.Write([]byte(fmt.Sprintf("PASS %s\n", password)))
		if err != nil {
			log.Fatalf("Error writing password: %v", err)
		}
	}
	nick := cfg.Server.Nick
	if nick == "" {
		nick = "gic"
	}
	_, err = conn.Write([]byte(fmt.Sprintf("NICK %s", nick)))
	if err != nil {
		log.Fatalf("Failed sending initial NICK cmd")
	}
	umsg := fmt.Sprintf("USER %s localhost %s :%s", nick, cfg.Server.Host, nick)
	_, err = conn.Write([]byte(umsg))
	if err != nil {
		log.Fatalf("Failed sending initial USER cmd")
	}
	switch cfg.Config.OutPath {
	case "":
	case "default":
		path := os.ExpandEnv("$HOME/.cache/gic")
		path = filepath.Join(path, serverName)
		log.Infof("creating %s\n", path)
		err := os.MkdirAll(path, 0700)
		if err != nil {
			log.Fatalf("Failed creating server dir %s: %v", path, err)
		}
		path = filepath.Join(path, "server.out")
		outFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0700)
		if err != nil {
			log.Fatalf("Failed opening output file %s: %v", path, err)
		}
	default:
		path := cfg.Config.OutPath
		outFile, err = os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0700)
		if err != nil {
			log.Fatalf("Failed opening output file %s: %v", path, err)
		}
	}
	defer outFile.Close()
	log.Infof("Ready to go\n")

	connInCh := make(chan string)
	go readConn(conn, connInCh)
	chSignal := make(chan os.Signal)
	signal.Notify(chSignal, os.Interrupt)
	inFileCh := make(chan string)
	go readFile(inFile, inFileCh)

	select {
		case inLine :=<-inFileCh:
			log.Infof("Command: %s", inLine)
		case servLine :=<-connInCh:
			log.Infof("Server: %s", servLine)
			_, err := outFile.WriteString(servLine)
			if err != nil {
				log.Errorf("Error writing to server file: %v\n", err)
			}
		case <-chSignal:
			log.Infof("Quitting")
			return
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-h" {
		usage(os.Args[0])
		os.Exit(0)
	}
	homedir := os.ExpandEnv("$HOME")
	configFile := filepath.Join(homedir, ".config", "gic", "config")
	if len(os.Args) == 3 && os.Args[1] == "-f" {
		configFile = os.Args[2]

	}
	f, err := os.Open(configFile)
	if err != nil {
		    log.Fatalf("Error opening config file %s: %v", configFile, err)
	}

	var cfg Config
	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		    log.Fatalf("Error parsing config file %s: %v", configFile, err)
	}
	f.Close()

	serve(cfg)
}
