var token = "";

function claim(address) {
  const data = { recipient: address, token: token };

  fetch('http://localhost:8000/claim', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(data),
  })
  .then (
    function(response) {
      var error = document.getElementById("error");
      var errorText = document.getElementById("errorText");
      var success = document.getElementById("success");
      var successText = document.getElementById("successText");

      if (response.status == 200) {
        error.style.display = "none";
        success.style.display = "block";
        successText.innerHTML = "Dogecoin sent.";
      } else {
        success.style.display = "none";
        error.style.display = "block";
        errorText.innerHTML = "API Error";
      }
    }
  )

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
  fetch("http://localhost:8000/info")
    .then (
      function(response) {
        if (response.status !== 200) {
          console.log(response.status);
        }

        response.json()
          .then (
            function(data) {

              var claimAmount = document.getElementById("claimAmount");
              claimAmount.innerHTML = "Current claim amount: " + data.amount;

              token = data.token;
            }
          )
      }
    )
}

getClaimAmount();