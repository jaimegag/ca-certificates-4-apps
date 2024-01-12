package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
)

func main() {

    app := "ls"
    arg0 := "-lart"
    arg1 := "/etc/ssl/certs"
    cmd := exec.Command(app, arg0, arg1)
    stdout, cerr := cmd.Output()
    if cerr != nil {
        fmt.Println(cerr.Error())
    }
    fmt.Println(string(stdout))

	if len(os.Args) != 2 {
		fmt.Println("USAGE: ca-certificates <url>")
		os.Exit(1)
	}
	_, err := http.Head(os.Args[1])
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(2)
	}
	fmt.Println("SUCCESS!")
}