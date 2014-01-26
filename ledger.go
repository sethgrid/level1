package main

import (
	"bufio"
	// "bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
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

func gitCommit() {
	cmd := exec.Command("git", "commit", "-am", "ledger update")
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
		fmt.Println("Line length:", len(line))
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
	fmt.Println(len(newFileContent))
	f.Write([]byte(newFileContent))
}

func main() {

	//for {
	updateLedger()
	gitCommit()

	//resetHead()
	head := getHeadHash()
	difficulty := getDifficulty()
	ofInterest := head[:len(difficulty)]

	fmt.Println("Current Hash\t" + string(head))
	fmt.Println("Relevant Part\t" + string(ofInterest))
	fmt.Println("Difficulty\t" + string(difficulty))

	// 	if bytes.Compare(ofInterest, difficulty) > 0 {
	// 		fmt.Println("Complete")
	// 		break
	// 	}
	// }
}
