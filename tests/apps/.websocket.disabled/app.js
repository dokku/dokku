var express = require('express')
  , exphbs  = require('express3-handlebars')
  , http = require('http')
  , connect = require('connect')
  , io = require('socket.io');



var app = express();

app.engine('handlebars',exphbs({extname: ".html", layoutsDir:"views/", defaultLayout: 'layout'}));
app.set('view engine', 'handlebars');
/* NOTE: We'll need to refer to the sessionStore container later. To
 *       accomplish this, we'll create our own and pass it to Express
 *       rather than letting it create its own. */
var sessionStore = new connect.session.MemoryStore();
/* NOTE: We'll need the site secret later too, so let's factor it out.
 *       The security implications of this are left to the reader. */
var SITE_SECRET = 'I am not wearing any pants';


// app.configure(function(){
//     app.set('view engine', 'handlebars');
//     // app.set("view options", { layout: false }) 
// });

app.set('views', __dirname + '/views');
// app.register('.html', require('handlebars'));
/*
 * ... skipping some of your app settings ...
 */
app.use(express.bodyParser());
/* NOTE: Pass the cookieParser the site secret. It used to be that
 *       express.session() got the secret, but in Express 3 that's
 *       no longer the case. */
app.use(express.cookieParser(SITE_SECRET));
/* NOTE: We'll need to know the key used to store the session, so
 *       we explicitly define what it should be. Also, we pass in
 *       our sessionStore container here. */
app.use(express.session({
    key: 'express.sid'
  , store: sessionStore
}));

app.use(express.static('public'));
/*
 * ... skipping the rest of your app settings ...
 */

app.get('/', function(req, res){
  res.render('index', {socket_url:"ciaourl"});
});

var server = http.createServer(app);

var app_port = process.env.PORT || 3000;
server.listen(app_port, function(){
  console.log("Express server listening on port "+ app_port);
});
 
/**
 * Socket.io
 */
var sio = io.listen(server);
 

sio.sockets.on('connection', function (socket) {

  // temperature
  setInterval(function () {
    var val = (new Date).getTime();
    socket.emit('new value', { val: "I'm a websocket message: "+val });
  }, 2000);

});