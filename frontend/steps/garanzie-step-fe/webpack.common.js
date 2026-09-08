const path = require('path')
const webpack = require('webpack')
const { VueLoaderPlugin } = require('vue-loader')
const nodeExternals = require('webpack-node-externals')

// Due configurazioni webpack, due target diversi.
//
// È la parte che spiazza di più arrivando da una SPA normale: qui non si
// costruisce UN bundle, se ne costruiscono TRE, per due ambienti diversi.
//
//   server.js         gira in Node      -> il server Hypernova
//   web_client.js     gira nel browser  -> hydration della vista desktop
//   mobile_client.js  gira nel browser  -> hydration della vista mobile
//
// Poste usa esattamente questo schema (webpack 4, due config esportate).

const STEP_ID = process.env.STEP_ID || 'garanzie'
const ID_VIEW_WEB = process.env.ID_VIEW_WEB || 'garanzie-step_web'
const ID_VIEW_MOBILE = process.env.ID_VIEW_MOBILE || 'garanzie-step_mobile'

// Le variabili d'ambiente vanno "iniettate" nel bundle: nel browser
// process.env non esiste, quindi webpack sostituisce le occorrenze con il
// valore letterale in fase di compilazione.
const definePlugin = new webpack.DefinePlugin({
  'process.env.STEP_ID': JSON.stringify(STEP_ID),
  'process.env.ID_VIEW_WEB': JSON.stringify(ID_VIEW_WEB),
  'process.env.ID_VIEW_MOBILE': JSON.stringify(ID_VIEW_MOBILE),
})

const regoleComuni = [
  { test: /\.vue$/, loader: 'vue-loader' },
  {
    test: /\.js$/,
    exclude: /node_modules/,
    use: { loader: 'babel-loader' },
  },
]

const server = {
  target: 'node',
  entry: path.join(__dirname, 'src/index.js'),
  output: {
    path: path.join(__dirname, 'dist'),
    filename: 'server.js',
  },
  // Il bundle server NON deve contenere node_modules: sono già su disco.
  // Senza questo, express e hypernova finirebbero dentro server.js e
  // qualcosa si romperebbe (i moduli nativi non sono impacchettabili).
  externals: [nodeExternals()],
  module: {
    rules: [
      ...regoleComuni,
      // Lato server il CSS non va emesso in un file: si scarta.
      // Gli stili arrivano al browser con il bundle client.
      { test: /\.css$/, use: 'null-loader' },
    ],
  },
  plugins: [new VueLoaderPlugin(), definePlugin],
}

const client = {
  target: 'web',
  entry: {
    web: path.join(__dirname, 'src/web/client.js'),
    mobile: path.join(__dirname, 'src/mobile/client.js'),
  },
  output: {
    path: path.join(__dirname, 'dist'),
    filename: '[name]_client.js',
  },
  // Nel browser questi moduli Node non esistono: si dice a webpack di
  // sostituirli con oggetti vuoti invece di fallire.
  node: { fs: 'empty', net: 'empty', module: 'empty' },
  module: {
    rules: [
      ...regoleComuni,
      { test: /\.css$/, use: ['vue-style-loader', 'css-loader'] },
    ],
  },
  plugins: [new VueLoaderPlugin(), definePlugin],
}

module.exports = { server, client }
