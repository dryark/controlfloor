package main

import (
	"fmt"
	uj "github.com/nanoscopic/ujsonin/v2/mod"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"time"
)

type CDevice struct {
	udid string
}

type ConfigText struct {
	deviceVideo string
}

type Config struct {
	listen       string
	https        bool
	crt          string
	key          string
	auth         string
	root         uj.JNode
	idleTimeout  int
	maxHeight    int
	text         *ConfigText
	disableCache bool
	theme        string
	notes        uj.JNode
}

func (self *Config) String() string {
	https := "false"
	if self.https {
		https = "true"
	}
	return fmt.Sprintf("Listen: %s\nHTTPS: %s\n", self.listen, https)
}

func GetStr(root uj.JNode, path string) string {
	node := root.Get(path)
	if node == nil {
		fmt.Fprintf(os.Stderr, "%s is not set in either config.json or default.json", path)
		os.Exit(1)
	}
	return node.String()
}
func GetBool(root uj.JNode, path string) bool {
	node := root.Get(path)
	if node == nil {
		fmt.Fprintf(os.Stderr, "%s is not set in either config.json or default.json", path)
		os.Exit(1)
	}
	return node.Bool()
}
func GetInt(root uj.JNode, path string) int {
	node := root.Get(path)
	if node == nil {
		fmt.Fprintf(os.Stderr, "%s is not set in either config.json or default.json", path)
		os.Exit(1)
	}
	return node.Int()
}

func NewConfig(configPath string, defaultsPath string) *Config {
	config := Config{
		auth: "builtin",
	}

	root := loadConfig(configPath, defaultsPath)
	config.root = root

	config.listen = GetStr(root, "listen")
	config.https = GetBool(root, "https")
	if config.https {
		config.key = GetStr(root, "key")
		config.crt = GetStr(root, "crt")
	}

	idleTimeout := GetStr(root, "idleTimeout")
	if idleTimeout == "" {
		config.idleTimeout = 0
	} else {
		dur, _ := time.ParseDuration(idleTimeout)
		config.idleTimeout = int(dur.Seconds())
	}

	authNode := config.root.Get("auth")
	if authNode != nil {
		config.auth = GetStr(authNode, "type")
	}

	config.maxHeight = GetInt(root, "video.maxHeight")

	config.text = &ConfigText{
		deviceVideo: GetStr(root, "text.deviceVideo"),
	}

	config.disableCache = GetBool(root, "disableCache")

	config.theme = GetStr(root, "theme")

	config.notes = root.Get("notes")

	return &config
}

func loadConfig(configPath string, defaultsPath string) uj.JNode {
	fh1, serr1 := os.Stat(defaultsPath)
	if serr1 != nil {
		log.WithFields(log.Fields{
			"type":          "err_read_defaults",
			"error":         serr1,
			"defaults_path": defaultsPath,
		}).Fatal("Could not read specified defaults path")
	}
	defaultsFile := defaultsPath
	switch mode := fh1.Mode(); {
	case mode.IsDir():
		defaultsFile = fmt.Sprintf("%s/default.json", defaultsPath)
	}
	content1, err1 := ioutil.ReadFile(defaultsFile)
	if err1 != nil {
		log.Fatal(err1)
	}

	defaults, _ := uj.Parse(content1)

	fh, serr := os.Stat(configPath)
	if serr != nil {
		log.WithFields(log.Fields{
			"type":        "err_read_config",
			"error":       serr,
			"config_path": configPath,
		}).Fatal("Could not read specified config path")
	}
	configFile := configPath
	switch mode := fh.Mode(); {
	case mode.IsDir():
		configFile = fmt.Sprintf("%s/config.json", configPath)
	}
	content, err := ioutil.ReadFile(configFile)
	if err != nil {
		log.Fatal(err)
	}

	root, _ := uj.Parse(content)

	defaults.Overlay(root)
	//defaults.Dump()

	return defaults
}
