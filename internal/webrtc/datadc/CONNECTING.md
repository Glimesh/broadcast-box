# Raw Data Channel Quick Reference

Use the WebRTC data channel label `bb-data-v1` to broadcast raw payloads to other active peers on the same stream.

## What the server expects

- Data channel label: `bb-data-v1`
- Arbitrary payloads, either text or binary
- Maximum payload size: 256 KiB (matches Chrome's default)

The server does not define an application protocol for this channel. Every message payload up to the maximum size is copied and forwarded as-is to currently open `bb-data-v1` channels for the same stream, excluding the sender.

## Simple client example

```ts
const DATA_LABEL = "bb-data-v1";

const dataChannel = peerConnection.createDataChannel(DATA_LABEL);

dataChannel.addEventListener("open", () => {
  dataChannel.send("hello");
  dataChannel.send(new Uint8Array([1, 2, 3]));
});

dataChannel.addEventListener("message", async (event) => {
  if (typeof event.data === "string") {
    console.log("text", event.data);
    return;
  }

  const bytes = new Uint8Array(await event.data.arrayBuffer());
  console.log("binary", bytes);
});
```

## Connection flow

1. Create `RTCPeerConnection`.
2. Create a data channel with label `bb-data-v1` before SDP offer.
3. Handle inbound messages with your app-specific protocol.
4. Send text or binary payloads directly.

Only peers with currently open `bb-data-v1` channels receive broadcasts. The server does not store payloads, acknowledge delivery, validate payload content, or replay messages to peers that connect later.
