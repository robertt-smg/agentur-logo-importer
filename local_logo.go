package main

import (
	"os"
	"path/filepath"
	"strings"

	glog "github.com/kpango/glg"
)

// getFilesFromLocal mimics getFilesFromServer but for a local folder
func (t *this) getFilesFromLocal(localPath string) map[string]*logo {
	logos := make(map[string]*logo)
	format := t.cfg.Section("file").Key("format").String()
	validExt := strings.ToLower(format)

	_ = filepath.Walk(localPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}
		if strings.Contains(path, "ignore") {
			return nil
		}
		if strings.Contains(path, "convert") {
			return nil
		}
		if strings.Contains(path, "done") {
			return nil
		}
		fileExt := strings.ToLower(filepath.Ext(path))
		if !strings.Contains(validExt, fileExt) {
			return nil
		}
		// mimic SFTP folder logic
		parts := strings.Split(filepath.ToSlash(filepath.Dir(path)), "/")
		agency := "-"
		filename := filepath.Base(path)
		if len(parts) > 1 {
			if len(parts) >= 2 {
				agency = parts[len(parts)-1]
				fparts := strings.Split(filename, ".")
				if len(fparts) > 1 {
					agency = fparts[0]
				}
			}
			l := logo{}
			l.Country = parts[len(parts)-2]
			l.Serverpath = path
			l.Agency = agency
			logos[agency] = &l
		}
		return nil
	})
	return logos
}

// copyLocalFilesToLokal mimics copyServerFilesToLokal but for local files
func (t *this) copyLocalFilesToLokal(logos map[string]*logo, localPath string) {
	destPath := t.cfg.Section("paths").Key("lokalPath").String()

	tmpPath := filepath.Join(localPath, "convert")
	err := os.MkdirAll(tmpPath, 0755)
	if err != nil {
		glog.Fatal(err)
		return
	}
	donePath := filepath.Join(localPath, "done")
	err = os.MkdirAll(donePath, 0755)
	if err != nil {
		glog.Fatal(err)
		return
	}
	for _, logo := range logos {
		srcFile, err := os.Open(logo.Serverpath)
		if err != nil {
			logo.Err = err
			continue
		}

		_, err = os.Stat(localPath)
		if err != nil && os.IsNotExist(err) {
			os.MkdirAll(localPath, os.ModePerm)
		}
		filename := filepath.Join(tmpPath, logo.Agency+".jpg")
		dstFile, err := os.Create(filename)
		if err != nil {
			logo.Err = err
			continue
		}
		defer dstFile.Close()
		_, err = dstFile.ReadFrom(srcFile)
		if err != nil {
			logo.Err = err
		} else {
			srcFile.Close()
			logo.Localpath = filename
			logo.Destpath = filepath.Join(destPath, logo.Agency+".jpg")
			logo.Downloaded = true
			err = os.Rename(logo.Serverpath, filepath.Join(donePath, logo.Agency+".jpg"))
		}

		if err != nil {
			glog.Fatal(err)
			return
		}
	}
}
