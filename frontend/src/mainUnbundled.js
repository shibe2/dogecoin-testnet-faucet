var token = "";
var responseFromClaim = "";
var responseStatus;
var addressVersions;

var URL_BACKEND = 'http://localhost:8000';

var bs58check = require("bs58check");

var error = document.getElementById("error");
var errorText = document.getElementById("errorText");
var success = document.getElementById("success");
var successText = document.getElementById("successText");
var claimAmount = document.getElementById("claimAmount");
var submitButton = document.getElementById("submitButton");

function errorForValidateAddr(errorString) {
  success.style.display = "none";
  error.style.display = "block";
  errorText.innerHTML = errorString;
}

function getWaitDuration(futureDate) {
  var currentDate = new Date();
  var difference = Math.round((futureDate.getTime() - currentDate.getTime()) / 1000 / 60);

  if (Math.sign(difference) === -1) {
    success.style.display = "none";
    error.style.display = "block";
    errorText.innerHTML = "Your device's clock is not set properly. Please set your clock to the correct time and try again.";

    return false;
  }

  var hours = Math.floor(difference / 60);
  var minutes = difference % 60;

  if (difference < 60 && difference >= 1) {
    return `${minutes}m`;
  } 
  if (difference < 1) {
    return "1m";
  } 

  return `${hours}h ${minutes}m`;
}

function checkError(data) {
  
  switch (responseStatus) {
    case 503:
      success.style.display = "none";
      error.style.display = "block";
      submitButton.disabled = true;
      
      switch (data.error) {
        case "ServiceUnavailable":
          errorText.innerHTML = "The faucet is currently unavailable.";
          claimAmount.innerHTML = "Current claim amount: Faucet unavailable";
          break;

        case "NoFunds":
          errorText.innerHTML = "The faucet is out of funds.";
          claimAmount.innerHTML = "Current claim amount: Faucet out of funds.";
          break;

        case "ServicePaused":
          errorText.innerHTML = "The faucet is paused.";
          claimAmount.innerHTML = "Current claim amount: Faucet paused.";
      }

      return true;

    case 500:
      success.style.display = "none";
      error.style.display = "block";
      submitButton.disabled = true;

      switch (data.error) {
        case "FailedToSend":
          errorText.innerHTML = "Transaction has failed.";
          break;

        case "InternalError":
          errorText.innerHTML = "The faucet is experiencing an internal error.";
          claimAmount.innerHTML = "Current claim amount: Faucet unavailable";
          break;
      }

      return true;

    case 400:
      success.style.display = "none";
      error.style.display = "block";

      switch (data.requestErrors[0]["error"]) {
        case "InvalidValue":
          errorText.innerHTML = "The address you entered is invalid. Please check your address for errors or enter a new one.";
          break;
        
        case "InvalidFormat":
          errorText.innerHTML = "This should not normally happen. Check the console for details.";
          console.log(data);
          break;

        case "MissingValue":
          errorText.innerHTML = "This should not normally happen. Check the console for details.";
          console.log(data);
          break;
      }

      return true;
  
  }
  return false;
}

function claim(address) {
  const data = { recipient: address, token: token };

  fetch(`${URL_BACKEND}/claim`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  .then (response => {
    responseFromClaim = response;
    responseStatus = response.status;
    return response.json()
  })
  .then (data => {
    if (checkError(data)) {
      console.log("here");
    } else {

      var transationLink = document.getElementById("transactionLink");

      console.log(data);

      switch (responseFromClaim.status) {
        case 200:
          error.style.display = "none";
          success.style.display = "block";
          successText.innerHTML = data.amount + " Dogecoin sent."
          transactionLink.href = "https://sochain.com/tx/DOGETEST/" + data.txid;

          break;

        case 403:
          switch (data.rejectReason) {
            case "MustWait":
              success.style.display = "none";
              error.style.display = "block";

              var waitTime = new Date (data.wait);
              var timeToWaitString = getWaitDuration(waitTime);
              errorText.innerHTML = `Please wait ${timeToWaitString} before claiming again.`;

              break;

            case "InvalidToken":
              success.style.display = "none";
              error.style.display = "block";
              errorText.innerHTML = "Token is invalid. Please refresh.";
          }
      }
    }
  })

}

function validateAddr() {
  var address = document.forms["testnetAddrForm"]["testnetAddr"].value;

  console.log(address);

  try {
    var decoded = bs58check.decode(address);
  } catch (error) {
    console.log("error");
    
    errorForValidateAddr("Please enter a valid Testnet address.");

    return false;
  }

  if (decoded.length === 21) {
    if (!addressVersions) {
      console.log("claiming on first checks");
      claim(address);
    } else {
      for (var i = 0; i < addressVersions.length; i++) {
        if (addressVersions[i] == decoded[0]) {
          console.log("claiming on addr version");
          claim(address);
          return false;
        } else {
          errorForValidateAddr(`Incorrect address version. Correct versions are ${addressVersions}.`);
        }
      }
    }
  } else {
    errorForValidateAddr("Please enter a valid Testnet address.");
  }

  return false; // stop form submission so page doesn't refresh
}

function getClaimAmount() {
  fetch(`${URL_BACKEND}/info`, {
    headers: {}
  })
  .then (
    function(response) {
      if (response.status !== 200) {
        console.log(response.status);
      }

      responseStatus = response.status;

      response.json()
      .then (
        function(data) {
          // delete data["wait"];
          console.log(data);

          var claimAmount = document.getElementById("claimAmount");
          claimAmount.innerHTML = "Current claim amount: " + data.amount;

          token = data.token;
          addressVersions = data.addressVersions;

          checkError(data);

          if (data.wait) {
            var submitButton = document.getElementById("submitButton");

            var waitTime = new Date (data.wait);

            if (getWaitDuration(waitTime)) {
              var timeToWaitString = getWaitDuration(waitTime);

              submitButton.innerHTML = `Wait ${timeToWaitString} for next claim.`;
              submitButton.disabled = true;
            } else {
              submitButton.innerHTML = `Clock out of sync.`;
              submitButton.disabled = true;
            }
          }
        }
      )
    }
  )
}

getClaimAmount();

window.validateAddr = validateAddr;