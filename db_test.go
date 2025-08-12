package main

import (
	"os"
	"path/filepath"
	"testing"

	glog "github.com/kpango/glg"
)

func TestSubagent_Solamento(t *testing.T) {

	cfg := getIniCfg(nil)
	app := new(cfg)
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	l := logo{}
	l.Country = "DE"
	l.Serverpath = "/test"
	l.Agency = "103770"
	l.Localpath = dir + "/logos/103770.jpg"
	l.Downloaded = true

	//create logos  map
	logos := make(map[string]*logo)
	logos[l.Agency] = &l

	//app.uploadLogos(logos)

	ret := app.loadSubagents(&l)

	if ret != nil {
		if len(ret) > 1 {
			glog.Println("Success")
		}
	}

}
