package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestUpload(t *testing.T) {

	cfg := getIniCfg(nil)
	app := new(cfg)
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

	l := logo{}
	l.Country = "DE"
	l.Serverpath = "/test"
	l.Agency = "000306"
	l.Localpath = dir + "/logos/123456.jpg"
	l.Downloaded = true

	//create logos  map
	logos := make(map[string]*logo)
	logos[l.Agency] = &l

	app.uploadLogos(logos)
}
func TestUpload_Solamento(t *testing.T) {

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

	app.uploadLogos(logos)
}
