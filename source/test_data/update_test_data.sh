#!/bin/bash

# Set the paths to the root directories containing the files
dir_paths=(
    "air_booking"
    "clickstream"
    "guest_profile"
    "hotel_booking"
    "hotel_stay"
    "pax_profile"
    "customer_service_interaction"
)

# Set the destination path for the copied files

# Loop through each root directory
for dir_path in "${dir_paths[@]}"
do
    dest_path=$dir_path
    # Get a list of all the files in the directory and its subdirectories
    echo "Getting a list of all files in directory $dir_path"
    file_list=$(find "examples/$dir_path" -type f)

    # Count the number of files in the list
    file_count=$(echo "$file_list" | wc -l)
    echo "Found $file_count files in directory $dir_path"

    # Generate two random numbers between 1 and the number of files in the list
    random_number1=$((RANDOM % $file_count + 1))
    random_number2=$((RANDOM % $file_count + 1))

    # Make sure the two random numbers are different
    while [ $random_number1 -eq $random_number2 ]
    do
        random_number2=$((RANDOM % $file_count + 1))
    done
    echo "Randomly generated two different numbers: $random_number1 and $random_number2"


    # Get the filenames for the two randomly chosen files
    file1=$(echo "$file_list" | sed -n "${random_number1}p")
    file2=$(echo "$file_list" | sed -n "${random_number2}p")
    echo "Chose two random files to copy: $file1 and $file2"


    echo "Copying files to $dest_path"
    cp "$file1" "$dest_path/data1.jsonl"
    cp "$file2" "$dest_path/data2.jsonl"
    
done