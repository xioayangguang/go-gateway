package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

func GetRootPath() string {
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		dir, _ = os.Getwd()
		return dir
	}
	return dir
}

func GetRootFile() string {
	paths, fileName := filepath.Split(os.Args[0])
	ext := filepath.Ext(fileName)
	abs := strings.TrimSuffix(fileName, ext)
	return paths + abs
}

/*
保存pid
命令key 唯一标识
*/
func SavePidToFile(key string) {
	pid := os.Getpid()
	path := GetRootFile() + "_" + key + ".lock"
	_ = ioutil.WriteFile(path, []byte(fmt.Sprintf("%d", pid)), 0666)
}

/*
删除pid文件
命令key 唯一标识
*/
func DeleteSavePidToFile(key string) {
	path := GetRootFile() + "_" + key + ".lock"
	_ = os.Remove(path)
}

/*
获取pid
命令key 唯一标识
*/
func GetPidForFile(key string) int {
	path := GetRootFile() + "_" + key + ".lock"
	str, err := ioutil.ReadFile(path)
	if err != nil {
		return 0
	}
	pid, err := strconv.Atoi(string(str))
	if err != nil {
		return 0
	}
	return pid
}

func ListenStopSignal(handle func(sig os.Signal)) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigs
		handle(sig)
	}()
}

/*
停止进程
命令key 唯一标识
windows 不发送信号，直接停止
*/
func StopSignal(key string) error {
	pid := GetPidForFile(key)
	if pid == 0 {
		return errors.New("找不到pid记录文件")
	}
	// 通过pid获取子进程
	pro, err := os.FindProcess(pid)
	if err != nil {
		return errors.New("找不到进程信息,文件过期")
	}
	err = pro.Signal(syscall.SIGINT)
	if err != nil {
		if runtime.GOOS == "windows" {
			err = pro.Kill()
			if err != nil {
				return nil
			}
			DeleteSavePidToFile(key)
		}
		return nil
	}
	return nil
}
