const beforeRequest = (event) => {
  const gameData = localStorage.getItem("gameData");

  console.log(event);

  if (gameData) {
    event.detail.formData.set("gameData", gameData);
  }
};

document.addEventListener("afterResponse", (event) => {
  const gameData = event.detail.value;

  if (gameData) {
    localStorage.setItem("gameData", gameData);
  }
});