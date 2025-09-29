#!/usr/bin/env bash
# shellcheck disable=SC2155
#
# Creates images from Go examples using charmbracelet/freeze
# Install freeze: go install github.com/charmbracelet/freeze@latest

set -e

start_capture_after="// START CAPTURE"
end_capture_after="// END CAPTURE"

# Check if freeze is installed
if ! command -v freeze &> /dev/null; then
  echo "Error: freeze is not installed."
  echo "Install with: go install github.com/charmbracelet/freeze@latest"
  exit 1
fi

function create_image() {
  local example="$1"
  local name=$(basename "$example" .go)
  local pretty_image_name="${name}_pretty.png"
  local spew_image_name="${name}_spew.png"
  local font_name="Source Code Pro"

  echo "Creating images for $name..."

  # Create pretty output image
  freeze --execute "go run $example" \
    --output "images/$pretty_image_name" \
    --font.family "$font_name" \
    --border.radius 8 \
    --padding 20

  # Create spew output image
  freeze --execute "go run $example spew" \
    --output "images/$spew_image_name" \
    --font.family "$font_name" \
    --border.radius 8 \
    --padding 20
}

function get_start_line() {
  local file="$1"
  local start_line_number=$(grep -n "$start_capture_after" "$file" | cut -d: -f1)
  echo $((start_line_number + 1))
}

function get_end_line() {
  local file="$1"
  local end_line_number=$(grep -n "$end_capture_after" "$file" | cut -d: -f1)
  echo $((end_line_number - 1))
}

# Create images directory if it doesn't exist
mkdir -p images

# Create images for all examples
echo "Creating images for all examples..."
for example in examples/*.go; do
  if [[ -f "$example" ]]; then
    create_image "$example"
  fi
done

echo "All images created in images/ directory!"

