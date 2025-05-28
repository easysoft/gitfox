const { merge } = require('webpack-merge');
const HTMLWebpackPlugin = require('html-webpack-plugin')
const devConfig = require('./webpack.dev');

const config = merge(devConfig, {
  devServer: {
    proxy: {
      '/api': {
        changeOrigin: false
      }
    },
  },
});

config.plugins.forEach((plugin) => {
  if (plugin instanceof HTMLWebpackPlugin) {
    plugin.userOptions.publicPath = '/';
    plugin.userOptions.templateParameters = {
      __DEV__: true,
    }
  }
});

module.exports = config;
