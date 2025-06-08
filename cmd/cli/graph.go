package main

import (
	"fmt"
	commit "git-graph/pkg/commit"
	graph "git-graph/pkg/graph"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		args = []string{"--all"}
	}

	commits, err := commit.ParseCommits(args)
	if err != nil {
		log.Fatal(err)
	}

	result := graph.ProcessCommits(&commits)
	fmt.Println(result)
}
