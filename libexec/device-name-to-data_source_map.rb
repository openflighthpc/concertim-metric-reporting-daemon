#!/usr/bin/env ruby

require 'dalli'
require 'dalli/client'
require 'json'

address = ARGV.first || "localhost:11211"
mc = Dalli::Client.new(address, { serializer: Marshal })

results = {"hostnameMap" => {}, "memcacheMap" => {}}

dsm = mc.get("hacor:data_source_map")
map_to_grid = "unspecified"
map_to_cluster = "unspecified"

cluster = dsm[map_to_grid][map_to_cluster]
cluster.each do |map_to_host, host_id|
  host_data = mc.get(host_id)
  host_name = host_data[:name]
  results["hostnameMap"][host_name] = map_to_host
end

results["memcacheMap"] = dsm

puts results.to_json
