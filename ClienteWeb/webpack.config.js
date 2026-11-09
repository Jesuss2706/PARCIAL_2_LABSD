// npx webpack  ->  empaqueta src/cliente_receptor_fragmentos.js (y sus imports de grpc-web,
// google-protobuf y @stomp/stompjs) en dist/main.js, que es el que carga index.html.
const path = require("path");

module.exports = {
  mode: "development",
  devtool: false,
  entry: "./src/cliente_receptor_fragmentos.js",
  output: {
    path: path.resolve(__dirname, "dist"),
    filename: "main.js",
  },
};
