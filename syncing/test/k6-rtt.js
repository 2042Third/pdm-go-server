import { sleep } from 'k6';
import { Trend } from 'k6/metrics';
import ws from 'k6/ws';

// Create a custom metric to track RTT
const rttMetric = new Trend('websocket_rtt');

// Store message timestamps
const messageTimestamps = new Map();

export default function () {
  const url = 'ws://localhost:8082/ws';

  return ws.connect(url, {}, function (socket) {
    socket.on('open', () => {
      console.log('WebSocket connection established');
      // Send first message
      sendMessage(socket, 1);
    });

    socket.on('message', (data) => {
      const endTime = Date.now();
      const messageId = data.toString(); // Convert message to string
      const startTime = messageTimestamps.get(messageId);

      if (startTime) {
        const duration = endTime - startTime;
        console.log(`Message ${messageId} RTT: ${duration}ms`);
        rttMetric.add(duration);
        messageTimestamps.delete(messageId);

        // Send next message if we haven't sent 10 yet
        const nextMessageId = parseInt(messageId) + 1;
        if (nextMessageId <= 10) {
          sendMessage(socket, nextMessageId);
        } else {
          socket.close();
        }
      }
    });

    socket.on('error', (e) => {
      console.error('WebSocket error:', e);
    });

    socket.on('close', () => {
      console.log('WebSocket connection closed');
    });

    // Set a timeout to close the connection if it hangs
    setTimeout(function () {
      socket.close();
    }, 10000);
  });
}

function sendMessage(socket, messageId) {
  messageTimestamps.set(messageId.toString(), Date.now());
  socket.send(messageId.toString());
}

export function handleSummary(data) {
  if (!data.metrics.websocket_rtt) {
    console.log('No RTT metrics collected');
    return {
      stdout: JSON.stringify({
        error: 'No RTT metrics collected'
      })
    };
  }

  return {
    stdout: JSON.stringify({
      avg_rtt: data.metrics.websocket_rtt.avg,
      min_rtt: data.metrics.websocket_rtt.min,
      max_rtt: data.metrics.websocket_rtt.max,
      p95_rtt: data.metrics.websocket_rtt.p(95)
    }, null, 2)
  };
}