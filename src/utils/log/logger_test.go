package mylog

import (
	"os"
	"testing"
)

func TestNew(t *testing.T) {
	logger := New("F:\\gowork\\src\\luffy\\log\\agent", "test.log", "DEBUG", os.Stdout)
	type Entity struct {
		Code int
		Msg  string
	}
	logger.Info(Entity{500, "床前明月光"})
	logger.Info(Entity{Code: 502, Msg: "地上鞋三双"})

}
