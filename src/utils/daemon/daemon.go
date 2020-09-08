/*
Linux Mac 下运行
守护进程是生存期长的一种进程。它们独立于控制终端并且周期性的执行某种任务或等待处理某些发生的事件。
守护进程必须与其运行前的环境隔离开来。这些环境包括未关闭的文件描述符、控制终端、会话和进程组、工作目录以及文件创建掩码等。这些环境通常是守护进程从执行它的父进程（特别是shell）中继承下来的。
本程序只fork一次子进程，fork第二次主要目的是防止进程再次打开一个控制终端（不是必要的）。因为打开一个控制终端的前台条件是该进程必须是会话组长，再fork一次，子进程ID != sid（sid是进程父进程的sid），所以也无法打开新的控制终端
*/
package daemon

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func InitProcess() {
	if syscall.Getppid() == 1 {
		if err := os.Chdir("./"); err != nil {
			panic(err)
		}
		syscall.Umask(0)
		return
	}
	fmt.Println("go daemon!!!")
	fp, err := os.OpenFile("daemon.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = fp.Close()
	}()
	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Stdout = fp
	cmd.Stderr = fp
	cmd.Stdin = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}

	if err := cmd.Start(); err != nil {
		panic(err)
	}

	_, _ = fp.WriteString(fmt.Sprintf(
		"[PID] %d Start At %s\n", cmd.Process.Pid, time.Now().Format("2006-01-02 15:04:05")))
	os.Exit(0)
}
