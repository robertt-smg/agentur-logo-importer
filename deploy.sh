#!/bin/bash

# Exit on error
set -e

# Configuration
REMOTE_USER="administrator"
REMOTE_HOST="f01.smg-air-conso.com"
REMOTE_BASE_PATH="D:/proquest/airquest/SMG"

# Get project (repo) name and current git tag
PROJECT_NAME=$(basename "$(git rev-parse --show-toplevel)")
GIT_TAG=$(git describe --tags --abbrev=0)

# Local dist folder
DIST_DIR="dist"

# Remote target directory
REMOTE_TARGET_PATH="$REMOTE_BASE_PATH/$PROJECT_NAME/$GIT_TAG"

# Print info
echo "Deploying $DIST_DIR to $REMOTE_USER@$REMOTE_HOST:$REMOTE_TARGET_PATH"

# Create remote directory (using Windows path, so use PowerShell on remote)
ssh $REMOTE_USER@$REMOTE_HOST "powershell -Command \"New-Item -Path '$REMOTE_TARGET_PATH' -ItemType Directory -Force\""

# Copy dist folder contents to remote versioned directory
scp -r "$DIST_DIR"/* $REMOTE_USER@$REMOTE_HOST:"$REMOTE_TARGET_PATH"/

# Create <project name>_vers.sh file with set Version=<Version tag>
VERS_FILE="$PROJECT_NAME"_vers.sh
VERS_CONTENT="export Version=$GIT_TAG\n"
echo -e "$VERS_CONTENT" > "$VERS_FILE"
scp "$VERS_FILE" $REMOTE_USER@$REMOTE_HOST:"$REMOTE_BASE_PATH/$VERS_FILE"
rm "$VERS_FILE"

echo "Deployment complete!"
