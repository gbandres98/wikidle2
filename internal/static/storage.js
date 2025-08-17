const beforeRequest = (event) => {
  const gameData = localStorage.getItem("gameData");

  console.log("beforeRequest");
  console.log(event);

  if (gameData) {
    event.detail.formData.set("gameData", gameData);
  }
};

const afterRequest = () => {
  const gameDataContainer = document.getElementById("game-data")

  if (!gameDataContainer || gameDataContainer.textContent === "")
    return

  localStorage.setItem("gameData", gameDataContainer.textContent);
};