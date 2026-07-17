export interface WSMessage {
  subject: string;
  data: string;
}

export class WebSocketService {
  private url: string;
  private socket: WebSocket | null = null;
  private onMessageCallback: (msg: WSMessage) => void;
  private reconnectInterval = 3000;
  private isConnecting = false;

  constructor(url: string, onMessage: (msg: WSMessage) => void) {
    this.url = url;
    this.onMessageCallback = onMessage;
  }

  connect() {
    if (this.socket && (this.socket.readyState === WebSocket.OPEN || this.socket.readyState === WebSocket.CONNECTING)) {
      return;
    }

    this.isConnecting = true;
    console.log(`Connecting to backend WebSocket at ${this.url}...`);

    try {
      this.socket = new WebSocket(this.url);

      this.socket.onopen = () => {
        console.log("Successfully connected to backend WebSocket!");
        this.isConnecting = false;
      };

      this.socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as WSMessage;
          if (msg.subject && msg.data) {
            this.onMessageCallback(msg);
          }
        } catch (e) {
          console.error("Error parsing WebSocket JSON message:", e);
        }
      };

      this.socket.onerror = (error) => {
        console.error("WebSocket error:", error);
      };

      this.socket.onclose = () => {
        console.log("WebSocket connection closed. Attempting to reconnect...");
        this.isConnecting = false;
        setTimeout(() => this.connect(), this.reconnectInterval);
      };
    } catch (error) {
      console.error("Error attempting to connect to WebSocket:", error);
      this.isConnecting = false;
      setTimeout(() => this.connect(), this.reconnectInterval);
    }
  }

  send(data: string) {
    if (this.socket && this.socket.readyState === WebSocket.OPEN) {
      this.socket.send(data);
    } else {
      console.warn("Attempted to send data over a closed or uninitialized WebSocket.");
    }
  }

  close() {
    if (this.socket) {
      this.socket.close();
    }
  }
}
