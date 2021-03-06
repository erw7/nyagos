package frame

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/zetamatta/nyagos/dos"
)

var Version string

func VersionOrStamp() string {
	if Version != "" {
		return Version
	} else {
		return "snapshot"
	}
}

func LoadScripts(shellEngine func(string) error,
	langEngine func(string) ([]byte, error)) error {

	exeName, err := os.Executable()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	exeFolder := filepath.Dir(exeName)
	nyagos_d := filepath.Join(exeFolder, "nyagos.d")
	files, err := ioutil.ReadDir(nyagos_d)
	if err == nil {
		for _, finfo1 := range files {
			name1 := finfo1.Name()
			path1 := filepath.Join(nyagos_d, name1)
			name1_ := strings.ToLower(name1)

			var err error
			if strings.HasSuffix(name1_, ".lua") {
				_, err = langEngine(path1)
			} else if strings.HasSuffix(name1_, ".ny") {
				err = shellEngine(path1)
			}
			if err != nil {
				fmt.Fprintf(os.Stderr, "%s: %s\n", name1, err.Error())
			}
		}
	}
	fname := filepath.Join(exeFolder, ".nyagos")
	if _, err := os.Stat(fname); err == nil {
		if _, err := langEngine(fname); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
	barNyagos(shellEngine, exeFolder)
	if err := dotNyagos(langEngine); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
	barNyagos(shellEngine, dos.GetHome())
	return nil
}

func dotNyagos(langEngine func(string) ([]byte, error)) error {
	dot_nyagos := filepath.Join(dos.GetHome(), ".nyagos")
	dotStat, err := os.Stat(dot_nyagos)
	if err != nil {
		return nil
	}
	cachePath := filepath.Join(AppDataDir(), runtime.GOARCH+".nyagos.luac")
	cacheStat, err := os.Stat(cachePath)
	if err == nil && cacheStat.Size() != 0 &&
		!dotStat.ModTime().After(cacheStat.ModTime()) {
		_, err = langEngine(cachePath)
		return err
	}
	chank, err := langEngine(dot_nyagos)
	if err != nil {
		return err
	}
	if chank != nil {
		return ioutil.WriteFile(cachePath, chank, os.FileMode(0644))
	} else {
		return nil
	}
}

func barNyagos(shellEngine func(string) error, folder string) {
	bar_nyagos := filepath.Join(folder, "_nyagos")
	fd, err := os.Open(bar_nyagos)
	if err != nil {
		return
	}
	err = shellEngine(bar_nyagos)
	if err != nil {
		fmt.Fprint(os.Stderr, err.Error())
	}
	fd.Close()
}

var appdatapath_ string

func AppDataDir() string {
	if appdatapath_ == "" {
		appdatapath_ = filepath.Join(os.Getenv("APPDATA"), "NYAOS_ORG")
		os.Mkdir(appdatapath_, 0777)
	}
	return appdatapath_
}
