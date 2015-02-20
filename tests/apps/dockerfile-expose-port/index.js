var express = require('express')
var app = express();

// Port is the same as exposed port in the Dockerfile
app.set('port', 3000)
app.use(express.static(__dirname + '/public'))

app.get('/', function(request, response) {
  response.send('Hello World!')
})

app.listen(app.get('port'), function() {
  console.log("Node app is running at localhost:" + app.get('port'))
})
