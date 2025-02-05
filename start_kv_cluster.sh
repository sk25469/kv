#!/bin/bash

# Function to start a KV cluster with a given config file
start_kv_cluster() {
    local config_file=$1
    echo "Starting KV cluster with config file: $config_file"
    sudo ./kv -config="$config_file" &
}

# Array of configuration files
config_files=(
    "/home/sahilsarwar/projects/kv/conf/kv.conf"
    # "/home/sahilsarwar/projects/kv/conf/slave1.conf"
    # "/home/sahilsarwar/projects/kv/conf/slave2.conf"
    # "/home/sahilsarwar/projects/kv/conf/slave3.conf"
    # Add more config files as needed
)

go build

# Loop through the config files and start a KV cluster for each
for config_file in "${config_files[@]}"; do
    start_kv_cluster "$config_file"
done

# Wait for all background processes to finish
wait

echo "All KV clusters started."