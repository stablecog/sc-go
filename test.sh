LATEST_RELEASE=$(curl -s https://api.github.com/repos/actions/runner/releases/latest | grep -oP '"tag_name": "\K(.*)(?=")')
echo $LATEST_RELEASE