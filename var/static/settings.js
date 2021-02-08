
let modal = document.querySelector("#myModal"),
span = document.querySelector('.close'),
//readers = document.querySelector('#readers'),
reader = document.querySelector('input[name=readername]');

document.querySelector("#newreader").addEventListener('click', (ev) => {
    //reader.value = readers.value;
    modal.style.display = "block";
    reader.focus();
});


document.querySelector('form#addreader button').addEventListener('click', (ev) => {
    let form = document.querySelector('form#addreader'),
    formData = new FormData(form);
    console.log('valid: ', form.checkValidity());
    if (!form.checkValidity())
        return;

    async function postData(url = '', data = {}) {
        // Default options are marked with *
        const response = await fetch(url, {
            method: 'POST', // *GET, POST, PUT, DELETE, etc.
            mode: 'cors', // no-cors, *cors, same-origin
            cache: 'no-cache', // *default, no-cache, reload, force-cache, only-if-cached
            credentials: 'same-origin', // include, *same-origin, omit
            headers: {
            'Content-Type': 'application/json'
            // 'Content-Type': 'application/x-www-form-urlencoded',
            },
            redirect: 'follow', // manual, *follow, error
            referrerPolicy: 'no-referrer', // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
            body: JSON.stringify(data) // body data type must match "Content-Type" header
        });
        return response.json(); // parses JSON response into native JavaScript objects
    };
    let data = {
        name: formData.get("readername"),
    };
    //console.log(data);
    postData('/addreader', data)
    .then(data => {
        console.log(data); // JSON data parsed by `data.json()` call
        if (data && data.status === "ok") {
            document.location.href = "/settings";
        }
        else if (data.message && data.message !== "") {
            alert(data.message);
        }
    });

    ev.preventDefault();
}, false);

// When the user clicks on <span> (x), close the modal
span.onclick = function() {
    modal.style.display = "none";
}

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == modal) {
        //video.pause();
        modal.style.display = "none";
    }
}

document.body.addEventListener('keyup', function(e) {
    if (e.key == "Escape") {
        modal.style.display = "none";
    }
});
