#!/usr/bin/env ruby

require 'json'
require 'time'

address = ARGV.first || "localhost:11211"
puts "fake connecting to #{address}..."

puts "reading results from stdin..."
results = JSON.parse(STDIN.read)

hosts = results["hosts"] || []
puts "fake updating #{hosts.length} hosts..."
hosts.each do |host|
  metrics = host["metrics"]
  # puts "  fake updating host #{host["name"]}:#{host["memcache_key"]}..."
  # puts "    setting #{metrics.length} metrics #{metrics.keys}"
  # unless host["mtime"].nil?
  #   puts "    setting mtime #{host["mtime"]}"
  # end
end

hosts_by_metric = results["hosts_by_metric"]
puts "fake recording #{hosts_by_metric.length} metric -> device maps..."

uniq = results["unique_metrics"]
puts "fake recording #{uniq.length} unique metrics..."

exit 0
