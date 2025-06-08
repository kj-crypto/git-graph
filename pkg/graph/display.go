package graph

import (
	"fmt"
	"strings"
)

const (
	VERTICAL          = "│"
	HORIZONTAL        = "─"
	DOWN_RIGHT_CORNER = "╯"
	UP_RIGTH_CORNER   = "╮"
	UP_LEFT_CORNER    = "╭"
	COMMIT            = "●"
	MERGE_COMMIT      = "○"
	T_DOWN_CONNECTOR  = "┬"
	T_UP_CONNECTOR    = "┴"
	T_LEFT_CONNECTOR  = "├"
	T_RIGHT_CONNECTOR = "┤"
	CROSS_CONNECTOR   = "┼"
)

var COLLORS_PALLETE = []string{
	"\033[38;2;255;182;193m",
	"\033[38;2;173;216;230m",
	"\033[38;2;255;223;170m",
	"\033[38;2;199;214;189m",
	"\033[38;2;188;143;143m",
	"\033[38;2;221;160;221m",
}

const RESET_COLOR = "\033[0m"

const X_SPACING = 4
const Y_SPACING = 2

type gridCell struct {
	glyph        string
	destinationX int
}

func (g *gridCell) getColor() string {
	return COLLORS_PALLETE[g.destinationX%len(COLLORS_PALLETE)]
}

func DrawGraph(commits_map map[string]*Commit, maxX, maxY int) string {
	commits := make(map[int]string)

	// Create grid with spaces
	grid := make([][]gridCell, maxY*Y_SPACING+1)
	for y := range grid {
		grid[y] = make([]gridCell, maxX*X_SPACING+1)
		for x := range grid[y] {
			grid[y][x] = gridCell{" ", x}
		}
	}

	if_dummy_commits := func(commit *Commit) bool {
		return strings.HasPrefix(commit.Hash, "dummy_")
	}

	for _, commit := range commits_map {
		commit_glyph := COMMIT
		if if_dummy_commits(commit) {
			commit_glyph = VERTICAL
		}

		is_merge_commit := false

		if len(commit.Parents) > 1 {
			commit_glyph = MERGE_COMMIT
			is_merge_commit = true
		}

		grid[commit.Y_pos*Y_SPACING][commit.X_pos*X_SPACING] = gridCell{commit_glyph, commit.X_pos}
		if !if_dummy_commits(commit) {
			commits[commit.Y_pos*Y_SPACING] = commit.Format(20)
		}

		for parent_no, parent_hash := range commit.Parents {
			parent, exists := commits_map[parent_hash]
			if !exists {
				continue
			}
			logger.Debug(fmt.Sprintf("%s -> %s", commit.Hash[:8], parent.Hash[:8]))

			x_distance := parent.X_pos - commit.X_pos
			y_distance := parent.Y_pos - commit.Y_pos

			if y_distance < 0 {
				logger.Fatal(fmt.Sprintf("y_distance < 0 %d from %s to %s", y_distance, commit.Hash[:8], parent.Hash[:8]))
			}

			y_start := commit.Y_pos * Y_SPACING
			y_end := parent.Y_pos * Y_SPACING
			x_end := 0
			destinationX := parent.X_pos

			/* resolve merge from right

			   M   ┃  ┃
			   ┣━━━┳━━┓
			   ┃   ┃  ┃
			   ┃   ●  ┃
			   ┃   ┃  ●
			*/
			if x_distance > 0 {
				x_start := commit.X_pos * X_SPACING
				x_end = x_start + x_distance*X_SPACING
				destinationX = parent.X_pos

				if grid[y_start+1][x_start].glyph == T_RIGHT_CONNECTOR {
					grid[y_start+1][x_start] = gridCell{CROSS_CONNECTOR, commit.X_pos}
				} else {
					grid[y_start+1][x_start] = gridCell{T_LEFT_CONNECTOR, commit.X_pos}
				}

				for i := x_start + 1; i < x_end; i++ {
					cell := grid[y_start+1][i]

					if cell.glyph == VERTICAL || cell.glyph == " " {
						grid[y_start+1][i] = gridCell{HORIZONTAL, destinationX}
					} else if cell.glyph == UP_RIGTH_CORNER {
						grid[y_start+1][i] = gridCell{T_DOWN_CONNECTOR, cell.destinationX}
					}

					if destinationX < cell.destinationX {
						grid[y_start+1][i].destinationX = destinationX
					}

				}

				if grid[y_start+1][x_end].glyph == " " || grid[y_start+1][x_end].glyph == VERTICAL {
					grid[y_start+1][x_end] = gridCell{UP_RIGTH_CORNER, destinationX}
				} else if grid[y_start+1][x_end].glyph == HORIZONTAL {
					grid[y_start+1][x_end] = gridCell{T_DOWN_CONNECTOR, destinationX}
				}

				/* resolve branching to right

				         ●  ┃
				      ●  ┃  ┃
				   ┣━━┻━━┻━━┛
				   ●
				   ┃
				*/
			} else if x_distance < 0 && (!is_merge_commit || (is_merge_commit && parent_no == 0)) {
				x_start := parent.X_pos * X_SPACING
				x_end = x_start + (-1)*x_distance*X_SPACING

				destinationX = commit.X_pos

				if grid[y_end-1][x_start].glyph == T_RIGHT_CONNECTOR {
					grid[y_end-1][x_start] = gridCell{CROSS_CONNECTOR, parent.X_pos}
				} else {
					grid[y_end-1][x_start] = gridCell{T_LEFT_CONNECTOR, parent.X_pos}
				}

				for i := x_start + 1; i < x_end; i++ {
					cell := grid[y_end-1][i]

					if cell.glyph == VERTICAL || cell.glyph == " " {
						grid[y_end-1][i] = gridCell{HORIZONTAL, destinationX}
					} else if cell.glyph == DOWN_RIGHT_CORNER {
						grid[y_end-1][i] = gridCell{T_UP_CONNECTOR, cell.destinationX}
					}

					if destinationX < cell.destinationX {
						grid[y_end-1][i].destinationX = destinationX
					}

				}

				if grid[y_end-1][x_end].glyph == " " || grid[y_end-1][x_end].glyph == VERTICAL {
					grid[y_end-1][x_end] = gridCell{DOWN_RIGHT_CORNER, destinationX}
				} else if grid[y_end-1][x_end].glyph == HORIZONTAL {
					grid[y_end-1][x_end] = gridCell{T_UP_CONNECTOR, destinationX}
				}
				/*
					resolve merge to left

					┃   M
					┣━━━┫
					┃   ┃
					●   ┃
					┃   ┃

				*/

			} else if x_distance < 0 && is_merge_commit && parent_no != 0 {
				x_start := parent.X_pos * X_SPACING
				x_end = x_start + (-1)*x_distance*X_SPACING

				if grid[y_start+1][x_start].glyph == T_RIGHT_CONNECTOR {
					grid[y_start+1][x_start] = gridCell{CROSS_CONNECTOR, parent.X_pos}
				} else {
					grid[y_start+1][x_start] = gridCell{T_LEFT_CONNECTOR, parent.X_pos}
				}

				for i := x_start + 1; i < x_end; i++ {
					grid[y_start+1][i] = gridCell{HORIZONTAL, parent.X_pos}
				}

				if grid[y_start+1][x_end].glyph == T_LEFT_CONNECTOR {
					grid[y_start+1][x_end] = gridCell{CROSS_CONNECTOR, commit.X_pos}
				} else {
					grid[y_start+1][x_end] = gridCell{T_RIGHT_CONNECTOR, commit.X_pos}
				}

				continue

			} else {
				x_end = commit.X_pos * X_SPACING
			}
			// go down
			for i := y_start + 1; i < y_end; i++ {
				if grid[i][x_end].glyph == " " {
					grid[i][x_end] = gridCell{VERTICAL, destinationX}
				}
			}
		}
	}
	return gridToString(grid, commits)
}

func gridToString(grid [][]gridCell, commits map[int]string) string {
	var result strings.Builder
	for i, row := range grid {
		for _, cell := range row {
			result.WriteString(cell.getColor() + cell.glyph + RESET_COLOR)
		}
		if _, exists := commits[i]; exists {
			result.WriteString(strings.Repeat(" ", 2*X_SPACING) + commits[i])
		}
		result.WriteString("\n")
	}

	return result.String()
}
