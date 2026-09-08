const merge = require('webpack-merge')
const { server, client } = require('./webpack.common')

module.exports = [
  merge(server, { mode: 'development', devtool: 'source-map' }),
  merge(client, { mode: 'development', devtool: 'source-map' }),
]
