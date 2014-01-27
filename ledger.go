package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var repo, user, repoDir string

func init() {
	flag.StringVar(&repo, "repo", "lvl1-ycbmropw@stripe-ctf.com:level1", "level 1 of stripe")
	flag.StringVar(&user, "user", "user-bhbu1b3t", "user supplied by stripe")

	flag.Parse()

	repoSplit := strings.Split(repo, ":")
	repoDir = repoSplit[1]
}

func main() {
	runtime.GOMAXPROCS(2)
START:
	startOver()
	clone()
	comm := make(chan string)
	quit := make(chan bool)

	updateLedger()
	gitAddLedger()
	head := getHeadHash()
	tree := getTree()
	difficulty := getDifficulty()

	go mine(comm, quit, head, tree, difficulty)

	timeout := time.After(5 * time.Minute)
	for {
		select {
		case info := <-comm:
			fmt.Println(info)
			close(quit)
			// a break here only leaves the app hanging. debug later.
			os.Exit(0)
		case <-timeout:
			goto START
		default:
		}
	}
}

func getHeadHash() []byte {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = repoDir
	head, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting head hash")
	}
	// remove trailing newline char
	return bytes.TrimSpace(head)
}

func resetHead() {
	cmd := exec.Command("git", "reset", "origin/HEAD", "--hard")
	cmd.Dir = repoDir
	cmd.Run()
}

func gitCommit(msg string) {
	cmd := exec.Command("git", "commit", "-am", msg)
	cmd.Dir = repoDir
	cmd.Run()
}

func gitReset(hash string) {
	cmd := exec.Command("git", "reset", "-hard", hash)
	cmd.Dir = repoDir
	cmd.Run()
}

func gitAddLedger() {
	cmd := exec.Command("git", "add", "LEDGER.txt")
	cmd.Dir = repoDir
	cmd.Run()
}

func getTree() []byte {
	cmd := exec.Command("git", "write-tree")
	cmd.Dir = repoDir
	tree, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting git write-tree", err)
	}
	// remove trailing newlline
	return bytes.TrimSpace(tree)
}

func gitPush() {
	cmd := exec.Command("git", "push", "origin", "master", "-ff")
	cmd.Dir = repoDir
	cmd.Run()
}

func clone() {
	fmt.Println("git clone", repo, "...")
	cmd := exec.Command("git", "clone", repo)
	cmd.Run()
}

// original source: https://github.com/bwilkins/stripe-ctf3-level1-miner-src/blob/master/miner.go
func getSha1(body string) []byte {
	s := sha1.New()
	hash_body := fmt.Sprintf("commit %d%s%s", len(body), []byte{0}, body)
	io.WriteString(s, hash_body)
	checksum := s.Sum(nil)
	checksum_hex := []byte(fmt.Sprintf("%x", checksum))

	return checksum_hex
}

// original source: https://github.com/bwilkins/stripe-ctf3-level1-miner-src/blob/master/miner.go
func GetGitCoin(body, commitHash []byte) (bool, error) {
	// because we are directly calculating the commit hash, we need to add it in to the git db
	err := ioutil.WriteFile(repoDir+"/tmpledger", body, 0644)
	if err != nil {
		fmt.Println("Error writing tmpledger", err)
	}
	hash_cmd := exec.Command("git", "hash-object", "-t", "commit", "-w", "tmpledger")
	hash_cmd.Dir = repoDir
	err = hash_cmd.Run()

	if err == nil {
		updateRef := exec.Command("git", "update-ref", "refs/heads/master", string(commitHash))
		updateRef.Dir = repoDir
		updateRef.Run()
		if err == nil {
			push := exec.Command("git", "push", "origin", "master")
			push.Dir = repoDir
			push.Run()
			return true, nil
		}
	}
	return false, err
}

func getDifficulty() []byte {
	f, err := os.Open(repoDir + "/difficulty.txt")
	if err != nil {
		fmt.Println("Error opening difficulty:", err)
	}
	defer f.Close()
	reader := bufio.NewReader(f)

	line, _, err := reader.ReadLine()
	if err != nil {
		fmt.Println("Error reading difficulty", err)
	}
	return line
}

func updateLedger() {
	fmt.Println("Updating ledger...")
	var newFileContent string
	userFound := false

	f, err := os.OpenFile(repoDir+"/LEDGER.txt", os.O_RDWR, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, user) {
			userFound = true
			data := strings.Split(line, ": ")
			count, err := strconv.Atoi(data[1])
			if err != nil {
				fmt.Println("Error converting to int:", err)
			}
			count += 1
			newCount := strconv.Itoa(count)
			newLine := user + ": " + newCount + "\n"
			newFileContent += newLine
		} else {
			newFileContent += line + "\n"
		}
	}
	err = scanner.Err()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !userFound {
		newFileContent += user + ": 1\n"
	}

	f.Truncate(0)
	f.Seek(0, 0)
	f.Write([]byte(newFileContent))
}

func startOver() {
	fmt.Println("Starting new run")
	cmd := exec.Command("rm", "-rf", repoDir+"*")
	cmd.Run()
}

func mine(comm chan string, quit chan bool, head, tree, difficulty []byte) {
	fmt.Println("Mining...")
	counter := 0
	unix := time.Now().Unix()

	for {
		select {
		case <-quit:
			return
		default:
		}
		counter++

		bodyFmt := `tree %s
parent %s
author CTF %s <me@example.com> %d +0000
committer CTF %s <me@example.com> %d +0000

I can haz GitCoin?

Counter: %d`

		body := fmt.Sprintf(bodyFmt,
			tree, head, user, unix, user, unix, counter)
		sha1Sum := getSha1(body)
		ofInterest := sha1Sum[:len(difficulty)]

		if bytes.Compare(ofInterest, difficulty) < 0 {
			_, err := GetGitCoin([]byte(body), sha1Sum)
			if err != nil {
				fmt.Println("Error getting GitCoin", err)
				comm <- "Mined a gitcoin with " + string(sha1Sum)
				break
			}
		}
		if counter%50000 == 0 {
			// give visual feedback that things are moving
			fmt.Print(".")
		}
	}
}
