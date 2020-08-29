# Dogecoin TestNet Faucet

A TestNet faucet for dogecoin.

## Development

Be sure you have the latest version of nodeJS installed.

###  Install dependencies

```
$ npm install
```

### Start localhost server

```
$ npm start
```

### Start mock API

Be sure to have Docker installed and open a new terminal.

```
$ make deps
$ make run
```

You can now use the api at localhost:8000
Example : 
```
$ curl http://localhost:8000/info
{
  "addressVersions": [
    113,
    196
  ],
  "amount": 100,
  "token": "WLXhQ7dxIzSNRMseNEFYA",
  "wait": "2000-01-23T04:56:07Z"
}
```

## Note

* Doesn't automatically reload when files are changed.
* Already include semantic-ui libraries.