require 'rubygems'
require 'rack'
require 'json'

SERVER_ID=(0...10).map{ ('a'..'z').to_a[rand(26)] }.join

class HelloWorld
  def call(env)
    if env['REQUEST_PATH'] == '/env'
      [200, {"Content-Type" => "application/json", "Server-Id" => SERVER_ID}, ENV.to_hash.to_json]
    elsif env['REQUEST_PATH'] =~ %r{/env/(.+)}
      [200, {"Content-Type" => "text/plain", "Server-Id" => SERVER_ID}, ENV[$1] || 'n/a']
    else
      [200, {"Content-Type" => "text/html", "Server-Id" => SERVER_ID}, "ruby/rack"]
    end
  end
end

run HelloWorld.new