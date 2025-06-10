package main

import (
	"flag"
	"fmt"
	commit "git-graph/pkg/commit"
	graph "git-graph/pkg/graph"
	"git-graph/pkg/ui"
	"log"
	"regexp"
	"strings"
)

func prepareLines(input string) []map[string]string {
	var result []map[string]string

	pattern := regexp.MustCompile(`(.+) ([a-f0-9]{8}) (.+)`)
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		matches := pattern.FindStringSubmatch(line)
		if len(matches) == 4 {
			result = append(result, map[string]string{
				"hash":  matches[2],
				"body":  matches[3],
				"graph": matches[1],
			})
		} else if len(matches) == 0 {
			result = append(result, map[string]string{
				"graph": line,
			})
		}
	}
	return result
}

func argParse() []string {
	flag.Usage = func() {
		fmt.Println(`Usage: git-graph [options]

Options:
	--all		Show all commits. Default option.
	--help		Show this help message

Examples:
	git-graph
	git-graph 0ef00000..HEAD
	git-graph --help`)
		flag.PrintDefaults()
	}
	flag.Parse()
	if flag.NArg() == 0 {
		return []string{"--all"}
	}
	return flag.Args()
}

func main() {
	args := argParse()

	commits, err := commit.ParseCommits(args)
	if err != nil {
		log.Fatal(err)
	}

	graph_str := graph.ProcessCommits(&commits)
	ui.Run(prepareLines(graph_str), graph.Y_SPACING)
}
