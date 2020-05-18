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
)

const ShellToUse = "bash"
const AppVersion = "v1.2.2"
var (
	gitBaseCommand  = "git"
	latestTag = ""
	newTag = ""
	tagMessage = ""
	logFileFolder = "release-logs"
	releaseFileName = "release-log.md"
	isCommitLog = false
	gitRemoteUrl, gitRemoteName, projectPath, outputPath string
	haveBreakChange = false
	haveLog = false
	writeNewFile = new(bool)
	logFromBeginning = new(bool)
	showVersionNumber = new(bool)
	//conventional commit types
	features []string
	fixes []string
	chores []string
)

func main() {
	parseCliOptions()
	if *showVersionNumber {
		printVersionInfo()
	}

	findGitRemote()
	findLatestTag()
	collectGitLogs()
	if haveLog {		
		//Get new tag from user
		for {
			// get tag form user
			newTag, tagMessage = getTagFromUserInput()
			if len(newTag) > 0 && len(tagMessage) > 0 {
				break
			}
		}

		writeReleaseLog()
		commitLog()
		makeNewTag()
		pushLatestCommitAndTagToRemote()

	} else {
		fmt.Println("There are no changes made between "+latestTag +" and HEAD")
	}
}

func parseCliOptions() {
	// get cli option
	flag.StringVar(&projectPath, "d", ".", "project directory path")
	flag.StringVar(&outputPath, "o", ".", "output file path")
	writeNewFile = flag.Bool("n", false, "write new release log file")
	logFromBeginning = flag.Bool("b", false, "get logs from the beginning")
	showVersionNumber = flag.Bool("v", false, "show the version number")
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

func printVersionInfo() {
	fmt.Println("-: Awesome Release Logger(ARL) :-")
	fmt.Printf("Version: %s\n", AppVersion)
	fmt.Println("Copyright(c) 2020 TheDevsTech - GPL-3.0")
	os.Exit(0)
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
			gitRemoteUrl, gitRemoteName = getRemoteFromUserInput(remoteList)

		} else {
			for name, url := range remoteList {
				gitRemoteUrl = url
				gitRemoteName = name
				break
			}
		}

		//replace ssh clone url to https
		if len(gitRemoteUrl) > 0 {
			if strings.HasPrefix(gitRemoteUrl, "git@") {
				gitRemoteUrl = replaceMessage(gitRemoteUrl, ":", "/")
				gitRemoteUrl = replaceMessage(gitRemoteUrl, "git@", "https://")
			}
		}
	}
}

func getRemoteFromUserInput(remoteList map[string]string) (string, string) {
	remoteUrl := ""
	remoteName := ""
	for {
		choosenName := getUserChoice(remoteList)
		url, exists := remoteList[choosenName]
		if exists {
			remoteUrl = url
			remoteName = choosenName
			break
		}
	}
	return remoteUrl, remoteName
}

func getUserChoice(remoteList map[string]string) string {
	fmt.Println("Multiple git remote found. Please choose one and write it:")
	for name, _ := range remoteList {
		fmt.Println(name)
	}

	name := ""
	readUserInput("-> ", &name)
	return name
}

func findLatestTag()  {
	latestTagCommand := gitBaseCommand + " rev-list --tags --max-count=1"
	tagHas, err, _ := shellout(latestTagCommand)
	if err == nil && len(tagHas) > 0 {
		latestTagCommand = fmt.Sprintf("%s describe --tags %s" ,gitBaseCommand, tagHas)
		latestTagName, err, _ := shellout(latestTagCommand)
		if err == nil && len(latestTagName) > 0 {
			latestTag = strings.Replace(latestTagName, "\n", "", -1)
		}
	}
}

func collectGitLogs() {
	fmt.Println("Collecting logs...")
	logCommand := gitBaseCommand + " log --format=%B%H----DELIMITER----"
	if len(latestTag) > 0 && *logFromBeginning == false {
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
	// now make the tag
	tagCommand := fmt.Sprintf("%s tag -a -m '%s' %s", gitBaseCommand, tagMessage, newTag)
	_, err, errMsg := shellout(tagCommand)
	if err != nil {
		fmt.Print(errMsg)
		os.Exit(1)
	}

}

func getTagFromUserInput() (string, string)  {
	if len(latestTag) > 0 {
		fmt.Printf("Previous tag is %s\n", latestTag)
		if haveBreakChange {
			fmt.Println("You have breaking changes! So its might be good to update your major version number.")
		}
	}

	nTag := ""
	readUserInput("Enter new tag name: ", &nTag)
	message := ""
	readUserInput("Enter tag message: ", &message)
	return nTag, message
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
				chores = append(chores, formatMessage(message, sha, shortSha))
			} else if strings.HasPrefix(message, "fix:") {
				message = replaceMessage(message, "fix: ","")
				fixes = append(fixes, formatMessage(message, sha, shortSha))
			} else if strings.HasPrefix(message, "breaking change:") {
				message = replaceMessage(message, "breaking change: ","")
				features = append(features, formatMessage(message, sha, shortSha))
				haveBreakChange = true
			} else {
				if strings.HasPrefix(message, "feature:") {
					message = replaceMessage(message, "feature: ","")
				}
				if strings.HasPrefix(message, "feat:") {
					message = replaceMessage(message, "feat: ","")
				}
				features = append(features, formatMessage(message, sha, shortSha))

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
	today := time.Now()
	todayFormated := today.Format("2006-01-02")
	if *writeNewFile {
		releaseFileName = fmt.Sprintf("release-log-%s.md", todayFormated)
	}
	releaseFilePath := releaseFileName
	if outputPath != "." {
		releaseFilePath = fmt.Sprintf("%s%s", outputPath, releaseFileName)
	} else {
		outputPath = fmt.Sprintf("./%s", logFileFolder)
		if !directoryOrFileExists(outputPath){
			os.Mkdir(outputPath, os.ModePerm)
		}
		releaseFilePath = fmt.Sprintf("%s/%s", outputPath, releaseFileName)
		
		//If writing log inside of the repo, then need to commit the log
		isCommitLog = true
	}

	//get previous contents because we need to prepend the latest log
	oldContents := []string{}
	if directoryOrFileExists(releaseFilePath) && ! *writeNewFile {
		f, err := os.OpenFile(releaseFilePath, os.O_RDONLY, 0600)
		if err != nil {
			fmt.Println(err)
		} else {
			// read file and store content in memory
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				tmp := scanner.Text()
				oldContents = append(oldContents, tmp)
			}
		}
		defer f.Close()
	}


	// open release log file
	nf, err := os.OpenFile(releaseFilePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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

	writeLine(nf, fmt.Sprintf("# Version %s (%s)", newTag, todayFormated))

	if len(features) > 0 {
		writeLine(nf, "## Feature")
		for _, message := range features {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}

	if len(fixes) > 0 {
		writeLine(nf, "## Fix")
		for _, message := range fixes {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}

	if len(chores) > 0 {
		writeLine(nf, "## Chore")
		for _, message := range chores {
			writeLine(nf, "* " + message)
		}
		//write a empty line
		writeLine(nf, "")
	}

	//now write diff between two tags
	writeLine(nf, "## Diff")
	diffText := fmt.Sprintf("* %s/compare/%s...%s", gitRemoteUrl, latestTag, newTag)
	writeLine(nf, diffText)


	//now write old logs
    if len(oldContents) > 0 {
		//write empty lines
		writeLine(nf, "")
		writeLine(nf, "")
		for _, line := range oldContents {
			writeLine(nf, line)
		}
	}

	endMessage := "Log File: "+ releaseFilePath
	messageLen := len(endMessage)
	fmt.Println(strings.Repeat("-", messageLen))
	fmt.Println(endMessage)
	fmt.Println(strings.Repeat("-", messageLen))
}

func readUserInput(question string, inputStore *string) {
	fmt.Printf(question)
	reader := bufio.NewReader(os.Stdin)
	inputText, _ := reader.ReadString('\n')
	// convert CRLF to LF
	inputText = strings.Replace(inputText, "\n", "", -1)
	*inputStore = inputText
}

func commitLog() {
	//If writing log inside of the repo, then need to commit the log
	if isCommitLog {
		fmt.Println("Committing new logs")
		addAndCommitCmd := fmt.Sprintf("%s add . && %s commit -m 'added release log for tag: %s'", gitBaseCommand, gitBaseCommand, newTag)
		_, err, errMsg := shellout(addAndCommitCmd)
		if err != nil {
			fmt.Printf("can't commit log!\n error: %s", errMsg)
			os.Exit(1)
		}
	}
}

func pushLatestCommitAndTagToRemote() {
	//if has remote then push it
	if len(gitRemoteUrl) > 0 {
		fmt.Println("Pushing log and tag to remote...")
		pushBaseCmd := fmt.Sprintf("%s push %s", gitBaseCommand, gitRemoteName)
		pushCommitTagCmd := fmt.Sprintf("%s HEAD && %s %s",pushBaseCmd, pushBaseCmd, newTag)
		_, err, errMsg := shellout(pushCommitTagCmd)
		if err != nil {
			fmt.Printf("Push to remove failed!\n error: %s", errMsg)
			os.Exit(1)
		}

		fmt.Printf("Release log and tag: %s has been pushed to remote\n", newTag)
	}
}