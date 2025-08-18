const beforeRequest = (event) => {
  const gameData = localStorage.getItem("gameData");

  if (gameData) {
    event.detail.formData.set("gameData", gameData);
  }
};

const afterRequest = () => {
  console.log("afterRequest")
  const gameDataContainer = document.getElementById("game-data")

  if (!gameDataContainer || gameDataContainer.textContent === "")
    return

  localStorage.setItem("gameData", gameDataContainer.textContent);
};