const socket = new WebSocket('ws://127.0.0.1:9224/session');
const pending = new Map();
let id = 0;
socket.onmessage = ({ data }) => {
  const message = JSON.parse(data);
  if (message.id) {
    const task = pending.get(message.id);
    pending.delete(message.id);
    if (message.type === 'error') task.reject(message);
    else task.resolve(message.result);
  } else if (message.method === 'log.entryAdded') console.log(JSON.stringify(message.params));
};
await new Promise((resolve, reject) => {
  socket.onopen = resolve;
  socket.onerror = reject;
});
function call(method, params) {
  return new Promise((resolve, reject) => {
    const request = ++id;
    pending.set(request, { resolve, reject });
    socket.send(JSON.stringify({ id: request, method, params }));
  });
}
try {
  await call('session.end', {}).catch(() => {});
  await call('session.new', { capabilities: {} });
  await call('session.subscribe', { events: ['log.entryAdded'] });
  const { context } = await call('browsingContext.create', { type: 'tab' });
  await call('browsingContext.navigate', {
    context,
    url: 'http://127.0.0.1:5188/packages/ui/tests/browser/index.html',
    wait: 'complete',
  });
  const result = await call('script.evaluate', {
    target: { context },
    expression: `new Promise((resolve,reject) => { const deadline = Date.now()+20000; const poll = () => { const result=document.getElementById('result')?.textContent; if(result && result !== 'Starting') resolve(result); else if(Date.now()>deadline) reject(new Error('Timed out: '+document.body.innerText)); else setTimeout(poll,50); }; poll(); })`,
    awaitPromise: true,
  });
  console.log(JSON.stringify(result, null, 2));
  if (result.type !== 'success' || !JSON.parse(result.result.value).passed) process.exitCode = 1;
} finally {
  await call('session.end', {}).catch(() => {});
  socket.close();
}
