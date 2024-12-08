import { sleep } from 'k6';
import { Trend, Counter } from 'k6/metrics';
import ws from 'k6/ws';

const rttMetric = new Trend('websocket_rtt');
const messagesReceived = new Counter('messages_received');
const messagesSent = new Counter('messages_sent');

const messageTimestamps = new Map();

export const options = {
  vus: 50,
  duration: '40s',
  iterations: 50,
  teardownTimeout: '0s',
  noConnectionReuse: true,
  gracefulStop: '0s',
  thresholds: {
    'websocket_rtt': ['p(95)<1000'],  // 95% of RTT should be under 1000ms
    'messages_received': ['count>0'],
    'messages_sent': ['count>0']
  }
};

export default function () {
  const userId = `${__VU}`;
  const url = `ws://localhost/ws?userId=${userId}`;
  let messageCount = 0;

  const res = ws.connect(url, null, function (socket) {
    socket.on('open', () => {
      console.log(`VU ${__VU}: Connected`);
      sendMessage(socket, 1);
    });

    socket.on('message', (data) => {
      const endTime = Date.now();
      messagesReceived.add(1);

      const key = `${userId}-${messageCount + 1}`;
      const startTime = messageTimestamps.get(key);
      if (startTime) {
        const rtt = endTime - startTime;
        rttMetric.add(rtt);
        console.log(`VU ${__VU}: Message ${messageCount + 1} RTT: ${rtt}ms`);
      }

      messageCount++;

      if (messageCount < 5) {
        sleep(5);
        sendMessage(socket, messageCount + 1);
      } else {
        socket.close(1000);
      }
    });

    socket.on('close', () => {
      console.log(`VU ${__VU}: Connection closed after ${messageCount} messages`);
    });

    socket.setTimeout(() => {
      if (messageCount < 5) {
        console.log(`VU ${__VU}: Test timeout after ${messageCount} messages`);
        socket.close(1000);
      }
    }, 35000);
  });

  sleep(30);
}

function sendMessage(socket, messageId) {
  const message = JSON.stringify({
    messageId: messageId.toString(),
    userId: `${__VU}`,
    content: `Message ${messageId} from VU ${__VU}`
  });

  messageTimestamps.set(`${__VU}-${messageId}`, Date.now());
  socket.send(message);
  messagesSent.add(1);
  console.log(`VU ${__VU}: Sent message ${messageId}`);
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