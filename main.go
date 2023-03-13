package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Commit struct {
	XMLName            xml.Name `xml:"Commit"`
	Repo               string   `xml:"Repo"`
	Author             string   `xml:"Author"`
	Date               string   `xml:"Date"`
	Message            string   `xml:"Message"`
	LinesChanged       int      `xml:"LinesChanged"`
	LinesAdded         int      `xml:"LinesAdded"`
	LinesDeleted       int      `xml:"LinesDeleted"`
	RelatedClickUpTask string   `xml:"RelatedClickUpTask"`
}

type Commits struct {
	XMLName xml.Name `xml:"Commits"`
	Commits []Commit `xml:"Commit"`
}

var messageTracker = make(map[string]int)

func main() {
	// Define flags for start and end dates
	startDateFlag := flag.String("start", "", "Start date in YYYY-MM-DD format")
	endDateFlag := flag.String("end", "", "End date in YYYY-MM-DD format")

	// Define a flag for repository paths
	reposFlag := flag.String("repos", "", "Comma-separated list of repository paths")

	flag.Parse()

	if *startDateFlag == "" || *endDateFlag == "" || *reposFlag == "" {
		fmt.Println("Usage: go run script.go -start=YYYY-MM-DD -end=YYYY-MM-DD -repos=path/to/repo1,path/to/repo2")
		os.Exit(1)
	}

	startDate, err := time.Parse("2006-01-02", *startDateFlag)
	if err != nil {
		fmt.Println("Error parsing start date:", err)
		os.Exit(1)
	}

	endDate, err := time.Parse("2006-01-02", *endDateFlag)
	if err != nil {
		fmt.Println("Error parsing end date:", err)
		os.Exit(1)
	}

	cuTaskPattern := regexp.MustCompile(`(?i)CU-[^[:alnum:]]*([[:alnum:]]+)[^[:alnum:]]*`)

	// Get the repository paths from the command line argument
	repos := strings.Split(*reposFlag, ",")
	// Retrieve the commits for each repository and store them in a slice of Commit structs
	commits := make([]Commit, 0)
	for _, repo := range repos {
		repoPath, err := filepath.Abs(repo)
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}
		repoName := strings.Split(repoPath, "\\")[len(strings.Split(repoPath, "\\"))-1]

		// Run the Git log command and capture its output
		cmd := exec.Command("git", "-C", repoPath, "log", "--branches", "--remotes", "--tags", "--pretty=format:%an|%ad|%s", "--numstat", "--date=iso")
		out, err := cmd.Output()
		if err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			continue
		}

		// Parse the output into a slice of Commit structs
		for _, chunk := range strings.Split(string(out), "\n\n") {
			lines := strings.Split(strings.TrimSpace(chunk), "\n")
			commitInfo := strings.SplitN(lines[0], "|", 3)
			if len(commitInfo) != 3 {
				continue
			}

			author := commitInfo[0]
			date := commitInfo[1]
			message := commitInfo[2]

			commitDate, err := time.Parse("2006-01-02 15:04:05 -0700", date)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			if commitDate.Before(startDate) || commitDate.After(endDate.AddDate(0, 0, 1)) {
				continue
			}

			linesAdded := 0
			linesDeleted := 0
			for _, line := range lines[1:] {
				parts := strings.Split(line, "\t")
				if len(parts) != 3 {
					continue
				}
				if parts[0] == "-" {
					continue
				}
				if parts[1] == "-" {
					continue
				}
				addLines, err := strconv.Atoi(parts[0])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}

				deleteLines, err := strconv.Atoi(parts[1])
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				linesAdded += addLines
				linesDeleted += deleteLines
			}
			linesChanged := linesAdded + linesDeleted
			if messageTracker[fmt.Sprintf("%s_%s", repoName, message)] == 0 {
				commit := Commit{Repo: repoName, Author: author, Date: date, Message: message, LinesChanged: linesChanged, LinesAdded: linesAdded, LinesDeleted: linesDeleted}
				if match := cuTaskPattern.FindStringSubmatch(message); match != nil {
					taskId := match[1]
					commit.RelatedClickUpTask = fmt.Sprintf("https://app.clickup.com/t/%s", taskId)
				}

				commits = append(commits, commit)
				messageTracker[fmt.Sprintf("%s_%s", repoName, message)] = 1
			}
		}
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Date < commits[j].Date
	})
	commitsXml := Commits{Commits: commits}

	// Export the commits to an XML file
	output, err := xml.MarshalIndent(commitsXml, "", "    ")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	file, err := os.Create("commits.xml")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.Write([]byte(xml.Header))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_, err = file.Write(output)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("Commits exported to commits.xml")
}
