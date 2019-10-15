package ffmpeg

import (
	"os/exec"
	"time"
)

type location struct {
	lat, lng float64
}

type VideoOutput struct {
	CreationTime time.Time `tag:"creation_time"`
	Location     location  `tag:"location"`
}

func LoadVideo() {}

func execCmd() error {
	var err error
	cmd := exec.Command("name")
	err = cmd.Run()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
