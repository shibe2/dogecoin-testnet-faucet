function validateAddr() {
    var address = document.forms["testnetAddrForm"]["testnetAddr"].value;
  
    console.log(address);
  
    if (address.charAt(0) == "m" || address.charAt(0) == "n") {
      return true;
    } else {
      alert("Please enter a valid Testnet address");
      return false;
    }
  
}