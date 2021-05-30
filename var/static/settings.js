
let modal;

document.querySelector("#newreader").addEventListener('click', (ev) => {
    modal = document.querySelector("#readerModal");
    modal.style.display = "block";
    document.querySelector('#addreader input[name=name]').focus();
    ev.preventDefault();
});

document.querySelector("#newgroup").addEventListener('click', (ev) => {
    //reader.value = readers.value;
    modal = document.querySelector("#groupModal");
    modal.style.display = "block";
    //document.querySelector('input[name=name]').focus();
    document.querySelector('#addgroup input[name=name]').focus();
    ev.preventDefault();
});

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
        //body: data
    });
    return response.json(); // parses JSON response into native JavaScript objects
}

//document.querySelectorAll('form button').addEventListener('click', (ev) => {
document.querySelectorAll('form button').forEach( btn => {
    btn.addEventListener('click', (ev) => {
        let action,
            form = btn.closest("form"),
            modal = btn.closest("div.modal"),
            formData = new FormData(form);

        action = form.id;
        //console.log("action: ", action);
        console.log('valid: ', form.checkValidity());
        if (!form.checkValidity()) {
            console.log("Form not valid!");
            return;
        }
        let data = {
            name: formData.get("name"),
        };
        console.log(action, data);

        ev.preventDefault();
        postData('/' + action, data)
            .then(data => {
                console.log(data); // JSON data parsed by `data.json()` call
                if (data && data.status === "ok") {
                    document.location.href = "/settings";
                }
                else if (data.message && data.message !== "") {
                    alert(data.message);
                }
            });

    }, false);
});

// When the user clicks on <span> (x), close the modal
document.querySelectorAll('span.close').forEach( s => {
    s.addEventListener('click', ev => {
        if (modal) {
            modal.style.display = "none";
        }
    });
});

// When the user clicks anywhere outside of the modal, close it
window.onclick = function(event) {
    if (event.target == modal) {
        modal.style.display = "none";
    }
}

document.body.addEventListener('keyup', function(e) {
    if (e.key == "Escape") {
        console.log("modal: ", modal);
        modal.style.display = "none";
    }
});
