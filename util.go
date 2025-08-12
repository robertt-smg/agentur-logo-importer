package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	glog "github.com/kpango/glg"

	//"pciProxy/computop"

	_ "github.com/go-sql-driver/mysql"
)

func iniSectionName() string {
	b, _ := isDevAndPRQ()
	if b {
		return "development"
	}
	return "production"
}
func isDevAndPRQ() (ret bool, prq string) {

	ret = false
	name := os.Getenv("COMPUTERNAME") + "-DEV"
	//name := "SMG-SCA-W01" + "-DEV"
	parts := strings.Split(name, "-")

	if len(parts) > 2 {
		subpart := parts[2] + "     "
		prefix := subpart[0:1]
		basenum := subpart[3:4]

		if basenum == "6" { // PRQ67-69
			prq = subpart[0:5]
			ret = true
		}
		if prefix != "W" { // Desktops LT-*
			//code = subpart[0:5]
			prq = "DEV01"
			ret = true
		} else {
			prq = subpart[0:5]
		}
		// PROD = PRQ4x-5x
	}
	return ret, strings.Trim(prq, " ")
}

func getIniPath() (path string, dir string) {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		glog.Fatal(err)
	}
	_, prq := isDevAndPRQ()

	path = fmt.Sprintf("%s/%s.%s", dir, prq, "app.ini")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		path = fmt.Sprintf("%s/%s", dir, "app.ini")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			path = fmt.Sprintf("%s/../%s.%s", dir, prq, "app.ini")
			if _, err := os.Stat(path); os.IsNotExist(err) {
				path = fmt.Sprintf("%s/../%s", dir, "app.ini")
			}
		}
	}
	return path, dir
}
