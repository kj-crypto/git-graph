#!/usr/bin/env bash

DIR_NAME="debug_git_data"

configure_repo() {
	rm -rf .git $DIR_NAME
	mkdir -p $DIR_NAME

	mkdir -p .git .git/objects .git/refs .git/refs/heads
	echo "ref: refs/heads/master" > .git/HEAD

	git config user.name "commit bot"
	git config user.email "commit.bot@noreply.com"
	git config core.editor "vim"
}


commit_random_file() {
	local file_name=$1
	local timestamp=$2
    content=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 100)
    echo "$content" > "$DIR_NAME/$file_name"
    git add -f "$DIR_NAME/$file_name" && git commit -m "$file_name" --date="$timestamp +0000"
}


_sequence_of_commits() {
    local timestamp=$1
    local number_of_commits=$2
	branch_name=$(git rev-parse --abbrev-ref HEAD)
    for i in $(seq 1 "$number_of_commits"); do
        commit_random_file "$branch_name-c$i" "$timestamp"
        timestamp=$((timestamp + 60))
    done
}

sequence_of_commits() {
	local timestamp=$1
	local number_of_commits=$2
	_sequence_of_commits "$timestamp" "$number_of_commits" > /dev/null
	echo $((timestamp + ((number_of_commits + 1))*60))
}

configure_new_repo_with_confirmation() {
	RED="\033[1;31m"
	RESET="\033[0m"

	echo -e "$RED""This will override your current repo!""$RESET"
	read -p "Are you sure? (y/n) " -n 1 -r
	echo ""
	if [[ ! $REPLY =~ ^[Yy]$ ]]; then
		exit 1
	fi
	configure_repo
}


octopus_merge_from_inside() {
	local start_date=$1
	timestamp=$(date -d "$start_date" +%s)
	commit_random_file "master-c1" "$timestamp"
	timestamp=$(($timestamp + 60))

	git branch f1 master
	git checkout f1
	timestamp=$(sequence_of_commits "$timestamp" 3)

	git branch f2 master
	git checkout f2
	timestamp=$(sequence_of_commits "$timestamp" 4)

	git branch f3 master
	git checkout f3
	timestamp=$(sequence_of_commits "$timestamp" 5)

	git checkout master

	git merge --no-ff -m "Merge" -X theirs \
	$(git rev-parse 'f1~2') \
	$(git rev-parse 'f2~3') \
	$(git rev-parse 'f3~4')
	git commit --amend --no-edit --date="$timestamp +0000"
}

canonical_octopus_merge() {
	local start_date=$1
	local branch_number=$2
	timestamp=$(date -d "$start_date" +%s)
	commit_random_file "master-c1" "$timestamp"
	timestamp=$(($timestamp + 60))

	for i in $(seq 1 "$branch_number"); do
		git branch "b$i" master
		git checkout "b$i"
		timestamp=$(sequence_of_commits "$timestamp" 1)
		done

	git checkout master

	git merge --no-ff -m "Merge" -X theirs \
	$(eval echo "b"{1..$branch_number})

	git checkout "b$branch_number"
	timestamp=$(sequence_of_commits "$timestamp" 1)

	git checkout master
	git commit --amend --no-edit --date="$timestamp +0000"
}


_many_merges() {
	local start_date=$1

	timestamp=$(date -d "$start_date" +%s)
	commit_random_file "c1" "$timestamp"
	timestamp=$(($timestamp + 60))

	git branch b1 master
	git branch b2 master
	git branch b3 master
	git branch b5 master

	git checkout b2
	commit_random_file "b2-1" "$timestamp"
	timestamp=$(($timestamp + 60))

	git branch b4 b2

	git checkout b1
	git merge --no-ff -m "M b1<-b2" -X theirs b2
	git commit --amend --no-edit --date="$timestamp +0000"
	timestamp=$(($timestamp + 60))

	git checkout b3
	git merge --no-ff -m "M b2->b3" -X theirs b2
	git commit --amend --no-edit --date="$timestamp +0000"
	timestamp=$(($timestamp + 60))

	git checkout master
	commit_random_file "c2" "$timestamp"
	timestamp=$(($timestamp + 60))

	git checkout b2
	git merge --no-ff -m "b1->b2<-b3" -X theirs b1 b3
	git commit --amend --no-edit --date="$timestamp +0000"
	timestamp=$(($timestamp + 60))

	git checkout b4
	commit_random_file "b4-1" "$timestamp"
	timestamp=$(($timestamp + 60))

	git checkout master
	git merge --no-ff -m "M<-b2<-b4" -X theirs b2 b4
	git commit --amend --no-edit --date="$timestamp +0000"
	timestamp=$(($timestamp + 60))

	git checkout b4
	commit_random_file "b4-2" "$timestamp"
	timestamp=$(($timestamp + 60))

	git checkout b2
	commit_random_file "b2-2" "$timestamp"
	timestamp=$(($timestamp + 60))

	git merge --no-ff -m "m->b2<-b4" -X theirs \
	$(git rev-parse 'master~1') \
	$(git rev-parse 'b4~1')
	git commit --amend --no-edit --date="$timestamp +0000"
	timestamp=$(($timestamp + 60))

	git checkout master
	commit_random_file "c3" "$timestamp"
}

many_merges_minimal() {
	local start_date=$1
	_many_merges "$start_date"

	timestamp=$(git log --format="%at" | sort -r | head -1)
	timestamp=$(($timestamp + 60))

	git checkout b4
	commit_random_file "b4-3" "$timestamp"
}

many_merges_readable() {
	local start_date=$1
	_many_merges "$start_date"

	timestamp=$(git log --format="%at" | sort -r | head -1)
	timestamp=$(($timestamp + 60))
	
	git checkout b1
	timestamp=$(sequence_of_commits "$timestamp" 3)

	git checkout master
	commit_random_file "c4" "$timestamp"
	timestamp=$(($timestamp + 60))

	git checkout b3
	timestamp=$(sequence_of_commits "$timestamp" 2)

	git checkout b5
	timestamp=$(sequence_of_commits "$timestamp" 1)
}



help() {
	cat << EOF
Usage: $0 <command>

Commands:
-mr	many_merges_readable
-mm	many_merges_minimal
-com	create canonical octopus merge with 5 branches
-omi 	create octopus merge from inside
--help	show this help message
EOF
}


start_date="2025-05-27 12:00:00"

case $1 in
	-mr)
		configure_new_repo_with_confirmation
		many_merges_readable "$start_date"
		;;
	-mm)
		configure_new_repo_with_confirmation
		many_merges_minimal "$start_date"
		;;
	-com)
		configure_new_repo_with_confirmation
		canonical_octopus_merge "$start_date" 5
		;;
	-omi)
		configure_new_repo_with_confirmation
		octopus_merge_from_inside "$start_date"
		;;
	--help)
		help
		;;
	esac
