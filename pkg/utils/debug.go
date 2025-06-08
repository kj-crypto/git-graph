package utils

import (
	"encoding/json"
	"fmt"
	commit "git-graph/pkg/commit"
	"log"
	"os"
	"sort"
	"strings"
)

type CommitToStore struct {
	Hash    string   `json:"hash"`
	X_pos   int      `json:"x_pos"`
	Y_pos   int      `json:"y_pos"`
	Parents []string `json:"parents"`
	Message string   `json:"message"`
}

func SaveCommitPositionsToFile(commits map[string]*commit.Commit, file_path string) error {
	positions := make([]CommitToStore, 0)
	for _, commit := range commits {
		positions = append(positions, CommitToStore{
			Hash:    commit.Hash,
			X_pos:   commit.X_pos,
			Y_pos:   commit.Y_pos,
			Parents: commit.Parents,
			Message: commit.Message,
		})
	}

	jsonBytes, err := json.MarshalIndent(positions, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal to JSON: %v", err)
	}

	err = os.WriteFile(file_path, jsonBytes, 0644)
	log.Println("Saved commit positions to file:", file_path)
	return err
}

func SaveCommits() bool {
	return os.Getenv("GRAPH_SAVE_JSON") == "true"
}

func FormatGraphStructure(commits_map map[string]*commit.Commit, children_map map[string][]string) string {
	sorted_commits := make([]*commit.Commit, len(commits_map))
	i := 0
	for commit_hash := range commits_map {
		sorted_commits[i] = commits_map[commit_hash]
		i++
	}

	sort.Slice(sorted_commits, func(i, j int) bool {
		return sorted_commits[i].Y_pos < sorted_commits[j].Y_pos
	})
	result := ""
	for _, commit := range sorted_commits {
		hash := commit.Hash[:8]
		parents := "p=[ "
		for _, parent_hash := range commit.Parents {
			parents += parent_hash[:8] + ", "
		}
		parents += "]"

		children := "c=[ "
		for _, child_hash := range children_map[commit.Hash] {
			children += child_hash[:8] + ", "
		}
		children += "]"
		c_type := ""
		if strings.HasPrefix(commit.Message, "Merge pull request") {
			c_type = "(PR)"
		}
		result += fmt.Sprintf("\033[38;2;255;255;255m %8s\033[0m %4s {Y=%2d X=%2d G=%2d} -- %-26s %s\n", hash, c_type, commit.Y_pos, commit.X_pos, commit.GenerationNumber, parents, children)
	}
	return result
}
