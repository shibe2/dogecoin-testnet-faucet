const { request } = require("express");
var express = require("express");
var app = express();

var fs = require("fs");
var childProcess = require("child_process");

var webURL = "http://localhost:8080";

app.get("/getCommitHash", (req, res) => {
  console.log(`GET from ${req.connection.remoteAddress}`)

  fs.writeFile("commitHash", childProcess.execSync("git log --pretty=format:\"%h\" -1"), (err) => {if (err) throw err;});

  fs.readFile("commitHash", (err, data) => {
    if (err) {
      res.status(500).send({err});
    } else {
      res.setHeader('Access-Control-Allow-Origin', webURL);

      res.json([data.toString()]);
    }
  });
});

app.listen(8081, () => {
  console.log("API running on port 8081");
});