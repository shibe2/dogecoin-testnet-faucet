{ # try
    kill $(ps -aux | grep '[n]pm start' | awk '{print $2}')
} || { # catch
    echo "Kill failed."
}
