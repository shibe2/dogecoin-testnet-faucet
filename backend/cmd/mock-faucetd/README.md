# Mock Faucet API Server

## Usage

Create configuration file.

    mock-faucetd config create mock-faucetd.yaml

Edit *mock-faucetd.yaml* if needed.

By default it will listen on TCP port 80 or 443. To listen on a different port, set **listen** parameter.

    listen: :8080

To bind to a specific interface address, add it to the **listen** parameter as well.

    listen: localhost:8080

To serve static files, set **pubdir** to HTTP root directory.

Set other parameters as desired.

Start mock-faucetd.

    mock-faucetd serve mock-faucetd.yaml

Open control page in a browser to control API replies. By default it will be at *http\://localhost/mock.html*.

Use the API, which is by default available at *http\://localhost/api/*.

To stop mock-faucetd, press Ctrl-C or, on POSIX systems, send SIGINT.
