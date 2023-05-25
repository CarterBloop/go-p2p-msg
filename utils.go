package main

import (
	"os"
	"os/exec"
	"runtime"
)

func clearScreen() {
	var clearCmd *exec.Cmd
	if runtime.GOOS == "windows" {
		clearCmd = exec.Command("cmd", "/c", "cls")
	} else {
		clearCmd = exec.Command("clear")
	}
	clearCmd.Stdout = os.Stdout
	clearCmd.Run()
}