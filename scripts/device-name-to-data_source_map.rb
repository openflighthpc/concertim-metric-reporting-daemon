#!/usr/bin/env ruby1.9

require 'memcache'
require 'json'

mc = MemCache.new('127.0.0.1:11211')

results = {}

dsm = mc.get("hacor:data_source_map")
map_to_grid = "unspecified"
map_to_cluster = "unspecified"

cluster = dsm[map_to_grid][map_to_cluster]
cluster.each do |map_to_host, host_id|
  host_data = mc.get(host_id)
  host_name = host_data[:name]
  results[host_name] = map_to_host
end

puts results.to_json
