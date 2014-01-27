package main

import (
	"bufio"
	"bytes"
	"crypto/sha1"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func getHeadHash(i int) []byte {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "rev-parse", "HEAD")
	cmd.Dir = "level1_" + iStr
	head, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting head hash")
	}
	// remove trailing newline char
	return head[:len(head)-1]
}

func resetHead(i int) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "reset", "origin/HEAD", "--hard")
	cmd.Dir = "level1_" + iStr
	cmd.Run()
}

func gitCommit(i int, msg string) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "commit", "-am", msg)
	cmd.Dir = "level1_" + iStr
	cmd.Run()
}

func gitReset(i int, hash string) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "reset", "-hard", hash)
	cmd.Dir = "level1_" + iStr
	cmd.Run()
}

func gitAddLedger(i int) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "add", "LEDGER.txt")
	cmd.Dir = "level1_" + iStr
	cmd.Run()
}

func getTree(i int) []byte {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "write-tree")
	cmd.Dir = "level1_" + iStr
	tree, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting git write-tree", err)
	}
	return tree
}

func getSha1(body string) []byte {
	s := sha1.New()
	hash_body := fmt.Sprintf("commit %d%s%s", len(body), []byte{0}, body)
	io.WriteString(s, hash_body)
	checksum := s.Sum(nil)
	checksum_hex := []byte(fmt.Sprintf("%x", checksum))

	return checksum_hex
}

func gitPush(i int) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("git", "push", "origin", "master", "-ff")
	cmd.Dir = "level1_" + iStr
	cmd.Run()
}

func clone() {
	cmd := exec.Command("git", "clone", repo)
	cmd.Run()
}

func copyRepo(i int) {
	iStr := strconv.Itoa(i)
	cmd := exec.Command("cp", "-r", "level1", "level1_"+iStr)
	_, err := cmd.Output()
	if err != nil {
		fmt.Println("There was an error copying the repo", err)
	}
	fmt.Println("Copied repo to level1_" + iStr)
}

func getDifficulty(i int) []byte {
	iStr := strconv.Itoa(i)
	f, err := os.Open("level1_" + iStr + "/difficulty.txt")
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

func updateLedger(i int) {
	var newFileContent string
	userFound := false
	iStr := strconv.Itoa(i)

	f, err := os.OpenFile("level1_"+iStr+"/LEDGER.txt", os.O_RDWR, 0644)
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

var repo, user string

func init() {
	flag.StringVar(&repo, "repo", "lvl1-ycbmropw@stripe-ctf.com:level1", "level 1 of stripe")
	flag.StringVar(&user, "user", "user-bhbu1b3t", "user supplied by stripe")

	flag.Parse()
}

func mine(comm chan string, quit chan bool, i int, head, difficulty []byte) {
	counter := 0
	unix := time.Now().Unix()

	tree := getTree(i)

	for {
		select {
		case <-quit:
			return
		default:
		}
		counter++

		body := fmt.Sprintf("tree %s\nparent %s\nauthor CTF %s %d\n\nI Can haz gitcoin?\n\nCounter:%d",
			tree, head, user, unix, counter)
		sha1_sum := getSha1(body)
		ofInterest := sha1_sum[:len(difficulty)]

		if bytes.Compare(ofInterest, difficulty) < 0 {
			fmt.Println("Mined a gitcoin with ", string(sha1_sum))
			//gitReset(i, string(sha1_sum))
			iStr := strconv.Itoa(i)
			comm <- "Found it! in level1_" + iStr + " " + string(sha1_sum)
			gitPush(i)
			break
		}
		if counter%50000 == 0 {
			// fmt.Println(string(body))
			// fmt.Println(string(sha1_sum))
			// fmt.Println(string(ofInterest))
			// fmt.Println(string(difficulty))
			fmt.Print(".")
		}
	}

}

func main() {
	runtime.GOMAXPROCS(7)
	clone()
	comm := make(chan string)
	quit := make(chan bool)
	for i := 0; i < 250; i++ {
		copyRepo(i)
		updateLedger(i)
		gitAddLedger(i)
		//resetHead()
		head := getHeadHash(i)
		difficulty := getDifficulty(i)

		go mine(comm, quit, i, head, difficulty)
	}
	for {
		select {
		case info := <-comm:
			fmt.Println(info)
			close(quit)
			// a break here only leaves the app hanging. debug later.
			os.Exit(0)
		default:
		}
	}
}
