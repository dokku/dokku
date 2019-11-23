var express = require('express');

var app = express();

app.get('/', function(request, response) {
  response.send(process.env.CONFTEST);
});

app.get('/hello', function(request, response) {
  response.send(process.env.HELLO);
});

var port = process.env.PORT || 5000;
app.listen(port, function() {
  console.log("Listening on " + port);
});
