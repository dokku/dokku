This is a sample application to test heroku buildpack multi for integrating with [herokuish](https://github.com/gliderlabs/herokuish)

It uses the python, nodejs, and ruby buildpacks to create a simple flask app. The ruby buildpack provides kramdown for Markdown-to-HTML conversion. The node buildpack manages Bootstrap via npm and compiles SCSS with dart-sass. The python buildpack runs the Flask application.
