
let modal = document.querySelector("#myModal"),
    span = document.querySelector('.close'),
    readers = document.querySelector('#readers');

document.querySelector("#addentry").addEventListener('click', e => {
    //modal.querySelector('h3').innerText = e.target.innerText;
    //video.src = e.target.getAttribute('data-file');
    let reader = document.querySelector('input[name=reader]');
    reader.value = readers.value;
    modal.style.display = "block";
});
//console.log(a.getAttribute('data-file'));

// When the user clicks on <span> (x), close the modal
span.onclick = function() {
    //video.pause();
    modal.style.display = "none";
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == modal) {
        //video.pause();
        modal.style.display = "none";
    }
}

