var token = "";
var responseFromClaim = "";
var responseStatus;

var URL_BACKEND = 'http://localhost:8000';

function getWaitDuration(futureDate) {
  var currentDate = new Date();
  var difference = Math.round((currentDate.getTime() - futureDate.getTime()) / 1000 / 60);

  var hours = Math.abs(Math.floor(difference / 60));
  var minutes = Math.abs(difference % 60);

  return `${hours}:${minutes}`;
}

function checkError(data) {
  var error = document.getElementById("error");
  var errorText = document.getElementById("errorText");
  var success = document.getElementById("success");
  var claimAmount = document.getElementById("claimAmount");
  var submitButton = document.getElementById("submitButton");

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

      var error = document.getElementById("error");
      var errorText = document.getElementById("errorText");
      var success = document.getElementById("success");
      var successText = document.getElementById("successText");
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

  var error = document.getElementById("error");
  var errorText = document.getElementById("errorText");
  var success = document.getElementById("success");
  var successText = document.getElementById("successText");

  if (address.charAt(0) == "m" || address.charAt(0) == "n") {
    // success
    claim(address);
  } else {
    // error
    success.style.display = "none";
    error.style.display = "block";
    errorText.innerHTML = "Please enter a valid Testnet address.";
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

          checkError(data);

          if (data.wait) {
            var submitButton = document.getElementById("submitButton");

            var waitTime = new Date (data.wait);

            var timeToWaitString = getWaitDuration(waitTime);

            submitButton.innerHTML = `Wait until ${timeToWaitString} for next claim.`;
            submitButton.disabled = true;
          }
        }
      )
    }
  )
}

getClaimAmount();