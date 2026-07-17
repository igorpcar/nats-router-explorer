import JSONFormatter from 'json-formatter-js';
import { animateFlash } from './animations';

export function createFormattedJson(value: string): HTMLDivElement {
  const valueJson = JSON.parse(value);
  const formatter = new JSONFormatter(valueJson, Infinity, {
    hoverPreviewEnabled: true
  });
  return formatter.render();
}

export function updateDetailsPanel(topic: string, value: string) {
  const messageValue = document.getElementById("message-value");
  if (messageValue) {
    messageValue.innerHTML = "";
    try {
      messageValue.appendChild(createFormattedJson(value));
    } catch (e) {
      // if it is not a valid JSON, display simple text
      const pre = document.createElement("pre");
      pre.textContent = value;
      messageValue.appendChild(pre);
    }
    animateFlash(messageValue);
  }

  const topicDetailName = document.getElementById("topic-name");
  if (topicDetailName) {
    topicDetailName.textContent = topic;
  }
}

export function selectTopicValueHandler(topic: string): (event: Event) => void {
  return function (event: Event) {
    const titleContainer = event.currentTarget as HTMLElement;

    // deselect the previously selected element
    const alreadySelected = document.querySelector(".title-container.selected");
    alreadySelected?.classList.remove("selected");

    // mark the current container as selected
    titleContainer.classList.add("selected");

    // find the corresponding value
    const valueSpan = titleContainer.querySelector(".topic-value");
    const valueText = valueSpan ? valueSpan.textContent : "";

    // update the details
    updateDetailsPanel(topic, valueText || "");
  };
}
