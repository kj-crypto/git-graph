package commit

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	logger_pkg "git-graph/pkg/logger"
)

type Commit struct {
	Hash             string
	Message          string
	Timestamp        uint64
	Parents          []string
	HeadOfBranches   []string
	X_pos            int
	Y_pos            int
	GenerationNumber int
}

func (c Commit) Format(message_max_length int) string {
	hash_str := c.Hash[:8]
	message_str := c.Message
	time_str := time.Unix(int64(c.Timestamp), 0).Format("2006-01-02 15:04:05")
	if len(message_str) > message_max_length {
		message_str = message_str[:message_max_length-3] + "..."
	}
	branches_str := "( " + strings.Join(c.HeadOfBranches, ", ") + " )"
	if len(c.HeadOfBranches) > 0 {
		return fmt.Sprintf("%s %s %s %s", hash_str, message_str, time_str, branches_str)
	}
	return fmt.Sprintf("%s %s %s", hash_str, message_str, time_str)
}

var split_separator string = "␞"
var format_string string = fmt.Sprintf("--format=%%H%s%%s%s%%P%s%%at%s%%D", split_separator, split_separator, split_separator, split_separator)
var logger = logger_pkg.GetDefaultLogger()

func ParseCommits(args []string) (map[string]Commit, error) {
	// TODO: handle lack of repo
	cmd := exec.Command("git", "log", format_string)
	cmd.Args = append(cmd.Args, args...)

	output, err := cmd.Output()
	logger.Debug(string(output))

	if err != nil {
		return nil, err
	}
	commits := make(map[string]Commit)

	for index, line := range strings.Split(string(output), "\n") {
		items := strings.Split(line, split_separator)
		if len(items) < 5 {
			continue
		}
		parents := []string{}
		if items[2] != "" {
			parents = strings.Split(items[2], " ")
		}

		timestamp, err := strconv.ParseUint(items[3], 10, 64)
		if err != nil {
			return nil, err
		}

		c := Commit{
			Hash:      items[0],
			Message:   items[1],
			Timestamp: timestamp,
			Parents:   parents,
			X_pos:     0,
			Y_pos:     index,
		}

		if items[4] != "" {
			b := strings.Split(items[4], ",")
			braches := make([]string, 0)
			for _, branch := range b {
				branch = strings.TrimSpace(branch)
				braches = append(braches, branch)
			}
			c.HeadOfBranches = braches
		}
		commits[c.Hash] = c
	}

	return commits, nil
}

func GetCommitStats(commit_hash string) string {
	cmd := exec.Command("git", "show", "--stat", "--color=always", commit_hash)
	cmd.Env = append(os.Environ(), "TERM=xterm-256color")

	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return string(output)
}
