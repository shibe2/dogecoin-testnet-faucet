{ # try
    kill $(ps -aux | grep '[n]ode server.js' | awk '{print $2}')
} || { # catch
    echo "Kill failed."
}
