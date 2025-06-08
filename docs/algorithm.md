# Algorithm details
Git graph algorithm contains few steps:

1. Retrieving commits with their parents.`Parents[0]` is natural continuation of the branch.
2. Creating `children_map`, which contains list of children for each commit.
3. Retreiving `root_commits` and `top_commits`.
4. Computing generation number for each commits. It represents the distance from root commit and `Y_pos`
```
FOR commit IN commits
    generation = 0
    FOR parent IN commit.parents
        generation = MAX(generation, parent.generation)
    END FOR
    generation = generation + 1
    commit.generation = generation
END FOR
```

5. Compute `X_pos` using active lines.
```
FOR commit IN sorted commits
    IF diverge commit
        close all open lines except the branch continuation line
    END IF
    lane = find first free lane
    commit.x_pos = lane

    DELETE dummy commit IF direct connection exists
    BUMPUP all dummy commits IF collides with commit
    ADD dummy commits for merge commit IF needed

    FOR parent IN commit.parents
        assign to first available lane
    END FOR

END FOR
```

6. Add dummy commits to commits map
7. Draw graph based on computed positions
