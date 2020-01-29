package main

import (
    "bytes"
    "fmt"
    "log"
    "os/exec"
    "strings"
    "os"
)

const ShellToUse = "bash"

func Shellout(command string) (string) {
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd := exec.Command(ShellToUse, "-c", command)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()

    if err != nil {
        log.Printf("error: %v\n", err)
        fmt.Println("--- Error Trace ---")
        fmt.Println(stderr.String())
        os.Exit(1)
    }

    return stdout.String()
}

func main() {
    gitLogCommand := "git log --format=%B%H----DELIMITER----"
    out := Shellout(gitLogCommand)
    commitsArray := strings.Split(out, "----DELIMITER----\n")
    commitDict := make(map[string]string)
    for _, commit := range commitsArray {
        commitPart := strings.Split(commit, "\n")
        if len(commitPart) == 2 {
            commitDict[string(commitPart[1])] = string(commitPart[0])
        }
    }

    for sha, message := range commitDict {
        fmt.Println(sha, message)
    }






    //fmt.Println("--- stdout ---")
    //fmt.Println(out)
}
