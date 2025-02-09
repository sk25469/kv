#!/bin/bash

# Test parameters

cat << 'EOF' > /tmp/large_test.exp
#!/usr/bin/expect -f

# Arrays to track state
array set expected_values {}
set total_sets 0
set total_dels 0
set NUM_OPERATIONS 1000
set KEY_PREFIX "test_key_"
set VALUE_PREFIX "test_value_"


# First connection - Set and Delete values
spawn telnet localhost 7000
expect "Connected"

# Perform random operations
for {set i 0} {$i < 1000} {incr i} {
    set key "${KEY_PREFIX}$i"
    set value "${VALUE_PREFIX}$i"
    
    # 70% chance of SET, 30% chance of DEL
    if {rand() < 0.7} {
        send "SET $key $value\r"
        expect "write successfull"
        set expected_values($key) $value
        incr total_sets
    } else {
        send "DEL $key\r"
        expect "delete successfull"
        unset -nocomplain expected_values($key)
        incr total_dels
    }
}

puts "\nInitial operations complete:"
puts "Total SETs: $total_sets"
puts "Total DELs: $total_dels"
puts "Expected keys remaining: [array size expected_values]"

# Close first connection
send "quit\r"
expect eof

# Wait for server restart
puts "\nWaiting for server restart..."
sleep 10

# Second connection - Verify values
puts "\nStarting verification..."
spawn telnet localhost 7000
expect "Connected"

set correct_values 0
set incorrect_values 0
set missing_values 0

# Check each expected key
foreach {key value} [array get expected_values] {
    send "GET $key\r"
    expect {
        $value {
            incr correct_values
        }
        "nil" {
            incr missing_values
            puts "Missing key: $key"
        }
        default {
            incr incorrect_values
            puts "Incorrect value for key: $key"
        }
    }
}

puts "\nVerification complete:"
puts "Correct values: $correct_values"
puts "Incorrect values: $incorrect_values"
puts "Missing values: $missing_values"
puts "Accuracy: [expr {double($correct_values) / [array size expected_values] * 100}]%"

send "quit\r"
expect eof
EOF

# Make expect script executable
chmod +x /tmp/large_test.exp

echo "Starting large-scale accuracy test..."
/tmp/large_test.exp

# Clean up
rm /tmp/large_test.exp