import { sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import ws from 'k6/ws';

const rttMetric = new Trend('websocket_rtt');
const messagesReceived = new Counter('messages_received');
const messagesSent = new Counter('messages_sent');
const bytesSent = new Counter('bytes_sent');

const messageTimestamps = new Map();

export const options = {
  vus: 1,
  iterations: 1,
  thresholds: {
    'websocket_rtt': ['p(95)<1000'],
    'messages_received': ['count>0'],
    'messages_sent': ['count>0']
  }
};

// Generate a large document-like payload
function generateLargePayload(messageId) {
  const paragraphs = Array(50).fill(null).map((_, i) => ({
    id: `p${i}`,
    text: 'Lorem ipsum '.repeat(100), // ~1KB per paragraph
    styles: {
      fontSize: '14px',
      color: '#000000',
      bold: true,
      italic: Math.random() > 0.5,
      underline: Math.random() > 0.5
    },
    changes: Array(20).fill(null).map((_, j) => ({
      position: j * 10,
      insert: 'New text '.repeat(5),
      delete: 10,
      attributes: { bold: true, italic: true, color: '#ff0000' }
    }))
  }));

  return {
    messageId: messageId,
    userId: `${__VU}`,
    documentId: `doc-${__VU}`,
    timestamp: Date.now(),
    type: 'document_update',
    content: paragraphs,
    metadata: {
      version: messageId,
      lastEditor: `user-${__VU}`,
      collaborators: Array(10).fill(null).map((_, i) => `user-${i}`),
      permissions: {
        readers: Array(50).fill(null).map((_, i) => `reader-${i}`),
        editors: Array(20).fill(null).map((_, i) => `editor-${i}`),
      }
    }
  };
}

export default function () {
  const userId = `${__VU}`;
  // const url = `wss://yangyi.dev/ws?userId=${userId}`;
  const url = `ws://localhost/ws?userId=${userId}`;
  let messageCount = 0;

  const res = ws.connect(url, null, function (socket) {
    socket.on('open', () => {
      console.log(`VU ${__VU}: Connected`);

      // Initial burst of messages
      for (let i = 0; i < 3; i++) {
        sendMessage(socket, messageCount++);
      }

      // Continue sending messages periodically
      const interval = setInterval(() => {
        if (messageCount < 20) { // Increased from 5 to 20 messages per client
          sendMessage(socket, messageCount++);
        } else {
          clearInterval(interval);
          socket.close(1000);
        }
      }, 1000); // Send a message every second
    });

    socket.on('message', (data) => {
      const endTime = Date.now();
      messagesReceived.add(1);

      const key = `${userId}-${messageCount}`;
      const startTime = messageTimestamps.get(key);
      if (startTime) {
        const rtt = endTime - startTime;
        rttMetric.add(rtt);
        console.log(`VU ${__VU}: Message ${messageCount} RTT: ${rtt}ms`);
      }
    });

    socket.on('close', () => {
      console.log(`VU ${__VU}: Connection closed after ${messageCount} messages`);
    });

    socket.setTimeout(() => {
      console.log(`VU ${__VU}: Test timeout after ${messageCount} messages`);
      socket.close(1000);
    }, 35000);
  });

  sleep(30);
}

function sendMessage(socket, messageId) {
  const payload = generateLargePayload(messageId);
  const message = JSON.stringify(payload);

  messageTimestamps.set(`${__VU}-${messageId}`, Date.now());
  socket.send(message);
  messagesSent.add(1);
  bytesSent.add(message.length);

  console.log(`VU ${__VU}: Sent message ${messageId}, size: ${(message.length/1024).toFixed(2)}KB`);
}

export function handleSummary(data) {
  const wsMetrics = data.metrics.websocket_rtt;
  const summary = {
    test_summary: {
      total_vus: options.vus,
      duration: options.duration,
      iterations: options.iterations
    },
    connections: {
      total: data.metrics.ws_sessions ? data.metrics.ws_sessions.count : 0,
      connecting_time: {
        avg: (data.metrics.ws_connecting ? data.metrics.ws_connecting.avg : 0).toFixed(2),
        min: (data.metrics.ws_connecting ? data.metrics.ws_connecting.min : 0).toFixed(2),
        max: (data.metrics.ws_connecting ? data.metrics.ws_connecting.max : 0).toFixed(2),
        p95: (data.metrics.ws_connecting ? data.metrics.ws_connecting.p(95) : 0).toFixed(2)
      }
    },
    messages: {
      sent: data.metrics.messages_sent ? data.metrics.messages_sent.count : 0,
      received: data.metrics.messages_received ? data.metrics.messages_received.count : 0,
      rate: {
        sent_per_second: ((data.metrics.messages_sent ? data.metrics.messages_sent.count : 0) / (data.state.testRunDurationMs / 1000)).toFixed(2),
        received_per_second: ((data.metrics.messages_received ? data.metrics.messages_received.count : 0) / (data.state.testRunDurationMs / 1000)).toFixed(2)
      }
    },
    rtt_ms: wsMetrics ? {
      avg: wsMetrics.avg.toFixed(2),
      min: wsMetrics.min.toFixed(2),
      max: wsMetrics.max.toFixed(2),
      med: wsMetrics.med.toFixed(2),
      p90: wsMetrics.p(90).toFixed(2),
      p95: wsMetrics.p(95).toFixed(2)
    } : null,
    test_run_details: {
      duration_ms: data.state.testRunDurationMs,
      data_received: data.metrics.data_received ? data.metrics.data_received.count : 0,
      data_sent: data.metrics.data_sent ? data.metrics.data_sent.count : 0
    }
  };

  return {
    stdout: JSON.stringify(summary, null, 2)
  };
}