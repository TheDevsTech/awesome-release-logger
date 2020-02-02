package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"os"
	"time"
	"flag"
	"bufio"
	"strconv"
)

const ShellToUse = "bash"
const releaseFileName = "release-log.md"
var gitBaseCommand  = "git"
var latestTag = ""
var newTag = ""
var (
	major = int64(1)
	minor = int64(0)
	patch = int64(0)
)
//conventional commit types
var (
	features = make(map[string]string)
	fix = make(map[string]string)
	chore = make(map[string]string)
)
var haveBreakChange = false
var haveLog = false
var (
	gitRemoteUrl, projectPath, outputPath string
)

func main() {
	parseCliOptions()
	findGitRemote()
	findLatestTag()
	collectGitLogs()
	if haveLog {
		makeNewTag()
		writeReleaseLog()
	} else {
		fmt.Println("There are no changes made between "+latestTag +" and HEAD")
	}
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

		if !directoryOrFileExists(projectPath) {
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

		if !directoryOrFileExists(outputPath) {
			fmt.Println("Output path not exists!")
			os.Exit(1)
		}
	}
}

func directoryOrFileExists(path string) bool {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func shellout(command string) (string, error, string) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd := exec.Command(ShellToUse, "-c", command)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()

	return stdout.String(), err, stderr.String()
}

func findGitRemote() {
	remoteCommand := gitBaseCommand + " remote -v"
	remote, err, _ := shellout(remoteCommand)
	if err == nil && len(remote) > 0 {
		remoteArray := strings.Split(remote, "\n");
		remoteList := make(map[string]string)
		for _, line := range(remoteArray) {
			if len(line) > 0 {
				remotePart := strings.Fields(line)
				remoteList[remotePart[0]] = replaceMessage(remotePart[1], ".git", "")
			}
		}

		if len(remoteList) > 1 {
			gitRemoteUrl = getRemoteFromUserInput(remoteList)

		} else {
			for _, url := range remoteList {
				gitRemoteUrl = url
				break
			}
		}
	}
}

func getRemoteFromUserInput(remoteList map[string]string) string {
	remoteUrl := ""
	for {
		choosenName := getUserChoice(remoteList)
		url, exists := remoteList[choosenName]
		if exists {
			remoteUrl = url
			break
		}
	}
	return remoteUrl
}

func getUserChoice(remoteList map[string]string) string {
	fmt.Println("Multiple git remote found. Please choose one and write it:")
	for name, _ := range remoteList {
		fmt.Println(name)
	}
	fmt.Print("-> ")
	reader := bufio.NewReader(os.Stdin)
	name, _ := reader.ReadString('\n')
	// convert CRLF to LF
	name = strings.Replace(name, "\n", "", -1)
	return name
}

func findLatestTag()  {
	latestTagCommand := gitBaseCommand + " describe --long"
	tag, err, _ := shellout(latestTagCommand)
	if err == nil {
		tagPart := strings.Split(tag, "-")
		latestTag = tagPart[0]
	}
}

func collectGitLogs() {
	logCommand := gitBaseCommand + " log --format=%B%H----DELIMITER----"
	if len(latestTag) > 0 {
		cmdSlice := []string{
			gitBaseCommand,
			" log ",
			latestTag,
			"..",
			"HEAD --format=%B%H----DELIMITER----",
		}
		logCommand = strings.Join(cmdSlice, "")
	}
	logs, err, errMsg := shellout(logCommand)
	if err != nil {
		fmt.Println(errMsg)
	}

	if len(logs) > 0 {
		haveLog = true
		parseCommits(logs)
	}

}

func makeNewTag()  {
	//suggest a tag
	suggestTag := findSuggestTag()
	//loop for get user input
	for {
		// get tag form user
		nTag := getTagFromUserInput(suggestTag)
		if len(nTag) == 0 && len(suggestTag) > 0 {
			newTag = suggestTag
			break
		} else {
			nTag, err := validateTag(nTag)
			if len(err) == 0 {
				newTag = nTag
				break
			}
			fmt.Println(err)
		}
	}

	// now make the tag
	tagCommand := fmt.Sprintf("%s tag -a -m 'Version %s' %s", gitBaseCommand, newTag, newTag)
	_, err, _ := shellout(tagCommand)
	if err != nil {
		fmt.Println("Can't create tag. %v", err)
		os.Exit(1)
	}

}

func findSuggestTag() string  {
	if len(latestTag) > 0 {
		latestTagPart := strings.Split(latestTag, ".")
		isValidTag := true
		if m, err := strconv.ParseInt(latestTagPart[0], 10, 64); err == nil {
			major = m
		} else {
			isValidTag = false
		}

		if mi, err := strconv.ParseInt(latestTagPart[1], 10, 64); err == nil {
			minor = mi
		} else {
			isValidTag = false
		}

		if p, err := strconv.ParseInt(latestTagPart[1], 10, 64); err == nil {
			patch = p
		} else {
			isValidTag = false
		}
		if !isValidTag {
			return ""
		}
		if haveBreakChange {
			major = major + 1
			minor = 0
			patch = 0
		}
		if len(features) > 0 {
			minor = minor + 1
		}
		if len(fix) > 0 {
			patch = patch + 1
		}

	}

	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func getTagFromUserInput(sTag string) string  {
	if len(latestTag) > 0 {
		fmt.Println(fmt.Sprintf("Previous tag is %s", sTag))
	}
	message := "Enter new tag name"
	if len(sTag) > 0 {
		message = fmt.Sprintf("%s(%s)", message, sTag)
	}
	fmt.Print(message+":")
	reader := bufio.NewReader(os.Stdin)
	nTag, _ := reader.ReadString('\n')
	// convert CRLF to LF
	nTag = strings.Replace(nTag, "\n", "", -1)

	return nTag
}

func validateTag(tag string) (string, string)  {
	tagPart := strings.Split(tag, ".")
	if len(tagPart) != 3 {
		return "", "Tag should have 3 part major, minor, patch. i.e: 1.1.1"
	}

	if _, err := strconv.ParseInt(tagPart[0], 10, 64); err != nil {
		return "", "Tag major version is not integer!"
	}

	if _, err := strconv.ParseInt(tagPart[1], 10, 64); err != nil {
		return "", "Tag minor version is not integer!"
	}

	if _, err := strconv.ParseInt(tagPart[2], 10, 64); err != nil {
		return "", "Tag patch version is not integer!"
	}

	return tag, ""
}

func replaceMessage(message string, search string, replace string) string  {
	return strings.Replace(message, search, replace, len(search))
}

func formatMessage(message string, sha string, shortSha string) string  {
	messageSlice := []string{}
	if len(gitRemoteUrl) > 0 {
		messageSlice = []string{message,
			" ",
			"([",
			shortSha,
			"](",
			gitRemoteUrl,
			"/commit/",
			sha,
			"))",
		}
	} else {
		messageSlice = []string{message,
			" ",
			"(",
			shortSha,
			")",
		}
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

	//get previous contents because we need to prepend the latest log
	oldContents := []string{}
	if directoryOrFileExists(releaseFilePath) {
		f, err := os.OpenFile(releaseFilePath, os.O_RDONLY, 0600)
		if err != nil {
			fmt.Println(err)
		} else {
			// read file and store content in memory
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				if tmp := scanner.Text(); len(tmp) != 0 {
					oldContents = append(oldContents, tmp)
				}
			}
		}
		defer f.Close()
	}


	// open release log file
	nf, err := os.OpenFile(releaseFilePath, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// close file on exit and check for its returned error
	defer func() {
		if err := nf.Close(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}()

	today := time.Now()
	writeLine(nf, fmt.Sprintf("# Version %s (%s)", newTag, today.Format("2006-01-02")))

	if len(features) > 0 {
		writeLine(nf, "## Feature")
		for _, message := range features {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}

	if len(fix) > 0 {
		writeLine(nf, "## Fix")
		for _, message := range fix {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}

	if len(chore) > 0 {
		writeLine(nf, "## Chore")
		for _, message := range chore {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}
    if len(oldContents) > 0 {
		//write empty lines
		writeLine(nf, "")
		writeLine(nf, "")
		for _, line := range oldContents {
			writeLine(nf, line)
		}
	}


	fmt.Println("----------Release Log----------")
	fmt.Println("\tFile: release-log.md")
	fmt.Println("-------------------------------")
}
