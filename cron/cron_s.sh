#!/bin/bash
# Cron helper script that allows to run a cron job on less than 1 minute intervals.
# Basically depending on input will repeat the job on an interval within 1 minute

if [ $# -ne 2 ]; then
    echo "Usage: $0 <10/15>s <command>"
    exit 1
fi

case $1 in
    *[!0-9]* )  echo "$1 argument should be numeric" && exit 1;;
esac

# $1 should be 10 or 15
if [ $1 -ne 10 ] && [ $1 -ne 15 ]; then
    echo "$1 argument should be 10 or 15"
    exit 1
fi

# Ensure it dies after 60s, with 0 exit code
trap 'exit 0' SIGINT SIGQUIT SIGTERM
{
    sleep 60
    kill $$
} &

# Run $2 command every $1 seconds
if [ $1 -eq 10 ]; then
  ITERATIONS=6
elif [ $1 -eq 15 ]; then
  ITERATIONS=4
fi

for (( i=1; i<=$ITERATIONS; i++ ))
do 
  $2
  # Check non zero exit code and fail
  if [ $? -ne 0 ]; then
    echo "Job failed"
    exit 1
  fi
  # If last iteration, don't sleep
  if [ $i -eq $ITERATIONS ]; then
    break
  fi
  sleep $1
done

exit 0