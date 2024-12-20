#!/bin/bash

# Define the constant cgroup path
cgroup_path="postgresql"  # Replace with your actual cgroup path

process_name='postgres -c config_file=/etc/postgresql/postgresql.conf'

# Find the PID of the process
pid=$(ps aux | grep -v grep | grep "$process_name" | awk '{print $2}')

if [ -z "$pid" ]; then
    echo "No process found with name: $process_name"
    exit 0
    :
fi

# Echo the PID into the cgroup's cgroup.procs file
echo $pid | sudo tee /sys/fs/cgroup/cgroup.procs
