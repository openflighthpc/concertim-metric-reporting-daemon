#!/usr/bin/env ruby

require 'dalli'
require 'dalli/client'
require 'json'

address = ARGV.first || "localhost:11211"
mc = Dalli::Client.new(address, { serializer: Marshal })

results = {"deviceIdToGangliaHostName" => {}, "dsmToMemcacheKey" => {}}

dsm = mc.get("hacor:data_source_map")
map_to_grid = "unspecified"
map_to_cluster = "unspecified"

cluster = dsm[map_to_grid][map_to_cluster]
cluster.each do |map_to_host, host_memcache_key|
  host_data = mc.get(host_memcache_key)
  host_id = host_data[:id]
  results["deviceIdToGangliaHostName"][host_id] = map_to_host
end

results["dsmToMemcacheKey"] = dsm

puts results.to_json
