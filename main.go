package main

import (
    "bytes"
    "fmt"
    "os/exec"
    "strings"
    "os"
    "time"
    "flag"
)

const ShellToUse = "bash"
const releaseFileName = "release-log.md"
var gitBaseCommand  = "git"
//conventional commit types
var (
	features = make(map[string]string)
	fix = make(map[string]string)
	chore = make(map[string]string)
)
var haveBreakChange = false
var (
	gitRemoteUrl, projectPath, outputPath string
)

func main() {
	parseCliOptions()
	logCommand := gitBaseCommand + " log --format=%B%H----DELIMITER----"
	out := shellout(logCommand)
	parseCommits(out)
	writeReleaseLog()
}

func parseCliOptions() {
	// get cli option
	flag.StringVar(&projectPath, "d", ".", "project directory path")
	flag.StringVar(&outputPath, "o", ".", "output file path")
	flag.Parse()

	// .git directory discovery
	if projectPath != "." {
		if !strings.HasSuffix(projectPath, "/") {
			projectPath = fmt.Sprintf("%s%s",projectPath, "/")
		}

		if !directoryExists(projectPath) {
			fmt.Println("Project path not exists!")
			os.Exit(1)
		}
		gitBaseCommand = fmt.Sprintf("%s %s%s%s", gitBaseCommand, "--git-dir=", projectPath, ".git")
	}

	// output file location
	if outputPath != "." {
		if !strings.HasSuffix(outputPath, "/") {
			outputPath = fmt.Sprintf("%s/", outputPath)
		}

		if !directoryExists(outputPath) {
			fmt.Println("Output path not exists!")
			os.Exit(1)
		}
	}
}

func directoryExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func shellout(command string) (string) {
    var stdout bytes.Buffer
    var stderr bytes.Buffer
    cmd := exec.Command(ShellToUse, "-c", command)
    cmd.Stdout = &stdout
    cmd.Stderr = &stderr
    err := cmd.Run()

    if err != nil {
        fmt.Print(stderr.String())
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

func parseCommits(commits string)  {
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
        fmt.Println(err)
    }
}

func writeReleaseLog()  {
	releaseFilePath := releaseFileName
	if outputPath != "." {
		releaseFilePath = fmt.Sprintf("%s%s", outputPath, releaseFileName)
	}
    // open release log file
    f, err := os.OpenFile(releaseFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
    if err != nil {
        fmt.Println(err)
    }
    // close file on exit and check for its returned error
    defer func() {
        if err := f.Close(); err != nil {
            fmt.Println(err)
            os.Exit(1)
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
