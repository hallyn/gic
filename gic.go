package main

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v2"
	"github.com/apex/log"
	"github.com/pkg/errors"
)

type Config struct {
	Server struct {
		Port     int    `yaml:"port"`
		Host     string `yaml:"host"`
		SSL      bool   `yaml:"ssl"`
		Password string `yaml:"password"`
		Nick     string `yaml:"nick"`
	} `yaml:"server"`
}

func usage(me string) {
	fmt.Printf("Usage: %s [-f configfile]\n", me)
	fmt.Printf("    configfile is ~/.config/gic/config")
}

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

func serve(cfg Config) {
	fmt.Printf("serving %v\n", cfg)
	password := ""
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
	fmt.Printf("Password is %s\n", password) // DROPME
	if cfg.Server.SSL {
		log.Fatalf("ssl not yet supported")
	}
	if cfg.Server.Port < 1 {
		cfg.Server.Port = 6667
		if cfg.Server.SSL {
			cfg.Server.Port = 6669
		}
	}
	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalf("Error connecting to %s: %v", address, err)
	}
	defer conn.Close()
	log.Infof("Connected to %s", address)
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
