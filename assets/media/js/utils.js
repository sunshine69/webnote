
/*
Usage example - note it understand html tag in popup
<div class="popup" onclick="ShowPopup('myPopup')">Click me to toggle the popup!
  <span class="popuptext" id="myPopup">A <i>Simple</i> Popup!</span>
</div> */

function ShowPopup(popID) {
    var popup = document.getElementById(popID);
    popup.classList.toggle("show");
}