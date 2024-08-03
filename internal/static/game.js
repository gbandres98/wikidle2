const onGameWin = () => {
  document.querySelector("#word-input-wrapper").remove();
  document.querySelector("#attempts").remove();
  document.querySelector(".controls").addEventListener("click", toggleModal);
  toggleModal();
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

document.addEventListener("DOMContentLoaded", () => {
  document.querySelector("#attempts").scroll(0, 999999999);
});

// Modal

const isOpenClass = "modal-is-open";
const openingClass = "modal-is-opening";
const closingClass = "modal-is-closing";
const scrollbarWidthCssVar = "--pico-scrollbar-width";
const animationDuration = 400; // ms
let visibleModal = null;

const toggleModal = () => {
  const modal = document.getElementById("game-win-modal");
  if (!modal) return;
  modal && (modal.open ? closeModal(modal) : openModal(modal));
};

const openModal = (modal) => {
  const { documentElement: html } = document;
  const scrollbarWidth = getScrollbarWidth();
  if (scrollbarWidth) {
    html.style.setProperty(scrollbarWidthCssVar, `${scrollbarWidth}px`);
  }
  html.classList.add(isOpenClass, openingClass);
  setTimeout(() => {
    visibleModal = modal;
    html.classList.remove(openingClass);
  }, animationDuration);
  modal.showModal();
};

const closeModal = (modal) => {
  visibleModal = null;
  const { documentElement: html } = document;
  html.classList.add(closingClass);
  setTimeout(() => {
    html.classList.remove(closingClass, isOpenClass);
    html.style.removeProperty(scrollbarWidthCssVar);
    modal.close();
  }, animationDuration);
};

// Close with a click outside
document.addEventListener("click", (event) => {
  if (visibleModal === null) return;
  const modalContent = visibleModal.querySelector("article");
  const isClickInside = modalContent.contains(event.target);
  !isClickInside && closeModal(visibleModal);
});

// Close with Esc key
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape" && visibleModal) {
    closeModal(visibleModal);
  }
});

// Get scrollbar width
const getScrollbarWidth = () => {
  const scrollbarWidth =
    window.innerWidth - document.documentElement.clientWidth;
  return scrollbarWidth;
};

// Is scrollbar visible
const isScrollbarVisible = () => {
  return document.body.scrollHeight > screen.height;
};

const scrollToTop = () => {
  window.scrollTo({ top: 0, behavior: "smooth" });
};

document.addEventListener("scroll", () => {
  const scrollButton = document.getElementById("up-button");
  if (scrollButton) {
    if (window.scrollY > 200) {
      scrollButton.classList.remove("hidden");
    } else {
      scrollButton.classList.add("hidden");
    }
  }
});

const unhighlightWords = () => {
  document.querySelectorAll(`.highlight`).forEach((e) => {
    e.classList.remove("highlight");
  });
};
