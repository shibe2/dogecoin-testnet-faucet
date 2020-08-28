var token = "";
var responseFromClaim = "";
var responseStatus;

function check503(data) {
  var error = document.getElementById("error");
  var errorText = document.getElementById("errorText");
  var success = document.getElementById("success");
  var successText = document.getElementById("successText");
  var claimAmount = document.getElementById("claimAmount");
  var submitButton = document.getElementById("submitButton");

  if (responseStatus == 503) {
    success.style.display = "none";
    error.style.display = "block";
    submitButton.disabled = true;
    
    switch (data.error) {
      case "ServiceUnavailable":
        errorText.innerHTML = "The faucet is currently unavailable.";
        claimAmount.innerHTML = "Current claim amount: Faucet unavailable";

      case "NoFunds":
        errorText.innerHTML = "The faucet is out of funds.";
        claimAmount.innerHTML = "Current claim amount: Faucet out of funds.";

      case "ServicePaused":
        errorText.innerHTML = "The faucet is paused.";
        claimAmount.innerHTML = "Current claim amount: Faucet paused.";
    }
  }
}

function claim(address) {
  const data = { recipient: address, token: token };

  fetch('http://localhost:8000/claim', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  .then (response => {
    responseFromClaim = response;
    return response.json()
  })
  .then (data => {
    var error = document.getElementById("error");
    var errorText = document.getElementById("errorText");
    var success = document.getElementById("success");
    var successText = document.getElementById("successText");

    console.log(data);

    switch (responseFromClaim.status) {
      case 200:
        error.style.display = "none";
        success.style.display = "block";
        successText.innerHTML = data.amount + " Dogecoin sent.";

        break;

      case 403:
        switch (data.rejectReason) {
          case "MustWait":
            success.style.display = "none";
            error.style.display = "block";
            errorText.innerHTML = "Please wait 24 hours since your last claim until you claim again.";

            break;

          case "InvalidToken":
            success.style.display = "none";
            error.style.display = "block";
            errorText.innerHTML = "Token is invalid. Please refresh.";
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
  fetch("http://localhost:8000/info", {
    headers: {
      // debug with error codes here
    }
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

          var claimAmount = document.getElementById("claimAmount");
          claimAmount.innerHTML = "Current claim amount: " + data.amount;

          token = data.token;
          console.log(token);

          check503(data);
        }
      )
    }
  )
}

getClaimAmount();