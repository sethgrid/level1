package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func getHeadHash() []byte {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	head, err := cmd.Output()
	if err != nil {
		fmt.Println("Error getting head hash")
	}
	// remove trailing newline char
	return head[:len(head)-1]
}

func resetHead() {
	cmd := exec.Command("git", "reset", "origin/HEAD", "--hard")
	cmd.Run()
}

func gitCommit(msg string) {
	cmd := exec.Command("git", "commit", "-am", msg)
	cmd.Run()
}

func getDifficulty() []byte {
	f, err := os.Open("difficulty.txt")
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
	// just appends. Needs to read file, find if entry is there
	// and increment, else append
	user := "user-bhbu1b3t"
	var newFileContent string
	userFound := false

	f, err := os.OpenFile("LEDGER.txt", os.O_RDWR, 0644)
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
}

func main() {

	updateLedger()
	for {
		gitCommit(fmt.Sprintf("I can haz gitcoin? %s", time.Now()))

		//resetHead()
		head := getHeadHash()
		difficulty := getDifficulty()
		ofInterest := head[:len(difficulty)]

		fmt.Println("Current Hash\t" + string(head))
		fmt.Println("Relevant Part\t" + string(ofInterest))
		fmt.Println("Difficulty\t" + string(difficulty))

		if bytes.Compare(ofInterest, difficulty) < 0 {
			fmt.Println("Complete")
			break
		}
	}
}
