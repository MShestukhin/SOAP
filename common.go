package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

func loging(message, path string) {
	if exs, _ := exists(path + "log"); !exs {
		os.Mkdir(path+"log", 0777)
	}
	day := time.Now().Format("2006.01.02")
	nameLog := path + "log/" + day + "log" + ".log"
	f, err := os.OpenFile(nameLog, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		f, err = os.Create(nameLog)
		if err != nil {
			fmt.Println(err)
		}
	}
	defer f.Close()
	f.WriteString(time.Now().Format("2006.01.02 15:04:05") + "   " + message + "\n")
	f.Sync()
}

func checkError(message, path string, err error) {
	if err != nil {
		fmt.Println(err)
		stringGoError := err.Error()
		f, err := os.OpenFile(path+"error.log", os.O_RDWR|os.O_APPEND, 0660)
		if err != nil {
			f, err = os.Create(path + "error.log")
			if err != nil {
				fmt.Println(err)
			}
		}
		defer f.Close()
		f.WriteString(time.Now().Format("2006.01.02 15:04:05") + "   " + message + "   " + stringGoError + "\n")
		f.Sync()
		// log.Fatal(message, err)
	}
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func LoadConfiguration(configPath string) Config {
	var config Config
	configFile, err := os.Open(configPath)
	defer configFile.Close()
	checkError("not config file.", "./", err)
	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&config)
	checkError("not valid json.", "./", err)
	return config
}

func getAllFilesInDir(puth string, need, str []string) []string {
	files, _ := ioutil.ReadDir(puth)
	//fmt.Println(files)
	for _, f := range files {
		if f.IsDir() {
			str = getAllFilesInDir(puth+f.Name()+"/", need, str)
		} else {
			if contains(need, strings.ToLower(path.Ext(f.Name()))) {
				str = append(str, puth+f.Name())
			}
		}
	}
	return str
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
