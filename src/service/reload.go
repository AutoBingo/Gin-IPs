/*
ctrl + c	SIGINT	强制进程结束
ctrl + z	SIGTSTP	任务中断，进程挂起
ctrl + \	SIGQUIT	进程结束 和 dump core
ctrl + d		EOF
SIGHUP	终止收到该信号的进程。若程序中没有捕捉该信号，当收到该信号时，进程就会退出（常用于 重启、重新加载进程）

1) SIGHUP	 2) SIGINT	 3) SIGQUIT	 4) SIGILL	 5) SIGTRAP
 6) SIGABRT	 7) SIGBUS	 8) SIGFPE	 9) SIGKILL	10) SIGUSR1
11) SIGSEGV	12) SIGUSR2	13) SIGPIPE	14) SIGALRM	15) SIGTERM
16) SIGSTKFLT	17) SIGCHLD	18) SIGCONT	19) SIGSTOP	20) SIGTSTP
21) SIGTTIN	22) SIGTTOU	23) SIGURG	24) SIGXCPU	25) SIGXFSZ
26) SIGVTALRM	27) SIGPROF	28) SIGWINCH	29) SIGIO	30) SIGPWR
31) SIGSYS	34) SIGRTMIN	35) SIGRTMIN+1	36) SIGRTMIN+2	37) SIGRTMIN+3
38) SIGRTMIN+4	39) SIGRTMIN+5	40) SIGRTMIN+6	41) SIGRTMIN+7	42) SIGRTMIN+8
43) SIGRTMIN+9	44) SIGRTMIN+10	45) SIGRTMIN+11	46) SIGRTMIN+12	47) SIGRTMIN+13
48) SIGRTMIN+14	49) SIGRTMIN+15	50) SIGRTMAX-14	51) SIGRTMAX-13	52) SIGRTMAX-12
53) SIGRTMAX-11	54) SIGRTMAX-10	55) SIGRTMAX-9	56) SIGRTMAX-8	57) SIGRTMAX-7
58) SIGRTMAX-6	59) SIGRTMAX-5	60) SIGRTMAX-4	61) SIGRTMAX-3	62) SIGRTMAX-2
63) SIGRTMAX-1	64) SIGRTMAX

在 kill 服务时，使用 kill -12 pid / kill -USR2 pid将服务停止。

服务在收到信号量 12 (SIGUSR2) 后， 不再处理新请求，

将已开始的请求处理完成，

将 标准输出、错误输出 和 socket 的描述符转交给之后新启动的程序

*/
package service

import (
	"Gin-IPs/src/dao"
	"Gin-IPs/src/utils/database/mongodb"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func (server *Server) Listen(graceful bool) error {
	addr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	httpServer := &http.Server{
		Addr:    addr,
		Handler: server.Router,
	}
	// 判断是否为 reload
	var err error
	if graceful {
		server.Logger.Info("listening on the existing file descriptor 3")
		//子进程的 0 1 2 是预留给 标准输入 标准输出 错误输出
		//因此传递的socket 描述符应该放在子进程的 3
		f := os.NewFile(3, "")
		// 获取 上个服务程序的 socket 的描述符
		server.Listener, err = net.FileListener(f)
	} else {
		server.Logger.Info("listening on a new file descriptor")
		server.Listener, err = net.Listen("tcp", httpServer.Addr)
		server.Logger.Infof("Actual pid is %d\n", syscall.Getpid())
	}
	if err != nil {
		server.Logger.Error(err)
		return err
	}

	go func() {
		// 开启服务
		if err := httpServer.Serve(server.Listener); err != nil && err != http.ErrServerClosed {
			err = errors.New(fmt.Sprintf("listen error:%v\n", err))
			server.Logger.Fatal(err) // 报错退出
		}
	}()
	return server.HandlerSignal(httpServer)
}

func (server *Server) HandlerSignal(httpServer *http.Server) error {
	sign := make(chan os.Signal)
	signal.Notify(sign, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR2)
	for {
		// 接收信号量
		sig := <-sign
		server.Logger.Infof("Signal receive: %v\n", sig)
		ctx, _ := context.WithTimeout(context.Background(), time.Second*10)
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			// 关闭服务
			server.Logger.Info("Shutdown Api Server")
			signal.Stop(sign) // 停止通道
			if err := httpServer.Shutdown(ctx); err != nil {
				err = errors.New(fmt.Sprintf("Shutdown Api Server Error: %s", err))
				return err
			}
			if err := destroyMgoPool(); err != nil {
				return err
			}
			return nil
		case syscall.SIGUSR2:
			server.Logger.Info("Reload Api Server")
			// 启动新服务
			if err := server.Reload(); err != nil {
				server.Logger.Errorf("Reload Api Server Error: %s", err)
				continue
			}
			// 关闭旧服务
			if err := httpServer.Shutdown(ctx); err != nil {
				err = errors.New(fmt.Sprintf("Shutdown Api Server Error: %s", err))
				return err
			}
			if err := destroyMgoPool(); err != nil {
				return err
			}
			server.Logger.Info("Reload Api Server Successful")
			return nil
		}
	}
}

func (server *Server) Reload() error {
	tl, ok := server.Listener.(*net.TCPListener)
	if !ok {
		return errors.New("listener is not tcp listener")
	}

	f, err := tl.File()
	if err != nil {
		return err
	}

	// 命令行启动新程序
	args := []string{"-graceful"}
	cmd := exec.Command(os.Args[0], args...)
	cmd.Stdout = os.Stdout         //  1
	cmd.Stderr = os.Stderr         //  2
	cmd.ExtraFiles = []*os.File{f} //  3
	if err := cmd.Start(); err != nil {
		return err
	}
	server.Logger.Infof("Forked New Pid %v: \n", cmd.Process.Pid)
	return nil
}

func destroyMgoPool() error {
	if err := mongodb.DestroyPool(dao.ModelClient.MgoPool); err != nil {
		return err
	}
	return nil
}
