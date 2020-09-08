package daemon

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestDaemon(t *testing.T) {
	for {
		for i := 0; i < 5; i++ {
			fmt.Println(time.Now().Format("2006-01-02 15:04:05.9999"))
			fmt.Println("hello ", os.Getpid())
			time.Sleep(5 * time.Second)
		}
		panic(errors.New("timeout"))

	}
}
