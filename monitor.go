package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"io"
	"io/ioutil"
	"os/exec"
	"time"
	"strings"
	"crypto/md5"
)

var Usage = `
go build 

ls 

./gomonitor -file /Users/chengxinyao/go/src/gojob/app/test.go

`
//检测的文件
var File = flag.String("file","", "检测的文件")

//上一次的md5
var PreMd5 string

//本次md5
var CurrentMd5 string


func GetMd5(str string) string {
    w := md5.New()
    io.WriteString(w, str)
    md5str := fmt.Sprintf("%x", w.Sum(nil))
    return md5str
}


func main(){
	flag.Parse()
	file := *File
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if strings.HasSuffix(file,".py") || strings.HasSuffix(file,".go"){
		PreMd5 = "aaaa"//默认赋值 run函数会检测md5
		//起一个定时器 监听文件 MD5变化 变化则自动调用RunCommand
		t := time.NewTicker(1 * time.Second)
		for {
			select {
			case <-t.C:
				run(file)
			case <- signalCh:
				fmt.Println("kill ")
				t.Stop()
				return 
        	}
	   }
	}else{
		fmt.Println("仅支持.go/.py文件监控")
	}
}



func run(file string){
	fileObj,_ := os.Open(file)
	defer fileObj.Close()
	data,_ := ioutil.ReadAll(fileObj)
	CurrentMd5 = GetMd5(string(data))
	command := "go"
	args := make([]string,0)
	if strings.HasSuffix(file,".py"){
		command = "python3"
	}else{
		args = append(args,"run")
	}
	//上一次的md5 != 本次的md5 则执行
	if PreMd5 != CurrentMd5{
		args = append(args,file)
		ExecuteCommand(command,args...)
		PreMd5 = CurrentMd5
	}
	return 

	
}




func ExecuteCommand(command string,args ...string) error {
	cmd := exec.Command(command,args...)
	stdout, err := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout
	if err != nil {
		return err
	}
	if err = cmd.Start(); err != nil {
		return err
	}
	for {
		tmp := make([]byte, 1024)
		_, err := stdout.Read(tmp)
		fmt.Print(string(tmp))
		if err != nil {
			break
		}
	}
	if err = cmd.Wait(); err != nil {
		return err
	}
	return nil
}