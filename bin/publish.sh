#!/usr/bin/env bash

set -euo pipefail

git_check=true

while getopts "G" opt; do
  case $opt in
    G) git_check=false ;;
    *) echo "Invalid option: -$opt" && exit 1 ;;
  esac
done

# Publishes a new version of the package as a github release.

# Returns the current version of the package.
function current_version() {
    git describe --tags --abbrev=0
}

function next_version() {
  curr_version=$(current_version)
  curr_version=${curr_version#v}
  IFS='.' read -r major minor patch <<< "$curr_version"
  new_major=$major
  new_minor=$minor
  new_patch=$patch

  if release_has_feat; then
    new_minor=$((minor + 1))
  else
    new_patch=$((patch + 1))
  fi

  echo "v${new_major}.${new_minor}.${new_patch}"
}

# Returns true if the current directory is the project root.
function is_project_root() {
  [ -f go.mod ]
}

function release_has_feat() {
  git log -1 --pretty=%B | grep -q "feat"
}

function release_has_fix() {
  git log -1 --pretty=%B | grep -q "fix"
}

function has_unreleased_changes() {
  git log "$(current_version)..HEAD" --pretty=%B | grep -v "Merge pull request"
}



function ensure_project_root() {
  if ! is_project_root; then
    echo "🚫 Not in project root"
    exit 1
  fi
}

function ensure_clean_repo() {
  unstaged=$(git status --porcelain)
  if [ -n "$unstaged" ]; then
    echo "🧹Repository is not clean."
    exit 1
  fi
}

function ensure_main_branch() {
  current_branch=$(git branch --show-current)
  if [ "$current_branch" != "main" ]; then
    echo "🚫 Not on main branch"
    exit 1
  fi
}

# Ensures if we can cleanly push commits to the remote repo, without any
# conflicts, and without actually pushing anything.
function ensure_clean_push() {
  # Fetch latest changes from remote to ensure we have up-to-date refs
  git fetch origin

  # Get the current branch
  current_branch=$(git branch --show-current)

  # Check if the remote branch exists
  if ! git show-ref --verify --quiet "refs/remotes/origin/$current_branch"; then
    echo "🚫 Remote branch origin/$current_branch does not exist"
    exit 1
  fi

  # Check if we're behind the remote (need to pull first)
  local_commit=$(git rev-parse HEAD)
  remote_commit=$(git rev-parse "origin/$current_branch")
  merge_base=$(git merge-base HEAD "origin/$current_branch")

  if [ "$local_commit" = "$remote_commit" ]; then
    # Already up to date, nothing to push
    return 0
  elif [ "$merge_base" = "$local_commit" ]; then
    # We're behind remote, need to pull first
    echo "🚫 Local branch is behind remote. Please pull first."
    exit 1
  elif [ "$merge_base" = "$remote_commit" ]; then
    # We're ahead of remote, can push cleanly
    return 0
  else
    # Branches have diverged, would need merge/rebase
    echo "🚫 Local and remote branches have diverged. Please merge or rebase first."
    exit 1
  fi
}

function is_ahead_of_remote() {
  current_branch=$(git branch --show-current)
  git rev-parse HEAD >/dev/null 2>&1
  git rev-parse "origin/$current_branch" >/dev/null 2>&1
  git merge-base HEAD "origin/$current_branch" >/dev/null 2>&1
}

function ensure_has_unreleased_changes() {
  if ! has_unreleased_changes; then
    echo "🚫 No unreleased changes"
    echo "☝️  Make a new commit"
    exit 1
  fi
}

function confirm() {
  echo "🔄 $1, proceed?"
  read -rp "y/n: " ok
  if [ "$ok" != "y" ]; then
    echo "🚫 Aborting"
    exit 1
  fi
}

function main() {
  ensure_project_root
  ensure_has_unreleased_changes

  if $git_check; then
    ensure_main_branch
    ensure_clean_repo
    ensure_clean_push
  fi

  curr_version=$(current_version)
  next_version=$(next_version)
  if [ "$curr_version" == "$next_version" ]; then
    echo "🚫 No changes to release"
    exit 1
  fi

  if is_ahead_of_remote; then
    confirm "Pushing changes to main branch"
    git push origin main
  fi

  echo "$curr_version → $next_version"
  confirm "Tagging new version"
  git tag -a "$next_version" -m "$next_version"

  confirm "Pushing tag to remote"
  git push origin "$next_version"

  confirm "Creating release"
  gh release create "$next_version" --generate-notes
}

main