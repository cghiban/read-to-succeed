let modal = document.querySelector("#myModal"),
  span = document.querySelector(".close"),
  stats = document.querySelector("#stats"),
  readers = document.querySelector("#readers"),
  reader = document.querySelector("select[name=reader]"),
  author = document.querySelector("input[name=author]"),
  title = document.querySelector("input[name=title]"),
  day = document.querySelector("input[name=day]"),
  duration = document.querySelector("input[name=duration]"),
  pages = document.querySelector("input[name=pages]");

if (stats) {
  stats.hidden = true;
}

if (!reader) {
  reader = document.querySelector("input[name=reader][type=hidden]");
}

let addButtons = document.querySelectorAll(".addentry, #addentry");
addButtons.forEach((b) => {
  b.addEventListener("click", (ev) => {
    if (b.getAttribute("data-reader")) {
      reader.value = b.getAttribute("data-reader");
    } else {
      reader.value = readers.value;
    }
    if (b.getAttribute("data-author")) {
      author.value = b.getAttribute("data-author");
    } else {
      author.value = "";
    }
    if (b.getAttribute("data-title")) {
      title.value = b.getAttribute("data-title");
    } else {
      title.value = "";
    }
    modal.style.display = "block";
  });
});

if (document.querySelector("#togglestats")) {
  document.querySelector("#togglestats").addEventListener("click", (ev) => {
    stats.hidden = !stats.hidden;
  });
}

if (readers) {
  readers.addEventListener("change", (ev) => {
    let v = ev.target.value;
    if (v === "") {
      document.location.href = "/";
    } else {
      document.location.href = "/?reader=" + v;
    }
  });
}

async function postData(url = "", data = {}) {
  const response = await fetch(url, {
    method: "POST",
    mode: "cors",
    cache: "no-cache",
    credentials: "same-origin",
    headers: {
      "Content-Type": "application/json",
    },
    redirect: "follow",
    referrerPolicy: "no-referrer",
    body: JSON.stringify(data),
  });
  return response.json();
}

document.querySelector("form#addreading button").addEventListener(
  "click",
  (ev) => {
    let form = document.querySelector("form#addreading"),
      formData = new FormData(form);
    console.log("valid: ", form.checkValidity());
    if (!form.checkValidity()) return;

    let data = {
      reader: formData.get("reader").trim(),
      author: formData.get("author").trim(),
      title: formData.get("title").trim(),
      note: formData.get("note").trim(),
      day: formData.get("day"),
      duration: parseInt(formData.get("duration"), 10),
      pages: parseInt(formData.get("pages"), 10),
    };
    console.log(data);
    postData("/add", data).then((data) => {
      console.log(data); // JSON data parsed by `data.json()` call
      if (data && data.status === "ok") {
        document.location.href = "/?reader=" + formData.get("reader");
      }
    });

    ev.preventDefault();
  },
  false,
);

if (modal) {
  // When the user clicks on <span> (x), close the modal
  span.onclick = function () {
    modal.style.display = "none";
  };

  // When the user clicks anywhere outside of the modal, close it
  window.onclick = function (event) {
    if (event.target == modal) {
      modal.style.display = "none";
    }
  };

  document.body.addEventListener("keyup", function (e) {
    if (e.key == "Escape") {
      modal.style.display = "none";
    }
  });
}
