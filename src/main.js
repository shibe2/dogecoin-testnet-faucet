function validateAddr() {
    var address = document.forms["testnetAddrForm"]["testnetAddr"].value;
  
    console.log(address);

    var error = document.getElementById("error");
    var errorText = document.getElementById("errorText");
    var success = document.getElementById("success");
    var successText = document.getElementById("successText");
  
    if (address.charAt(0) == "m" || address.charAt(0) == "n") {
      // success
      error.style.display = "none";
      success.style.display = "block";
      successText.innerHTML = "Dogecoin sent.";
    } else {
      // error
      success.style.display = "none";
      error.style.display = "block";
      errorText.innerHTML = "Please enter a valid Testnet address.";
    }

    return false; // stop form submission so page doesn't refresh
}