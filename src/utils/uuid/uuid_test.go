package uuid

import (
	"github.com/sony/sonyflake"
	"log"
	"testing"
)

func TestWorker_GetId(t *testing.T) {
	work, _ := NewSnowWorker(500)
	id := work.GetId()
	t.Log(id)
	t.Log(len(id))

	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	fid, err := flake.NextID()
	if err != nil {
		log.Fatalf("flake.NextID() failed with %s\n", err)
	}
	// Note: this is base16, could shorten by encoding as base62 string
	t.Logf("github.com/sony/sonyflake:   %x\n", fid)
}
