export function animateFlash(element: HTMLElement | null) {
  if (!element) return;

  // cancel running animations to avoid conflict
  element.getAnimations().forEach(a => a.cancel());

  const flashKeyframes = [
    { backgroundColor: "rgba(99, 102, 241, 0.25)" },
    { backgroundColor: "transparent" },
  ];

  const flashOptions = {
    duration: 150,
    easing: "ease-out"
  };

  element.animate(flashKeyframes, flashOptions);
}
