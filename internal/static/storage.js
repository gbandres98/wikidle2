const beforeRequest = (event) => {
  const gameData = localStorage.getItem("gameData");

  if (gameData) {
    event.detail.headers["gameData"] = gameData;
  }
};

document.addEventListener("afterResponse", (event) => {
  const gameData = event.detail.value;

  if (gameData) {
    localStorage.setItem("gameData", gameData);
  }
});
