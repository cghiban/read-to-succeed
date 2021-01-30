
let modal = document.querySelector("#myModal"),
    span = document.querySelector('.close'),
    readers = document.querySelector('#readers'),
    reader = document.querySelector('select[name=reader]'),
    author = document.querySelector('input[name=author]'),
    title = document.querySelector('input[name=title]'),
    day = document.querySelector('input[name=day]'),
    duration = document.querySelector('input[name=duration]');

document.querySelector("#addentry").addEventListener('click', (ev) => {
    reader.value = readers.value;
    modal.style.display = "block";
});

readers.addEventListener("change", (ev) => {
    let v = ev.target.value;
    if (v === "") {
        document.location.href = "/";
    }
    else {
        document.location.href = "/?reader=" + v;
    }
});

document.querySelector('button').addEventListener('click', (ev) => {
    let form = document.querySelector('form'),
    formData = new FormData(form);
    console.log('valid: ', form.checkValidity());
    if (!form.checkValidity())
        return;
    //form.method = 'POST';
    //form.action = '/add';
    //form.submit();

    /*var request = new XMLHttpRequest();
    request.open("POST", "/add");
    request.onload = function(oEvent) {
        if (request.status == 200) {
          //oOutput.innerHTML = "Uploaded!";
          console.log("Uploaded");
        } else {
          //oOutput.innerHTML = "Error " + oReq.status + " occurred when trying to upload your file.<br \/>";
          console.log("Error " + request.status + " occurred when trying to upload your file.<br \/>");
        }
      };
    request.send(new FormData(form));
    ev.preventDefault();*/

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
        reader: formData.get("reader"),
        author: formData.get("author"),
        title: formData.get("title"),
        day: formData.get("day"),
        duration: parseInt(formData.get("duration"),10),
    };
    console.log(data);
    postData('/add', data)
    .then(data => {
        console.log(data); // JSON data parsed by `data.json()` call
        if (data && data.status === "ok") {
            document.location.href = "/?reader=" + formData.get("reader");
        }
    });

    ev.preventDefault();
}, false);
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

document.body.addEventListener('keyup', function(e) {
    if (e.key == "Escape") {
        //msg.textContent += 'Escape pressed:'
        modal.style.display = "none";
    }
});

