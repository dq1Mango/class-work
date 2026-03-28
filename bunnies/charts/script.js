var chartDiv = document.getElementById("chart");

function fetchHtml(filename) {
  fetch(filename)
    .then((response) => {
      return response.text();
    })
    .then((html) => {
      chartDiv.innerHTML = html;
    });
}

fetchHtml("/race.html");
