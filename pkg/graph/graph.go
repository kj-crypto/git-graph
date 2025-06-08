package graph

import (
	"fmt"
	commit_pkg "git-graph/pkg/commit"
	logger_pkg "git-graph/pkg/logger"
	utils "git-graph/pkg/utils"
	"slices"
	"sort"
)

type Commit = commit_pkg.Commit
type CommitsMap = map[string]*Commit
type ChildrenMap = map[string][]string

var graphMaxX int
var graphMaxY int

var logger = logger_pkg.GetDefaultLogger()

func ComputeCommitsMap(commits *map[string]Commit) CommitsMap {
	commit_map := make(CommitsMap)
	for commit_hash, commit := range *commits {
		commit_map[commit_hash] = &commit
	}
	return commit_map
}

func ComputeChildrenMap(commits *map[string]Commit) ChildrenMap {
	children_map := make(ChildrenMap)
	for commit_hash, commit := range *commits {
		for _, parent := range commit.Parents {
			children_map[parent] = append(children_map[parent], commit_hash)
		}
	}
	return children_map
}

func GetRootCommits(commits_map CommitsMap) []string {
	root_commits := make([]string, 0)
	for _, commit := range commits_map {
		parents := commit.Parents
		if len(parents) == 0 {
			root_commits = append(root_commits, commit.Hash)
			continue
		}
		// This check root commits when listed commits in range not including true root commit
		for _, parent_hash := range parents {
			_, exists := commits_map[parent_hash]
			if !exists {
				root_commits = append(root_commits, commit.Hash)
				break
			}
		}
	}
	return root_commits
}

func GetTopCommits(commits_map CommitsMap, children_map ChildrenMap) []string {
	top_commits := make([]string, 0)
	for commit_hash := range commits_map {
		if _, exists := children_map[commit_hash]; !exists {
			top_commits = append(top_commits, commit_hash)
		}
	}
	return top_commits
}

func ComputeGenerationNumbers(commits_map CommitsMap, top_commits []string) map[string]int {
	generation_numbers := make(map[string]int)

	var compute_generation func(commit *Commit) int
	compute_generation = func(commit *Commit) int {
		_, exists := generation_numbers[commit.Hash]
		if exists {
			return generation_numbers[commit.Hash]
		}
		if len(commit.Parents) == 0 {
			generation_numbers[commit.Hash] = 0
			return 0
		}
		parent_generation_numbers := make([]int, 0)
		for _, parent_hash := range commit.Parents {
			parent_commit, exists := commits_map[parent_hash]
			if !exists {
				generation_numbers[commit.Hash] = 0
				return 0
			}
			parent_generation_numbers = append(parent_generation_numbers, compute_generation(parent_commit))
		}
		generation_numbers[commit.Hash] = slices.Max(parent_generation_numbers) + 1
		return generation_numbers[commit.Hash]
	}

	for _, commit_hash := range top_commits {
		compute_generation(commits_map[commit_hash])
	}
	return generation_numbers
}

func UpdateYPositions(commits_map CommitsMap, generation_numbers map[string]int) {
	max_generation := 0
	for _, generation_number := range generation_numbers {
		max_generation = utils.Max(max_generation, generation_number)
	}
	sorted_commits := make([]*Commit, 0)
	for _, commit := range commits_map {
		commit.GenerationNumber = max_generation - generation_numbers[commit.Hash]
		sorted_commits = append(sorted_commits, commit)
	}
	sort.Slice(sorted_commits, func(i, j int) bool {
		if sorted_commits[i].GenerationNumber == sorted_commits[j].GenerationNumber {
			if len(sorted_commits[i].Parents) == len(sorted_commits[j].Parents) {
				return sorted_commits[i].Timestamp > sorted_commits[j].Timestamp
			}
			return len(sorted_commits[i].Parents) < len(sorted_commits[j].Parents)
		}
		return sorted_commits[i].GenerationNumber < sorted_commits[j].GenerationNumber
	})
	for i, commit := range sorted_commits {
		commit.Y_pos = i
	}
}

func ActiveLanes(commits_map CommitsMap, children_map ChildrenMap) map[string]Commit {
	active_lanes := make(map[int]string)
	active_commits := utils.NewSet[string]()

	sorted_commits := make([]*Commit, 0)
	for _, commit := range commits_map {
		sorted_commits = append(sorted_commits, commit)
	}
	sort.Slice(sorted_commits, func(i, j int) bool {
		return sorted_commits[i].Y_pos < sorted_commits[j].Y_pos
	})

	check_if_branch_commits := func(commit_1, commit_2 string) bool {
		c1 := commits_map[commit_1]
		c2 := commits_map[commit_2]
		if c1.Y_pos > c2.Y_pos {
			c1, c2 = c2, c1
		}
		return c1.Parents[0] == c2.Hash
	}

	get_max_lanes_no := func() int {
		max_lane := 0
		for lane_no, commit_hash := range active_lanes {
			if commit_hash != "" {
				max_lane = utils.Max(max_lane, lane_no)
			}
		}
		return max_lane
	}

	check_diverge_commit := func(commit_hash string) bool {
		children, exists := children_map[commit_hash]
		if exists && len(children) > 1 {
			return true
		}
		return false
	}

	dummy_commits := make(map[string]*Commit)
	active_dummy_commits := make(map[string]*Commit)

	for _, commit := range sorted_commits {
		is_diverge_commit := check_diverge_commit(commit.Hash)

		lane := -1
		if is_diverge_commit {
			lanes := make([]int, 0)
			// Find only direct branch continuation
			for lane_no, commit_hash := range active_lanes {
				if lane_commit, exists := commits_map[commit_hash]; exists && lane_commit.Parents[0] == commit.Hash {
					lanes = append(lanes, lane_no)
				}
			}
			lane = slices.Min(lanes)
			// Close all lanes except the one with the minimum lane number
			for _, lane_no := range lanes {
				if lane_no != lane {
					active_lanes[lane_no] = ""
				}
			}
			active_lanes[lane] = commit.Hash
		} else {
			for lane_no, commit_hash := range active_lanes {
				if commit_hash == commit.Hash {
					lane = lane_no
					break
				}
			}
		}

		if lane == -1 {
			lane = 0
			for {
				if hash, exists := active_lanes[lane]; !exists || hash == "" {
					break
				}
				lane++
			}
		}
		commit.X_pos = lane

		for key, dummy_commit := range active_dummy_commits {
			// Delete dummy commit if direct connection exists
			y_end := commits_map[dummy_commit.Parents[0]].Y_pos
			if commit.Y_pos == y_end {
				var branch_commit *Commit
				for _, child_hash := range children_map[commit.Hash] {
					if check_if_branch_commits(commit.Hash, child_hash) {
						branch_commit = commits_map[child_hash]
					}
				}
				if branch_commit == nil || branch_commit.Y_pos < commits_map[dummy_commit.Message].Y_pos {
					delete(active_dummy_commits, key)
					delete(dummy_commits, key)
				}
			}

			if commit.Y_pos > y_end {
				delete(active_dummy_commits, key)
			}
		}

		// Adjust dummy commits if collides
		is_collided := false
		for _, dummy_commit := range active_dummy_commits {
			if commit.X_pos == dummy_commit.X_pos {
				is_collided = true
				break
			}
		}

		if is_collided {
			for _, dummy_commit := range active_dummy_commits {
				dummy_commit.X_pos++
				graphMaxX = utils.Max(graphMaxX, dummy_commit.X_pos)
			}
		}

		if len(commit.Parents) == 0 {
			continue
		}

		// Add dummy commit for merge commit if needed
		if len(commit.Parents) > 1 {
			new_dummy_commits := make([]*Commit, 0)
			for _, parent_hash := range commit.Parents[1:] {
				need_dummy_commit := true
				for _, commit_hash := range active_lanes {
					if commit_hash != "" && commits_map[commit_hash].Parents[0] == parent_hash &&
						commits_map[commit_hash].Y_pos < commit.Y_pos {
						need_dummy_commit = false
						break
					}
				}
				if need_dummy_commit {
					x_pos := utils.Max(commit.X_pos, get_max_lanes_no()) + 1
					y_pos := commit.Y_pos + 1
					for _, dummy_commit := range active_dummy_commits {
						if dummy_commit.X_pos == x_pos && dummy_commit.Parents[0] != parent_hash {
							x_pos++
						}
					}

					hash := fmt.Sprintf("dummy_%02d", len(dummy_commits))
					dummy_commit := Commit{
						Hash:    hash,
						Message: commit.Hash,
						Parents: []string{parent_hash},
						Y_pos:   y_pos,
						X_pos:   x_pos,
					}
					graphMaxX = utils.Max(graphMaxX, dummy_commit.X_pos)
					dummy_commits[hash] = &dummy_commit
					active_dummy_commits[hash] = &dummy_commit
					new_dummy_commits = append(new_dummy_commits, &dummy_commit)
				}
			}

			if len(new_dummy_commits) > 1 {
				for _, adc_1 := range new_dummy_commits {
					destination_1 := commits_map[adc_1.Parents[0]].Y_pos
					for _, adc_2 := range new_dummy_commits {
						destination_2 := commits_map[adc_2.Parents[0]].Y_pos
						if destination_1 > destination_2 && adc_1.X_pos < adc_2.X_pos {
							x_buf := adc_1.X_pos
							adc_1.X_pos = adc_2.X_pos
							adc_2.X_pos = x_buf
						}
					}
				}
			}
		}

		first_parent := commit.Parents[0]
		if !active_commits.Exists(first_parent) {
			if check_diverge_commit(first_parent) {
				active_commits.Delete(active_lanes[lane])
				active_commits.Add(commit.Hash)
				active_lanes[lane] = commit.Hash
			} else {
				active_commits.Delete(commit.Hash)
				active_commits.Add(first_parent)
				active_lanes[lane] = first_parent
			}
		}

		// Mark a different lane for each parent
		for _, parent_hash := range commit.Parents[1:] {
			if _, exists := commits_map[parent_hash]; exists {
				// Find the next available lane for this parent
				parent_lane := lane
				for {
					if hash, exists := active_lanes[parent_lane]; !exists || hash == "" {
						break
					}
					parent_lane++
				}
				if !active_commits.Exists(parent_hash) && !check_diverge_commit(parent_hash) {
					active_commits.Delete(active_lanes[parent_lane])
					active_commits.Add(parent_hash)
					active_lanes[parent_lane] = parent_hash
				}
			}
		}

		graphMaxX = utils.Max(graphMaxX, get_max_lanes_no())
	}

	returned_dummy_commits := make(map[string]Commit)
	for key, value := range dummy_commits {
		returned_dummy_commits[key] = *value
	}
	return returned_dummy_commits
}

func AddDummyCommits(commits_map map[string]*Commit, dummy_commits *map[string]Commit) {
	for key := range *dummy_commits {
		dummy_commit := (*dummy_commits)[key]
		start_commit := commits_map[dummy_commit.Message]
		end_commit_hash := dummy_commit.Parents[0]
		for parent_idx, parent_hash := range start_commit.Parents {
			if parent_hash == end_commit_hash {
				start_commit.Parents[parent_idx] = dummy_commit.Hash
				commits_map[dummy_commit.Hash] = &dummy_commit
				logger.Debug(fmt.Sprintf("dummy commit added %s -> %s -> %s", start_commit.Hash[:8], dummy_commit.Hash, end_commit_hash[:8]))
				break
			}
		}
	}
}

func ProcessCommits(commits *map[string]Commit) string {
	commits_map := ComputeCommitsMap(commits)
	children_map := ComputeChildrenMap(commits)

	graphMaxY = len(*commits)

	root_commits := GetRootCommits(commits_map)
	logger.Debug(fmt.Sprintf("root commits %v", root_commits))

	top_commits := GetTopCommits(commits_map, children_map)
	logger.Debug(fmt.Sprintf("top commits %v", top_commits))

	generations := ComputeGenerationNumbers(commits_map, top_commits)

	UpdateYPositions(commits_map, generations)
	dummy_commits := ActiveLanes(commits_map, children_map)
	AddDummyCommits(commits_map, &dummy_commits)

	if logger_pkg.IsDebug() {
		logger.Debug(utils.FormatGraphStructure(commits_map, children_map))
	}
	if utils.SaveCommits() {
		utils.SaveCommitPositionsToFile(commits_map, "commit_positions.json")
	}

	return DrawGraph(commits_map, graphMaxX, graphMaxY)
}
