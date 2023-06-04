package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"

	yaml "gopkg.in/yaml.v2"
)

type Host struct {
	HostName string `yaml:"hostName"`
	BootIP   string `yaml:"bootIP"`
	HwAddr   string `yaml:"hwAddr"`
	FileName string `yaml:"fileName"`
}

type Config struct {
	Address string   `yaml:"address"`
	Port    string   `yaml:"port"`
	Dir     string   `yaml:"dir"`
	Files   []string `yaml:"files"`
}

var hosts []Host
var config Config

func hostCloudConfigHandler(w http.ResponseWriter, req *http.Request) {
	userIP, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		log.Printf("Could't split Host:Port: %+v", err)
		fmt.Fprintf(w, http.StatusText(http.StatusNotFound))
		return
	}

	host_i := -1
	for i, host := range hosts {
		if host.BootIP == userIP {
			host_i = i
			break
		}
	}
	if host_i == -1 {
		log.Printf("Not found IPaddress[%v]", userIP)
		fmt.Fprintf(w, http.StatusText(http.StatusNotFound))
		return
	}
	http.ServeFile(w, req, config.Dir+hosts[host_i].FileName)
}

func main() {

	buf, err := ioutil.ReadFile("/usr/local/bootconf/config.yml")
	if err != nil {
		log.Printf("Could't read the config.yml: %+v", err)
		return
	}

	err = yaml.Unmarshal(buf, &config)
	if err != nil {
		log.Printf("Could't YAML convert : %+v", err)
		return
	}

	if config.Dir[len(config.Dir)-1:] != "/" {
		config.Dir += "/"
	}

	buf, err = ioutil.ReadFile(config.Dir + "hosts.yml")
	if err != nil {
		log.Printf("Could't read the hosts.yml : %+v", err)
		return
	}
	err = yaml.Unmarshal(buf, &hosts)
	if err != nil {
		log.Printf("Could't YAML convert : %+v", err)
		return
	}

	http.HandleFunc("/boot.ipxe", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, config.Dir+"boot.ipxe")
	})

	http.HandleFunc("/initial-cloud-config", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, config.Dir+"initial-cloud-config")
	})

	http.HandleFunc("/initrd", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, config.Dir+"initrd")
	})

	http.HandleFunc("/vmlinuz", func(w http.ResponseWriter, req *http.Request) {
		http.ServeFile(w, req, config.Dir+"vmlinuz")
	})

	http.HandleFunc("/host-cloud-config", hostCloudConfigHandler)

	err = http.ListenAndServe(config.Address+":"+config.Port, nil)
	if err != nil {
		log.Fatal("Unable to listen on port", err)
	}

}
