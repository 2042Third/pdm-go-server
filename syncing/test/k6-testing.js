import ws from 'k6/ws';
import { check } from 'k6';

export default function () {
  const url = 'ws://localhost:8082/ws'; // Replace with your URL
  const params = { headers: { 'X-My-Header': 'value' } };

  const response = ws.connect(url, params, function (socket) {
    socket.on('open', function () {
      // console.log('Connected');
      socket.send(JSON.stringify({msg: 'hello', code: 13}));
    });

    socket.on('message', function (data) {
      // console.log('Received:', data);
      check(data, {
        'is response valid': (r) => r === 'hello: 13',
      });
      socket.close();
    });

    socket.on('close', function () {
      // console.log('Connection closed');
    });

    socket.on('error', function (e) {
      console.log('Error:', e.error());
    });
  });

}
