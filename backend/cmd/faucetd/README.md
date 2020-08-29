# Faucet Back-end Service

## Getting Started

Create configuration file.

    faucetd config create faucetd.yaml

Edit *faucetd.yaml*.

Set **amount** to desired giveaway amount.

By default it will listen on TCP port 80 or 443. To listen on a different port, set **listen** parameter.

    listen: :8080

To bind to a specific interface address, add it to the **listen** parameter as well.

    listen: localhost:8080

To use HTTPS, provide **certfile** and **keyfile**.

To serve static files, set **pubdir** to HTTP root directory.

Set database driver and location.

    db:
        driver: sqlite3
        source: faucet.sqlite

If you run Dogecoin Core as the same user on the same host, leave RPC parameters on default. Otherwise set **url**, **username**, and **password** for access to the wallet. If you don't use RPC cookie file, remove **cookiefile** parameter or set it to empty string.

Set other parameters as desired.

Create database file.

    faucetd db create faucetd.yaml

Start faucetd.

    faucetd serve faucetd.yaml

To stop faucetd, press Ctrl-C.

## Subcommands

**faucetd config create** *configout.yaml*

Creates configuration file *configout.yaml* with default values. If the file already exists, it will be overwritten, and permissions will be kept. Otherwise a new file with default permissions will be created.

**faucetd config dump** *config.yaml*

Reads *config.yaml* and outputs effective configuration to stdout.

**faucetd config process** *config.yaml* *configout.yaml*

Reads *config.yaml* and writes *configout.yaml*. It can be used to format configuration file and to add missing parameters with default values. Input and output file can be the same. If a new file will be created, it will have default permissions.

**faucetd db create** *config.yaml*

Creates needed tables in a database specified in *config.yaml*.

**faucetd db sql** *driver_name*

Outputs SQL statements that create needed tables. *driver_name* selects SQL dialect; supported driver: sqlite3.

**faucetd serve** *config.yaml*

Starts faucet back-end service using configuration from *config.yaml*. To stop it, press Ctrl-C or, on POSIX systems, send SIGINT.

## Configuration

Relative path names in configuration file are relative to current directory in which faucetd will be started.

Durations/intervals are specified in hours, minutes and seconds. For example: "1h2m3s" or "62m3s" or "3723s".

You may want to restrict access to configuration file if it contains secrets.

**amount**

Amount of coins to send per claim. If it's less than **minamount**, the faucet will be paused. Default: 0.

**fee**

Estimated transaction fee. It is used to compute last claim's amount before the balance goes to zero. Actual fee is computed by the wallet. Default: 1.

**minamount**

Minimum amount of coins to send. Default: 2.

**stingyamount**

Amount of coins to send when the balance is low or giveaway rate is high. Default: 0.

**lowbalance**

When the balance drops to or below this amount, switch to **stingyamount** and send an alert if configured. If **stingyamount** is zero, keep using regular **amount**. Default: 0.

**ipclaiminterval**

Minimum interval between claims from the same IP address (for IPv4) or /64 subnet (for IPv6). Intervals for larger subnets are also enforced:

* 1/16 of **ipclaiminterval** between claims from the same /24 IPv4 or /56 IPv6 subnet,
* 1/256 of **ipclaiminterval** between claims from the same /16 IPv4 or /48 IPv6 subnet,

and so on. When **ipclaiminterval** is less than 1 second, intervals are not enforced. Default: 0s.

**ratelimit**

Soft limit on total giveaway rate, specified as **amount** of coins per **period**. When this rate is exceeded, send an alert if configured and switch to **stingyamount** until the rate drops. If **stingyamount** is zero, faucet will be temporarily paused. This feature is disabled when not configured.

**ratelimit**/**amount**

Maximum total giveaway amount during the period. Default: 0.

**ratelimit**/**period**

Period over which the amount is computed. Default: 0s.

**tokenkey**

Cryptographic key used to derive CSRF token from IP address and time. It can be 16 bytes encoded in Base64 with "!!binary" tag. Example:

    tokenkey: !!binary i1pLUHreQLj7MCDZjVX4Mw==

Other unspecified formats may be recognized but not recommended. When the key is absent or empty, CSRF tokens will not be used. Default empty. **config create** subcommand generates a random key.

**addressversions**

An array of accepted cryptocurrency address version values. When this parameter is absent or empty, addresses will not be validated by back-end service (but will be validated by the wallet). Default empty. **config create** subcommand sets **addressversions** to Dogecoin testnet versions: [113,196].

**alertprogram**

A program to execute when alert conditions are triggered. On low balance it will be executed as follows:

*alertprogram* balance *balance*

On rate limit it will be executed as follows:

*alertprogram* rate *amount* *period_in_seconds*

Shell commands and additional program arguments are not supported. When this parameter is absent or empty, alerts are disabled. Default: "".

**listen**

Address and TCP port separated by colon to listen for HTTP requests on. When the address is absent, listen on all addresses:

    listen: :8080

When the port number is absent, use default HTTP or HTTPS port:

    listen: ""

Both numeric forms and names are accepted:

    listen: 127.0.0.1:80

    listen: localhost:http

To specify IPv6 address with colons, enclose it in square brackets:

    listen: [::1]:8080

Default: "".

**certfile**

A file with TLS certificate chain for HTTPS. When it is not set or empty, use plain HTTP. Default: "".

**keyfile**

A file with private key for TLS certificate. When it is not set or empty, use plain HTTP. If the certificate and the key are in the same file, set both **certfile** and **keyfile** to the path to that file. Default: "".

**apiprefix**

Prefix (virtual directory) for API endpoints. It should begin with a slash. When empty, API endpoints will be at the web root. Default: "/api".

**pubdir**

A directory with static files to be served via HTTP. Front-end files can be placed there. When this parameter is absent or empty, static files will not be served. Default: "".

**alloworigin**

Value of Access-Control-Allow-Origin HTTP response header. When this parameter is set, CORS headers are sent with all HTTP responses. When this parameter is absent or empty, CORS headers are not sent. Default: "".

**usefwdaddr**

When this is true, use client address from X-Forwarded-For header (if it's present) instead of connecting IP address for rate limiting purposes. It should be used when the service is behind HTTP proxy server. Default: false.

**db**

SQL database to store persistent faucet data (claim log). When not configured, needed data will be stored in memory and will be lost when the service is restarted or stopped.

**db**/**driver**

Driver/connector for accessing the database. Supported driver: sqlite3. Default: "".

**db**/**source**

Database name and parameters specified as DSN string. For sqlite3 it's database file name. Default: "".

**rpc**

Wallet RPC server address and credentials.

**rpc**/**url**

RPC server address in form of URL. Default: "http\://localhost:44555".

**rpc**/**username**

HTTP authorization user name. It is used when cookie file is not configured or cannot be read. Default: "".

**rpc**/**password**

HTTP authorization password. It is used when cookie file is not configured or cannot be read. Default: "".

**rpc**/**cookiefile**

A file with RPC user name and password. When this parameter is not set or the file cannot be read, **username** and **password** parameters will be used instead. Default: "". **config create** subcommand sets this to default location of Dogecoin Core testnet cookie file. Remove this parameter or set to empty string if not using cookie file.

**log**

Format of log messages that are output to stderr.

**log**/**date**

Prefix log messages with date. You may want to set it to false if stderr is redirected to a logger that adds timestamps to messages. Default: true.

**log**/**time**

Prefix log messages with time. You may want to set it to false if stderr is redirected to a logger that adds timestamps to messages. Default: true.

**log**/**microseconds**

Use time with microsecond resolution. Default: false.

**log**/**utc**

Use UTC for date and time. When false, use local time. Default: false.
