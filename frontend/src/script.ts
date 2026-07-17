import 'bootstrap/dist/css/bootstrap.min.css';
import 'bootstrap/dist/js/bootstrap.bundle.min.js';
import { WebSocketService, WSMessage } from './services/websocket';
import { processNatsMsg } from './components/topicTree';

const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
const wsUrl = window.location.port === '8080'
  ? 'ws://localhost:3000/ws'
  : `${protocol}//${window.location.host}/ws`;

const siteFilter = document.getElementById("site-filter") as HTMLSelectElement;

function applySiteFilter(selectedSite: string) {
  const rootWrapper = document.querySelector("#messages > .topic-wrapper");
  if (!rootWrapper) return;

  const topLevelTopics = rootWrapper.querySelectorAll(":scope > .topic");
  topLevelTopics.forEach(node => {
    const id = node.id; // ex: "nats_topic.iot_domain_brazil"
    if (!selectedSite) {
      node.classList.remove("hide");
    } else {
      if (id.endsWith("_" + selectedSite)) {
        node.classList.remove("hide");
      } else {
        node.classList.add("hide");
      }
    }
  });

  // if the selected element is under a hidden node, deselect it and clear the details panel
  const selectedTitle = document.querySelector(".title-container.selected");
  if (selectedTitle) {
    let current: HTMLElement | null = selectedTitle as HTMLElement;
    let isHidden = false;
    while (current && current.id !== "messages") {
      if (current.classList.contains("hide")) {
        isHidden = true;
        break;
      }
      current = current.parentElement;
    }
    if (isHidden) {
      selectedTitle.classList.remove("selected");
      const topicDetailName = document.getElementById("topic-name");
      const messageValue = document.getElementById("message-value");
      if (topicDetailName) topicDetailName.textContent = "Nenhum tópico selecionado";
      if (messageValue) messageValue.innerHTML = "";
    }
  }
}

if (siteFilter) {
  siteFilter.addEventListener("change", () => {
    const value = siteFilter.value;
    applySiteFilter(value);
    
    // send the subscription command to the backend via WebSocket
    wsService.send(JSON.stringify({
      action: "subscribe",
      site: value
    }));
  });
}

function handleIncomingMessage(msg: WSMessage) {
  // 1. process the message in the topic tree
  processNatsMsg(msg.subject, msg.data);

  // 2. populate the site filter based on the message subject
  const parts = msg.subject.split(".");
  if (parts.length > 0 && siteFilter) {
    const firstSubtopic = parts[0];
    const subParts = firstSubtopic.split("_");
    if (subParts.length >= 3) {
      const siteTopic = subParts[2]; // ex: "brazil" from "iot_domain_brazil"
      const exists = Array.from(siteFilter.options).some(opcao => opcao.value === siteTopic);
      if (!exists) {
        siteFilter.add(new Option(siteTopic, siteTopic));
      }
    }
  }
}

// initialize the connection with the backend WebSocket
const wsService = new WebSocketService(wsUrl, handleIncomingMessage);
wsService.connect();
