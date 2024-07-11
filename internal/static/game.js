const onGameWin = () => {
  document.querySelector("#word-input-wrapper").remove();
  document.querySelector("#attempts").remove();
};

const wordScrollPosition = new Map();

const scrollToNextWord = (index) => {
  if (!wordScrollPosition.has(index)) {
    wordScrollPosition.set(index, 0);
  }

  let word = document.querySelector(
    `.word-${index}-${wordScrollPosition.get(index)}`
  );
  if (!word) {
    wordScrollPosition.set(index, 0);
    word = document.querySelector(
      `.word-${index}-${wordScrollPosition.get(index)}`
    );
  }

  word.scrollIntoView({ behavior: "smooth", block: "center" });
  word.classList.remove("hit");
  setTimeout(() => word.classList.add("hit"), 100);
  wordScrollPosition.set(index, wordScrollPosition.get(index) + 1);
};
