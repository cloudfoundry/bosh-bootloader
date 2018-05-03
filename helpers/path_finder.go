package helpers

import "os/exec"

type PathFinder struct{}

func NewPathFinder() PathFinder {
	return PathFinder{}
}

func (p PathFinder) CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}
