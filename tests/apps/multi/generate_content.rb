require 'kramdown'

input_path = File.join(__dir__, 'content.md')
output_path = File.join(__dir__, 'templates', 'content.html')

markdown = File.read(input_path)
html = Kramdown::Document.new(markdown).to_html

File.write(output_path, html)
puts "Generated #{output_path} from #{input_path}"
