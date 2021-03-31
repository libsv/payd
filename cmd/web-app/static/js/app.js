window.addEventListener("load", function(event) {
    let keypads = document.getElementsByClassName('keypad');
    let output = document.getElementById("satoshis");
    let cancel = document.getElementById("cancel");
    let back = document.getElementById("back");


    cancel.addEventListener('click', () =>{
        output.innerText = ''
    })

    back.addEventListener('click',() =>{
        output.innerText =output.innerText.slice(0, -1);
    })

    const myFunction = function () {
        let attribute = this.innerText;
        output.innerText += attribute;
    };

    for (let i = 0; i < keypads.length; i++) {
        keypads[i].addEventListener('click', myFunction, false);
    }

});
