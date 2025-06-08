# Git Graph
Git Graph is a tool for visualizing git history in the terminal.
It aims to be ascetic replacement for `git log --graph`

## Build



## Environment variables

- `GRAPH_LOG_LEVEL`: Set to `debug` to enable debug logging. Loggs will be written to `~/.git-graph/log` directory
- `GRAPH_SAVE_JSON`: Set to `true` to save commit positions and commits to `commit_positions.json` file, which can be render by [visualizer.py](./visualizer.py)


## Algorithm
The algorithm details is described in [docs/algorithm.md](./docs/algorithm.md)
