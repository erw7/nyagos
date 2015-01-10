package completion

import (
	"os"
	"path"
	"strings"

	"../alias"
	"../commands"
)

func listUpAllExecutableOnPath() []string {
	list := make([]string, 0, 100)
	pathEnv := os.Getenv("PATH")
	dirList := strings.Split(pathEnv, ";")
	for _, dir1 := range dirList {
		dirHandle, dirErr := os.Open(dir1)
		if dirErr != nil {
			continue
		}
		defer dirHandle.Close()
		files, filesErr := dirHandle.Readdir(0)
		if filesErr != nil {
			continue
		}
		for _, file1 := range files {
			if file1.IsDir() {
				continue
			}
			name := file1.Name()
			if isExecutable(name) {
				list = append(list, path.Base(name))
			}
		}
	}
	return list
}

func listUpCurrentAllExecutable(str string) ([]string, error) {
	listTmp, listErr := listUpFiles(str)
	if listErr != nil {
		return nil, listErr
	}
	list := make([]string, 0)
	for _, fname := range listTmp {
		if strings.HasSuffix(fname, "/") || strings.HasSuffix(fname, "\\") || isExecutable(fname) {
			list = append(list, fname)
		}
	}
	return list, nil
}

func listUpCommands(str string) ([]string, error) {
	list, listErr := listUpCurrentAllExecutable(str)
	if listErr != nil {
		return nil, listErr
	}
	strUpr := strings.ToUpper(str)
	for _, name := range listUpAllExecutableOnPath() {
		name1Upr := strings.ToUpper(name)
		if strings.HasPrefix(name1Upr, strUpr) {
			list = append(list, name)
		}
	}
	for name, _ := range commands.BuildInCommand {
		name1Upr := strings.ToUpper(name)
		if strings.HasPrefix(name1Upr, strUpr) {
			list = append(list, name)
		}
	}
	for name, _ := range alias.Table {
		name1Upr := strings.ToUpper(name)
		if strings.HasPrefix(name1Upr, strUpr) {
			list = append(list, name)
		}
	}

	// remove dupcalites
	uniq := make([]string, 0)
	lastone := ""
	for _, cur := range list {
		if cur != lastone {
			uniq = append(uniq, cur)
		}
		lastone = cur
	}
	return uniq, nil
}
