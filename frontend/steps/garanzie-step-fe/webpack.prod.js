const merge = require('webpack-merge')
const { server, client } = require('./webpack.common')

// webpack accetta un ARRAY di configurazioni: le costruisce tutte in una
// passata. È così che da un comando solo escono server.js, web_client.js
// e mobile_client.js.
module.exports = [
  merge(server, { mode: 'production' }),
  merge(client, { mode: 'production' }),
]
