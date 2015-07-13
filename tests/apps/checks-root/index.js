var express = require('express');
var app = express();

app.get('/', function(req, res) {
  res.sendStatus(404);
});

var port = process.env.PORT || 5000;
app.listen(port, function() {
  console.log('Listening on port ' + port);
});
