#!/bin/bash

# Install expect if not present
if ! command -v expect &> /dev/null; then
    echo "Installing expect..."
    sudo apt-get install expect -y
fi

# Create expect script
cat << 'EOF' > /tmp/telnet_commands.exp
#!/usr/bin/expect -f

# Initialize counters and timing variables
set set_count 0
set set_success 0
set set_total_time 0
set del_count 0
set del_success 0
set del_total_time 0
set get_count 0
set get_success 0
set get_total_time 0

# Store successful SET keys for GET operations
set stored_keys []

# Connect to server
spawn telnet localhost 7000
expect "Connected"

# Generate and send commands
for {set i 0} {$i < 100000} {incr i} {
    set key [format "key_%d_%d" $i [expr int(rand() * 10000)]]
    set value [format "value_%d_%d" $i [expr int(rand() * 1000)]]
    
    # 60% chance of SET, 20% chance of GET, 20% chance of DEL
    set op_chance [expr rand()]
    if {$op_chance < 0.6} {
        set start_time [clock milliseconds]
        send "SET $key $value\r"
        incr set_count
        
        expect {
            "write successfull" {
                incr set_success
                set end_time [clock milliseconds]
                set set_total_time [expr $set_total_time + ($end_time - $start_time)]
                lappend stored_keys $key
            }
            timeout { puts "SET operation timed out" }
        }
    } elseif {$op_chance < 0.8} {
        # GET operation - use a random stored key if available
        if {[llength $stored_keys] > 0} {
            set get_key [lindex $stored_keys [expr {int(rand() * [llength $stored_keys])}]]
            set start_time [clock milliseconds]
            send "GET $get_key\r"
            incr get_count
            
            expect {
                -re {value.*} {
                    incr get_success
                    set end_time [clock milliseconds]
                    set get_total_time [expr $get_total_time + ($end_time - $start_time)]
                }
                timeout { puts "GET operation timed out" }
            }
        }
    } else {
        set start_time [clock milliseconds]
        send "DEL $key\r"
        incr del_count
        
        expect {
            "delete successfull" {
                incr del_success
                set end_time [clock milliseconds]
                set del_total_time [expr $del_total_time + ($end_time - $start_time)]
            }
            timeout { puts "DEL operation timed out" }
        }
    }
}

# Calculate and display statistics
puts "\n=== Operation Statistics ==="
puts "SET Operations:"
puts "  Total: $set_count"
puts "  Successful: $set_success"
if {$set_success > 0} {
    puts [format "  Average time: %.2f ms" [expr double($set_total_time) / $set_success]]
}

puts "\nGET Operations:"
puts "  Total: $get_count"
puts "  Successful: $get_success"
if {$get_success > 0} {
    puts [format "  Average time: %.2f ms" [expr double($get_total_time) / $get_success]]
}

puts "\nDEL Operations:"
puts "  Total: $del_count"
puts "  Successful: $del_success"
if {$del_success > 0} {
    puts [format "  Average time: %.2f ms" [expr double($del_total_time) / $del_success]]
}

send "quit\r"
expect eof
EOF

# Make the expect script executable
chmod +x /tmp/telnet_commands.exp

# Run the expect script
/tmp/telnet_commands.exp

# Clean up
rm /tmp/telnet_commands.exp