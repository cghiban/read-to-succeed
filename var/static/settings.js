let modal;

document.querySelector("#newreader").addEventListener("click", (ev) => {
  modal = document.querySelector("#readerModal");
  modal.style.display = "block";
  document.querySelector("#addreader input[name=name]").focus();
  ev.preventDefault();
});

if (document.querySelector("#newgroup")) {
  document.querySelector("#newgroup").addEventListener("click", (ev) => {
    //reader.value = readers.value;
    modal = document.querySelector("#groupModal");
    modal.style.display = "block";
    //document.querySelector('input[name=name]').focus();
    document.querySelector("#addgroup input[name=name]").focus();
    ev.preventDefault();
  });
}

if (document.querySelector("#joinagroup")) {
  document.querySelector("#joinagroup").addEventListener("click", (ev) => {
    modal = document.querySelector("#joinGroupModal");
    modal.style.display = "block";
    document.querySelector("#findgroups input[name=name]").focus();
    ev.preventDefault();

    // populate select w/ readers
    let groupreader_select = document.querySelector("select[name=groupreader]");
    groupreader_select.childNodes.forEach((e) => e.remove());
    let tRows = document.querySelector("#readerstable").rows;
    for (let i = 1; i < tRows.length; i++) {
      const row = tRows[i],
        cells = row.cells;
      const o = document.createElement("option");
      o.innerText = cells[0].innerText;
      o.value = cells[0].getAttribute("data__reader_id");
      groupreader_select.appendChild(o);
    }
  });
}

document.querySelectorAll("#tgroupreaders a").forEach((a, i) => {
  if (i === 0) return;

  a.addEventListener("click", (ev) => {
    ev.preventDefault();
    if (!confirm("Are you sure?")) return;

    //let a = ev.target;
    //console.log(a.getAttribute("gid"), " -- ", a.getAttribute("rid"));
    let params = {
      group_id: parseInt(a.getAttribute("data-gid"), 10),
      reader_id: parseInt(a.getAttribute("data-rid"), 10),
    };
    postData("/leavegroup", params).then((data) => {
      console.log(data); // JSON data parsed by `data.json()` call
      if (data && data.status === "ok") {
        try {
          //a.remove();
          a.parentNode.remove();
        } catch {
          document.location.href = "/settings";
        }
      } else if (data.message && data.message !== "") {
        alert(data.message);
      }
    });
  });
});

async function postData(url = "", data = {}) {
  // Default options are marked with *
  const response = await fetch(url, {
    method: "POST", // *GET, POST, PUT, DELETE, etc.
    mode: "cors", // no-cors, *cors, same-origin
    cache: "no-cache", // *default, no-cache, reload, force-cache, only-if-cached
    credentials: "same-origin", // include, *same-origin, omit
    headers: {
      "Content-Type": "application/json", //'application/x-www-form-urlencoded',
    },
    redirect: "follow", // manual, *follow, error
    referrerPolicy: "no-referrer", // no-referrer, *no-referrer-when-downgrade, origin, origin-when-cross-origin, same-origin, strict-origin, strict-origin-when-cross-origin, unsafe-url
    body: JSON.stringify(data), // body data type must match "Content-Type" header
    //body: data
  });
  return response.json(); // parses JSON response into native JavaScript objects
}

//document.querySelectorAll('form button').addEventListener('click', (ev) => {
document.querySelectorAll("form button").forEach((btn) => {
  btn.addEventListener(
    "click",
    (ev) => {
      let action,
        form = btn.closest("form"),
        modal = btn.closest("div.modal"),
        formData = new FormData(form);

      action = form.id;
      //console.log("action: ", action);
      console.log("valid: ", form.checkValidity());
      if (!form.checkValidity()) {
        console.log("Form not valid!");
        return;
      }
      let data = {
        name: formData.get("name"),
      };
      console.log(action, data);

      ev.preventDefault();
      if (action === "findgroups") {
        data["reader"] = parseInt(formData.get("groupreader"), 10);
        doSearchGroups(data);
        return;
      }
      postData("/" + action, data).then((data) => {
        console.log(data); // JSON data parsed by `data.json()` call
        if (data && data.status === "ok") {
          document.location.href = "/settings";
        } else if (data.message && data.message !== "") {
          alert(data.message);
        }
      });
    },
    false,
  );
});

document.querySelectorAll("#usergroups input[type=checkbox]").forEach((cb) => {
  cb.addEventListener("change", (ev) => {
    console.log(cb, ev.target, cb.checked, cb.value);
    let gid = cb.value;

    let data = { status: cb.checked ? "public" : "private" };
    console.log(data);

    postData("/updategroup/" + gid, data).then((data) => {
      console.log(data); // JSON data parsed by `data.json()` call
      if (data && data.status === "ok") {
        document.location.href = "/settings";
      } else if (data.message && data.message !== "") {
        alert(data.message);
      }
    });
  });
});

//var _dbg;
let doSearchGroups = function (params) {
  //console.info("doing the search...", params);

  let grouplist_div = document.querySelector("#grouplist");
  grouplist_div.childNodes.forEach((e) => e.remove());

  postData("/findgroups", params).then((data) => {
    //console.log("---------");
    console.log(data);
    // _dbg = data;
    if (data.length == 0) {
      grouplist_div.innerHTML =
        "<p><small>Nothing found for this query</small></p>";
      return;
    }
    let ul = document.createElement("ul");
    //li = document.createElement("li");
    data.forEach((item) => {
      console.log(item);
      let li = document.createElement("li");
      li.innerHTML = item["name"] + " &nbsp; ";
      //let div_title = document.createElement("div");
      let add_link = document.createElement("a");
      add_link.appendChild(document.createTextNode("join group"));
      add_link.href = "#";
      li.appendChild(add_link);
      add_link.addEventListener("click", (e) => {
        e.preventDefault();
        console.log(" clicked.." + item["id"]);
        let reader = document.querySelector("select[name=groupreader]").value;
        if (!reader) {
          alert("Must pick a reader!");
          return;
        }
        e.target.style.display = "none";
        let args = {
          group: item["id"],
          reader: parseInt(reader, 10),
          query: params["name"],
        };

        postData("/joingroup", args).then((data) => {
          console.log(data);
          if (data && data.status === "ok") {
            document.location.href = "/settings";
          } else if (data.message && data.message !== "") {
            alert(data.message);
          }
        });
      });
      ul.appendChild(li);
    });
    grouplist_div.appendChild(ul);
  });
};

// When the user clicks on <span> (x), close the modal
document.querySelectorAll("span.close").forEach((s) => {
  s.addEventListener("click", (ev) => {
    if (modal) {
      modal.style.display = "none";
    }
  });
});

// When the user clicks anywhere outside of the modal, close it
window.onclick = function (event) {
  if (event.target == modal) {
    modal.style.display = "none";
  }
};

document.body.addEventListener("keyup", function (e) {
  if (e.key == "Escape" && modal) {
    console.log("modal: ", modal);
    modal.style.display = "none";
  }
});
