# Chat Connection Quick Reference

Use the WebRTC data channel label `bb-chat-v1` to connect to per-stream chat.

## What the server expects

- Data channel label: `bb-chat-v1`
- Outbound client message type: `chat.send`
- `text` length: 1-2000 chars
- `displayName` length: 1-80 chars

Client -> server payload:

```json
{
  "type": "chat.send",
  "clientMsgId": "uuid-or-any-unique-id",
  "text": "hello",
  "displayName": "alice"
}
```

Server -> client message types:

- `chat.connected`
- `chat.history` (contains `{ type: "message", message: ... }[]`)
- `chat.message` (single live message)
- `chat.ack` (echoes `clientMsgId`)
- `chat.error` (may include `clientMsgId`)

Message shape:

```json
{
  "id": "message-id",
  "ts": 1730000000,
  "text": "hello",
  "displayName": "alice"
}
```

## Simple client example

```ts
const CHAT_LABEL = "bb-chat-v1";

type Outbound =
  | { type: "chat.connected" }
  | { type: "chat.history"; events: Array<{ type: "message"; message: ChatMessage }> }
  | { type: "chat.message"; eventId: number; message: ChatMessage }
  | { type: "chat.ack"; clientMsgId: string }
  | { type: "chat.error"; error: string; clientMsgId?: string };

type ChatMessage = {
  id: string;
  ts: number;
  text: string;
  displayName: string;
};

const chatChannel = peerConnection.createDataChannel(CHAT_LABEL);
const pending = new Map<string, (error?: Error) => void>();

chatChannel.addEventListener("message", (event) => {
  const payload = JSON.parse(event.data) as Outbound;

  if (payload.type === "chat.history") {
    payload.events.forEach((e) => console.log("history", e.message));
  }

  if (payload.type === "chat.message") {
    console.log("live", payload.message);
  }

  if (payload.type === "chat.ack") {
    pending.get(payload.clientMsgId)?.();
    pending.delete(payload.clientMsgId);
  }

  if (payload.type === "chat.error" && payload.clientMsgId) {
    pending.get(payload.clientMsgId)?.(new Error(payload.error));
    pending.delete(payload.clientMsgId);
  }
});

function sendChat(text: string, displayName: string) {
  if (chatChannel.readyState !== "open") {
    throw new Error("chat channel is not open");
  }

  const clientMsgId = crypto.randomUUID();
  const payload = {
    type: "chat.send",
    clientMsgId,
    text,
    displayName,
  };

  return new Promise<void>((resolve, reject) => {
    pending.set(clientMsgId, (error?: Error) => {
      if (error) reject(error);
      else resolve();
    });

    chatChannel.send(JSON.stringify(payload));
  });
}
```

## Connection flow

1. Create `RTCPeerConnection`.
2. Create chat data channel with label `bb-chat-v1` before SDP offer.
3. Handle inbound `chat.connected`, `chat.history`, `chat.message`, `chat.ack`, and `chat.error` payloads.
4. Send messages as `{ "type": "chat.send", "clientMsgId", "text", "displayName" }`.
5. Treat `chat.ack` as send success and `chat.error` as send failure.
