var express = require('express');

var app = express();

app.get('/', function(request, response) {
  response.send('nodejs/express');
});

app.get('/healthcheck', function(request, response) {
  response.send(process.env.HEALTHCHECK_ENDPOINT);
});

var port = process.env.PORT || 5000;
app.listen(port, function() {
  console.log("Listening on " + port);
});
