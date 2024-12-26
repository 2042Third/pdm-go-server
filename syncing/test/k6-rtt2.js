import { sleep } from 'k6';
import { Trend, Counter, Rate } from 'k6/metrics';
import ws from 'k6/ws';

const rttMetric = new Trend('websocket_rtt');
const messagesReceived = new Counter('messages_received');
const messagesSent = new Counter('messages_sent');
const bytesSent = new Counter('bytes_sent');
const connectionRate = Rate('connection_success');
const activeConnections = new Counter('active_connections');

export const options = {
  vus: 2000,
  duration: '60s',
  thresholds: {
    'connection_success': ['rate>0.95'], // 95% connection success rate
    'active_connections': ['count>1500']  // Maintain at least 1500 active connections
  }
};

function generateLargePayload(messageId) {
  return {
    messageId: messageId,
    userId: `${__VU}`,
    documentId: `doc-${__VU}`,
    timestamp: Date.now(),
    content: 'A'.repeat(10000), // 10KB of base content
    changes: Array(100).fill(null).map(() => ({
      position: Math.floor(Math.random() * 10000),
      text: 'B'.repeat(100)
    }))
  };
}

export default function () {
  const userId = `${__VU}`;
  // const url = `wss://yangyi.dev/ws?userId=${userId}`;
  const url = `ws://localhost/ws?userId=${userId}`;

  const connectSuccess = new Promise((resolve, reject) => {
    const conn = ws.connect(url, null, function (socket) {
      activeConnections.add(1);

      socket.on('open', () => {
        connectionRate.add(1);
        resolve(true);

        // Send initial burst
        for (let i = 0; i < 10; i++) {
          sendMessage(socket, i);
          sleep(0.1); // Small delay between burst messages
        }

        // Continuous message sending
        let messageCount = 10;
        const sendInterval = setInterval(() => {
          if (socket.readyState === 1) {
            sendMessage(socket, messageCount++);
          } else {
            clearInterval(sendInterval);
          }
        }, 200); // Send every 200ms

        // Keep connection alive
        const pingInterval = setInterval(() => {
          if (socket.readyState === 1) {
            socket.ping();
          } else {
            clearInterval(pingInterval);
          }
        }, 5000);

        // Cleanup after 50 seconds
        setTimeout(() => {
          clearInterval(sendInterval);
          clearInterval(pingInterval);
          socket.close();
          activeConnections.add(-1);
        }, 50000);
      });

      socket.on('message', (data) => {
        messagesReceived.add(1);
      });

      socket.on('close', () => {
        activeConnections.add(-1);
      });

      socket.on('error', (e) => {
        console.error(`Socket error: ${e.message}`);
        activeConnections.add(-1);
      });
    });

    // Connection timeout
    setTimeout(() => {
      connectionRate.add(0);
      reject(new Error('Connection timeout'));
    }, 10000);
  });

  return connectSuccess;
}

function sendMessage(socket, messageId) {
  try {
    const payload = generateLargePayload(messageId);
    const message = JSON.stringify(payload);
    socket.send(message);
    messagesSent.add(1);
    bytesSent.add(message.length);
  } catch (e) {
    console.error(`Failed to send message: ${e.message}`);
  }
}