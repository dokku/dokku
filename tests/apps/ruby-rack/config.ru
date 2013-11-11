require 'rubygems'
require 'rack'
require 'json'

class HelloWorld
  def call(env)
    if env['REQUEST_PATH'] == '/env'
      [200, {"Content-Type" => "application/json"}, ENV.to_hash.to_json]
    elsif env['REQUEST_PATH'] =~ %r{/env/(.+)}
      [200, {"Content-Type" => "text/plain"}, ENV[$1] || 'n/a']
    else
      [200, {"Content-Type" => "text/html"}, "ruby/rack"]
    end
  end
end

run HelloWorld.new