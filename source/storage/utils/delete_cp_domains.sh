#!/bin/bash

# Run the command to list domains and store the output in a variable
output=$(aws customer-profiles list-domains)

# Extract domain names from the output using jq and loop through them
domain_names=$(echo "$output" | jq -r '.Items[].DomainName')
for domain_name in $domain_names; do
    # Run the delete-domain command for each domain name
    #yes Q | aws customer-profiles delete-domain --domain-name "$domain_name"
    expect -c "
    spawn aws customer-profiles delete-domain --domain-name \"$domain_name\"
    expect \"(END)\"
    send \"Q\r\"
    expect eof
    "
done