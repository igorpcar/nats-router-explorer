import { animateFlash } from '../ui/animations';
import { selectTopicValueHandler, updateDetailsPanel } from '../ui/details';

function createSvgIcon(type: 'folder' | 'tag'): SVGElement {
  const svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
  svg.setAttribute("width", "14");
  svg.setAttribute("height", "14");
  svg.setAttribute("viewBox", "0 0 24 24");
  svg.setAttribute("fill", "none");
  svg.setAttribute("stroke", "currentColor");
  svg.setAttribute("stroke-width", "2.5");
  svg.setAttribute("stroke-linecap", "round");
  svg.setAttribute("stroke-linejoin", "round");
  svg.classList.add("node-icon", type === 'folder' ? 'folder-icon' : 'tag-icon');

  if (type === 'folder') {
    const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path.setAttribute("d", "M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z");
    svg.appendChild(path);
  } else {
    const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
    path.setAttribute("d", "M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z");
    svg.appendChild(path);
  }
  return svg;
}

export function createTopicNode(id: string): HTMLDivElement {
  const topicNode = document.createElement("div");
  topicNode.id = id;
  topicNode.toggleAttribute("data-expand");
  topicNode.classList.add("topic");
  return topicNode;
}

export function createTitleContainer(subtopic: string, isLast: boolean): HTMLDivElement {
  const titleContainer = document.createElement("div");
  titleContainer.classList.add("title-container");

  const arrow = document.createElement("div");
  arrow.classList.add("arrow");
  arrow.textContent = "▼";

  const icon = createSvgIcon(isLast ? 'tag' : 'folder');

  const title = document.createElement("span");
  title.textContent = subtopic;
  title.className = "topic-title";

  titleContainer.appendChild(arrow);
  titleContainer.appendChild(icon);
  titleContainer.appendChild(title);

  return titleContainer;
}

export function createTopicValue(id: string, value: string): HTMLSpanElement {
  const valueSpan = document.createElement("span");
  valueSpan.id = id;
  valueSpan.textContent = value;
  valueSpan.classList.add("topic-value");
  return valueSpan;
}

export function toggleExpand(event: Event) {
  const clickedElement = event.currentTarget as HTMLElement;
  const parentTopic = clickedElement.closest(".topic");
  if (!parentTopic) return;

  const wrapper = parentTopic.querySelector(":scope > .topic-wrapper") as HTMLElement;
  if (!wrapper) return;

  if (parentTopic.hasAttribute("data-expand")) {
    const height = wrapper.scrollHeight;
    parentTopic.removeAttribute("data-expand");
    wrapper.animate([
      { height: `${height}px` },
      { height: "0px" }
    ], { duration: 150, easing: "ease-in-out" }).onfinish = () => {
      wrapper.classList.add("hide");
    };
  } else {
    wrapper.classList.remove("hide");
    parentTopic.setAttribute("data-expand", "");
    const height = wrapper.scrollHeight;
    wrapper.animate([
      { height: "0px" },
      { height: `${height}px` }
    ], { duration: 150, easing: "ease-in-out" });
  }
}

export function processNatsMsg(topic: string, value: string) {
  const subtopics = topic.split(".");
  let currentPath = "";
  const prefix = "nats_topic.";
  const prefixValue = "nats_value.";

  let messagesContainer = document.getElementById("messages");
  if (!messagesContainer) return;

  for (let i = 0; i < subtopics.length; i++) {
    const part = subtopics[i];
    const isLastElement = i === subtopics.length - 1;

    currentPath = currentPath === "" ? part : currentPath + "." + part;
    const elementId = prefix + currentPath;

    let topicNode = document.getElementById(elementId);

    // if the topic level does not exist yet, create it
    if (!topicNode) {
      topicNode = createTopicNode(elementId);
      
      // if it is the root level (i === 0), check if it should be hidden based on the selected filter
      if (i === 0) {
        const siteFilter = document.getElementById("site-filter") as HTMLSelectElement;
        if (siteFilter && siteFilter.value && !elementId.endsWith("_" + siteFilter.value)) {
          topicNode.classList.add("hide");
        }
      }

      const titleContainer = createTitleContainer(part, isLastElement);

      if (isLastElement) {
        titleContainer.classList.add("last-topic");
        const valueSpan = createTopicValue(prefixValue + currentPath, value);
        titleContainer.appendChild(valueSpan);
        titleContainer.addEventListener("click", selectTopicValueHandler(topic));
      } else {
        titleContainer.addEventListener("click", toggleExpand);
      }

      topicNode.appendChild(titleContainer);
      let wrapper = messagesContainer.querySelector(":scope > .topic-wrapper");
      if (!wrapper) {
        wrapper = document.createElement("div");
        wrapper.classList.add("topic-wrapper");
        wrapper.appendChild(topicNode);
        messagesContainer.appendChild(wrapper);
      } else {
        wrapper.appendChild(topicNode);
      }
    } 
    // if it already exists
    else {
      // if it is an intermediate node (parent) and is collapsed, make its header flash
      if (!isLastElement && !topicNode.hasAttribute("data-expand")) {
        const titleContainer = topicNode.querySelector(":scope > .title-container") as HTMLElement;
        if (titleContainer) {
          animateFlash(titleContainer);
        }
      }

      // if it is the last node, just update the value
      if (isLastElement) {
        const valueSpan = document.getElementById(prefixValue + currentPath);
        if (valueSpan) {
          valueSpan.textContent = value;
          animateFlash(valueSpan.parentElement);

          // if it is selected, also update the details panel
          if (valueSpan.parentElement?.classList.contains("selected")) {
            updateDetailsPanel(topic, value);
          }
        }
      }
    }

    messagesContainer = topicNode;
  }
}
