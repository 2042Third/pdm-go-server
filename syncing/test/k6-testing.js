import ws from 'k6/ws';
import { check } from 'k6';
import { Trend } from 'k6/metrics';

// Create a trend metric for RTT measurements
const rttTrend = new Trend('message_rtt');

export const options = {
  vus: 1,
  iterations: 1
};

export default function () {
  const url = `ws://localhost/ws?userId=${__VU}`;
  const params = { headers: { 'userId': `${__VU}` } };

  const transactions = [
    [
      { msg: 'hello', code: 13 },
      { msg: 'test', code: 42 },
      { msg: 'world', code: 99 },
    ],
    [
      { msg: 'foo', code: 7 },
      { msg: 'bar', code: 21 },
      { msg: 'baz', code: 63 },
    ],
    [
      { msg: 'ping', code: 1 },
      { msg: 'pong', code: 2 },
      { msg: 'boom', code: 3 },
    ],
  ];

  const expectedResponses = transactions.map((transaction) =>
    transaction.map(({ msg, code }) => `${msg}: ${code}`)
  );

  const messageSendTimes = new Map();
  let totalMessages = 0;
  let receivedMessages = 0;

  // Calculate total number of messages
  transactions.forEach(transaction => {
    totalMessages += transaction.length;
  });

  return new Promise((resolve) => {
    const socket = ws.connect(url, params, function (socket) {
      socket.on('open', function () {
        // console.log('Connected to WebSocket');
        transactions.forEach((transaction, transactionIndex) => {
          transaction.forEach((message, messageIndex) => {
            const messageKey = `${transactionIndex}-${messageIndex}`;
            messageSendTimes.set(messageKey, Date.now());
            socket.send(JSON.stringify(message));
          });
        });
      });

      let transactionIndex = 0;
      let messageIndex = 0;

      socket.on('message', function (data) {
        // Calculate RTT
        const messageKey = `${transactionIndex}-${messageIndex}`;
        const sendTime = messageSendTimes.get(messageKey);
        const rtt = Date.now() - sendTime;
        rttTrend.add(rtt);

        // console.log(`Message RTT: ${rtt}ms`);

        const expectedResponse = expectedResponses[transactionIndex][messageIndex];
        const valid = check(data, {
          [`is response for transaction ${transactionIndex + 1}, message ${messageIndex + 1} valid`]: (r) =>
            r === expectedResponse,
        });

        if (!valid) {
          console.error(
            `Mismatch: expected "${expectedResponse}" but received "${data}" for transaction ${transactionIndex + 1}, message ${messageIndex + 1}`
          );
        }

        messageIndex++;
        receivedMessages++;

        if (messageIndex >= transactions[transactionIndex].length) {
          messageIndex = 0;
          transactionIndex++;
        }

        // Close connection after receiving all messages
        if (receivedMessages === totalMessages) {
          // console.log('All messages processed, closing connection');
          socket.close();
          resolve(); // Resolve the promise to end the test
        }
      });

      socket.on('close', function () {
        // console.log('Connection closed');
        messageSendTimes.clear();
      });

      socket.on('error', function (e) {
        console.error('Error:', e.error());
        socket.close();
        resolve();
      });
    });
  });
}

export function handleSummary(data) {
  if (!data.metrics.message_rtt) {
    return {
      stdout: JSON.stringify({
        error: 'No RTT metrics collected'
      }, null, 2)
    };
  }

  return {
    stdout: JSON.stringify({
      rtt_stats: {
        min_ms: data.metrics.message_rtt.min,
        max_ms: data.metrics.message_rtt.max,
        avg_ms: data.metrics.message_rtt.avg,
        median_ms: data.metrics.message_rtt.med,
        p90_ms: data.metrics.message_rtt.values['p(90)'],
        p95_ms: data.metrics.message_rtt.values['p(95)'],
        total_messages: data.metrics.message_rtt.count
      }
    }, null, 2)
  };
}