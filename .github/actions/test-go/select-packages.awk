# Map repository-relative inputs to Go packages, then select their test consumers.
# Input 1: changed files. Input 2: tab-separated records emitted by go list.
# cspell:ignore gsub
BEGIN { FS = "\t" }

function normalize(path) {
    gsub(/\\/, "/", path)
    return path
}

FNR == NR { if ($0 != "") changed[$0] = 1; next }
$2 ~ /\.test$/ { next }

$1 == "P" { dependencies[$2] = dependencies[$2] " " $3 }
$1 == "D" {
    directory = normalize($3)
    root = normalize($4) "/"
    if (index(directory, root) != 1) { invalid = 1; next }
    directory = substr(directory, length(root) + 1)
    directories[$2] = directory
    packages[directory] = $2
}
$1 == "T" { fixtures[directories[$2] "/testdata/"] = $2 }
$1 == "F" {
    file = directories[$2] "/" normalize($3)
    embedded[file] = embedded[file] " " $2
}

END {
    if (invalid || length(packages) == 0 || length(changed) == 0) exit 1
    for (file in changed) {
        found = 0
        if (file ~ /\.go$/) {
            directory = file
            sub(/\/[^\/]+$/, "", directory)
            if (directory in packages) {
                seeds[packages[directory]] = 1
                found = 1
            }
        }
        count = split(embedded[file], owners, " ")
        for (i = 1; i <= count; i++) {
            seeds[owners[i]] = 1
            found = 1
        }
        # Repository fixtures belong to the package beside testdata, even when
        # a fixture is deleted or contains .go source for a test project.
        for (prefix in fixtures) if (index(file, prefix) == 1) {
            seeds[fixtures[prefix]] = 1
            found = 1
        }
        # Embedded docs and fixture docs take precedence over ordinary doc skips.
        if (!found && file !~ /\.(md|mdx)$/ && file !~ /\/(dictionary\.txt|_meta\.json)$/) {
            print "No Go package owns changed file: " file > "/dev/stderr"
            exit 1
        }
    }
    for (package in dependencies) {
        if (package in seeds) selected[package] = 1
        count = split(dependencies[package], imports, " ")
        for (i = 1; i <= count; i++) if (imports[i] in seeds) {
            selected[package] = 1
            break
        }
    }
    for (package in selected) print package
}
