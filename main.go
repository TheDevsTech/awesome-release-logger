package main

import (
    "bytes"
    "fmt"
    "log"
    "os/exec"
    "strings"
    "os"
    "time"
)

const ShellToUse = "bash"
//conventional commit types
var features = make(map[string]string)
var fix = make(map[string]string)
var chore = make(map[string]string)
var haveBreakChange = false
var gitRemoteUrl = "https://github.com/hrshadhin/awesome-logger"

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

func replaceMessage(message string, search string, replace string) string  {
    return strings.Replace(message, search, replace, len(search))
}

func formatMessage(message string, sha string, shortSha string) string  {
    messageSlice := []string{ message,
        " ",
        "([",
        shortSha,
        "](",
        gitRemoteUrl,
        "/commit/",
        sha,
        "))",
    }
   return strings.Join(messageSlice, "")
}

func ParseCommits(commits string)  {
    commitsArray := strings.Split(commits, "----DELIMITER----\n")
    for _, commit := range commitsArray {
        commitPart := strings.Split(commit, "\n")
        if len(commitPart) == 2 {
            message := commitPart[0]
            sha := commitPart[1]
            shortSha := sha[:7]
            // remove ! first for below replacement work properly
            if strings.Contains(message, "!:") {
                message = replaceMessage(message, "!", "")
                haveBreakChange = true
            }

            if strings.HasPrefix(message, "chore:") {
                message = replaceMessage(message, "chore: ","")
                chore[sha] = formatMessage(message, sha, shortSha)
            } else if strings.HasPrefix(message, "fix:") {
                message = replaceMessage(message, "fix: ","")
                fix[sha] = formatMessage(message, sha, shortSha)
            } else if strings.HasPrefix(message, "breaking change:") {
                message = replaceMessage(message, "breaking change: ","")
                features[sha] = formatMessage(message, sha, shortSha)
                haveBreakChange = true
            } else {
				if strings.HasPrefix(message, "feature:") {
					message = replaceMessage(message, "feature: ","")
				}
				if strings.HasPrefix(message, "feat:") {
					message = replaceMessage(message, "feat: ","")
				}
                features[sha] = formatMessage(message, sha, shortSha)
            }
        }
    }
}

func writeLine(f *os.File, line string)  {
    l := fmt.Sprintf("%s%s", line, "\n")
    if _, err := f.WriteString(l); err != nil {
        log.Panic(err)
    }
}

func writeReleaseLog()  {
    // open release log file
    f, err := os.OpenFile("release-log.md", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        log.Panic(err)
    }
    // close file on exit and check for its returned error
    defer func() {
        if err := f.Close(); err != nil {
            log.Panic(err)
        }
    }()

    today := time.Now()
    writeLine(f, fmt.Sprintf("# Version 0.1 (%s)", today.Format("2006-01-02")))

    if len(features) > 0 {
    	writeLine(f, "## Feature")
    	for _, message := range features {
    		writeLine(f, "* " + message)
		}
		//write a empty line
		writeLine(f, "")
	}

	if len(fix) > 0 {
		writeLine(f, "## Fix")
		for _, message := range fix {
			writeLine(f, "* " + message)
		}
		//write a empty line
		writeLine(f, "")
	}

	if len(chore) > 0 {
		writeLine(f, "## Chore")
		for _, message := range chore {
			writeLine(f, "* " + message)
		}
		//write a empty line
		writeLine(f, "")
	}
	
    fmt.Println("----------Release Log----------")
    fmt.Println("\tFile: release-log.md")
    fmt.Println("-------------------------------")
}

func main() {
    gitLogCommand := "git log --format=%B%H----DELIMITER----"
    out := Shellout(gitLogCommand)
    ParseCommits(out)
    writeReleaseLog()
}
