{ # try
    kill $(ps -aux | grep '[h]ttp' | awk '{print $2}')
} || { # catch
    echo "Kill failed."
}
