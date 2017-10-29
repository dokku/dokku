var express = require('express')
var app = express();

app.set('port', (process.env.PORT || 5000))
app.use(express.static(__dirname + '/public'))

app.get('/', function(request, response) {
  response.send('Hello World!')
})

app.get('/env/:key', function(request, response) {
  response.send(process.env[request.params["key"]])
})

app.listen(app.get('port'), function() {
  console.log("Node app is running at localhost:" + app.get('port'))
})

app.listen(3003, function() {
  console.log("Node app is running at localhost:" + 3003)
})
