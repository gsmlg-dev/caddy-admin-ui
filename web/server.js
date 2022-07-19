const express = require('express');
const next = require('next');
const { createProxyMiddleware } = require('http-proxy-middleware');

const port = process.env.PORT || 3000;
const dev = process.env.NODE_ENV !== 'production';
const app = next({ dev });
const handle = app.getRequestHandler();

const apiPaths = {};

const apis = ['load', 'stop', 'config', 'pki', 'reverse_proxy', 'adapt', 'metrics'];

const simpleRequestLogger = (proxyServer, options) => {
  proxyServer.on('proxyReq', (proxyReq, req, res) => {
    console.log(`[HPM] [${req.method}] ${req.url}`); // outputs: [HPM] GET /users
  });
};

apis.forEach((name) => {
  apiPaths[`/${name}`] = {
    target: `http://127.0.0.1:2019`,
    changeOrigin: true,
    plugins: [simpleRequestLogger],
  };
});

const isDevelopment = process.env.NODE_ENV !== 'production';

app
  .prepare()
  .then(() => {
    const server = express();

    if (isDevelopment) {
      Object.keys(apiPaths).forEach((key) => {
        server.use(key, createProxyMiddleware(apiPaths[key]));
      });
    }

    server.all('*', (req, res) => {
      return handle(req, res);
    });

    server.listen(port, (err) => {
      if (err) throw err;
      console.log(`> Ready on http://localhost:${port}`);
    });
  })
  .catch((err) => {
    console.log('Error:::::', err);
  });
