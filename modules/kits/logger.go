package kits

import (
	"errors"
	"fmt"
	"github.com/duke-git/lancet/v2/fileutil"
	"github.com/golang-module/carbon"
	"inner/conf/platform_conf"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

type Log struct {
}

func (*Log) Info(Msg interface{}) {
	Logger(fmt.Sprint(Msg), "info")
}
func (*Log) Error(Msg interface{}) {
	Logger(fmt.Sprint(Msg), "error")
}
func (*Log) Debug(Msg interface{}) {
	Logger(fmt.Sprint(Msg), "debug")
}

func Logger(Msg, MsgType string) {
	var err error
	cf := platform_conf.Setting()
	defer func() {
		if r := recover(); r != nil {
			err = errors.New(fmt.Sprint(r))
		}
		if err != nil {
			fmt.Println(err)
		}
	}()
	if err == nil {
		LogFiles := map[string]string{"info": cf.InfoFile, "error": cf.ErrorFile, "debug": cf.DebugFile}
		wf := WriteLogFile(LogFiles[MsgType])
		Log := log.New(wf, "", log.LstdFlags)
		if MsgType == "info" {
			Log.Println(Msg)
		} else {
			pc, file, line, ok := runtime.Caller(2)
			if ok {
				FuncName := runtime.FuncForPC(pc).Name()
				fu := strings.Split(FuncName, ".")
				if wf != nil {
					Log.Println("/src/" + strings.Split(file, "/src/")[1] + ":" +
						fu[len(fu)-1] + ":" + strconv.Itoa(line) + ":" + Msg)
				}
			}
		}
	}
}
func WriteLogFile(logfile string) *os.File {
	cf := platform_conf.Setting()
	logfile = cf.LogPath + logfile + "." + carbon.Now().ToDateString()
	_ = os.MkdirAll(cf.LogPath, 755)
	if !fileutil.IsExist(logfile) {
		_, _ = os.Create(logfile)
	}
	go func() {
		if carbon.Now().Hour() >= 1 && carbon.Now().Hour() <= 2 {
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println(errors.New(fmt.Sprint(r)))
					}
				}()
				_ = filepath.Walk(cf.LogPath, func(file string, info os.FileInfo, err error) error {
					stat, _ := os.Stat(file)
					if !strings.Contains(file, "."+carbon.Now().ToDateString()) || stat.Size() >= 1000*1000*1000 {
						_ = os.Remove(file)
					}
					return nil
				})
			}()
		}
	}()
	f, err := os.OpenFile(logfile, syscall.O_RDWR|syscall.O_APPEND, 0755)
	if err == nil {
		return f
	}
	return nil
}
