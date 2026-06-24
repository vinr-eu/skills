#!/usr/bin/env bash
set -euo pipefail

REPO="vinr-eu/skills"
REPO_URL="https://github.com/$REPO"
SKILLS_DIR="$HOME/.claude/skills"
CLONE_DIR="$HOME/.vinr/skills"

mkdir -p "$SKILLS_DIR"

link_skills() {
  local src="$1"
  for skill_dir in "$src"/*/; do
    [ -f "$skill_dir/SKILL.md" ] || continue
    local name
    name="$(basename "$skill_dir")"
    local target="$SKILLS_DIR/$name"

    if [ -L "$target" ]; then
      echo "  already installed: $name"
    elif [ -e "$target" ]; then
      echo "  skipped: $name (non-symlink already exists at $target)"
    else
      ln -s "$skill_dir" "$target"
      echo "  installed: $name"
    fi
  done
}

if command -v git &>/dev/null; then
  echo "Using git..."
  if [ -d "$CLONE_DIR/.git" ]; then
    echo "Updating existing clone at $CLONE_DIR"
    git -C "$CLONE_DIR" pull --ff-only
  else
    echo "Cloning $REPO_URL to $CLONE_DIR"
    git clone --depth 1 "$REPO_URL" "$CLONE_DIR"
  fi
  link_skills "$CLONE_DIR"
else
  echo "git not found, downloading via curl..."
  command -v curl &>/dev/null || { echo "Error: curl is required"; exit 1; }
  command -v tar  &>/dev/null || { echo "Error: tar is required";  exit 1; }

  TMP="$(mktemp -d)"
  trap 'rm -rf "$TMP"' EXIT

  echo "Downloading archive from $REPO_URL"
  curl -fsSL "$REPO_URL/archive/refs/heads/main.tar.gz" | tar -xz -C "$TMP"

  local_dir="$TMP/skills-main"
  link_skills "$local_dir"
  echo ""
  echo "Note: skills were copied, not linked. Re-run install.sh to update."
fi

echo ""
echo "Done. Run /reload-skills in Claude Code to activate."
