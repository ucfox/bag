package pconfig

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//对pylon框架配置的解析, pylon config(pconfig)

var conf map[string]string
var listfile []string
var configPath string

func init() {
	conf = make(map[string]string)
}

func Get(env string) string {
	if value, ok := conf[env]; ok {
		return value
	}

	return ""
}

func Init(path string) {
	_, err := os.Stat(path) //判断目录是否存在
	if err != nil {
		panic(err)
	}

	getFileList(path)
	if len(listfile) != 1 {
		panic("config file num exception")
	}

	configPath = listfile[0]
	ReadLine(configPath)
}

func Listfunc(path string, f os.FileInfo, err error) error {
	var strRet string

	if f == nil {
		panic(err)
	}
	if f.IsDir() {
		return nil
	}

	strRet += path //+ "\r\n"

	//用strings.HasSuffix(src, suffix)//判断src中是否包含 suffix结尾
	ok := strings.HasSuffix(strRet, "golang.env")
	if ok {
		listfile = append(listfile, strRet) //将目录push到listfile []string中
	}

	return nil
}

func getFileList(path string) string {
	//var strRet string
	err := filepath.Walk(path, Listfunc) //

	if err != nil {
		panic(fmt.Sprintf("filepath.Walk() returned %v\n", err))
	}

	return ""
}

func ReadLine(path string) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0660)

	if err != nil {
		panic(err)
	}

	err = cat(bufio.NewScanner(f))
	if err != nil {
		panic(err)
	}
}

func cat(scanner *bufio.Scanner) error {
	//var slice []string
	for scanner.Scan() {
		text := scanner.Text()
		/*配置中含＝号则出错  如： MONGO_HOST     :   "mongodb://test.com:7107/?replicaSet=repl_7107&maxPoolSize=15"
		           会忽略第二个＝后的配置
				  slice = strings.Split(scanner.Text(), "=")
				  conf[strings.TrimSpace(slice[0])] = strings.Trim(strings.TrimSpace(slice[1]), "\"")
		*/
		equalIndex := strings.Index(text, "=")
		key := strings.TrimSpace(text[:equalIndex])
		value := strings.Trim(strings.TrimSpace(text[equalIndex+1:]), "\"")
		conf[key] = value
	}
	return scanner.Err()
}
