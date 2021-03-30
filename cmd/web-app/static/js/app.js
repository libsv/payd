window.addEventListener("load", function(event) {
    let keypads = document.getElementsByClassName('keypad');
    let output = document.getElementById("satoshis");

    const myFunction = function () {
        let attribute = this.innerText;
        output.innerText += attribute;
    };

    for (let i = 0; i < keypads.length; i++) {
        keypads[i].addEventListener('click', myFunction, false);
    }
});
