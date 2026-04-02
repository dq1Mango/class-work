const slider = document.getElementById("progress");
const output = document.getElementById("display");

output.innerHTML = slider.value;

slider.oninput = function () {
  output.innerHTML = this.value;
};
