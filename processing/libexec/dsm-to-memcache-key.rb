#!/usr/bin/env ruby

require 'dalli'
require 'dalli/client'
require 'json'

address = "localhost:11211"
STDERR.puts "connecting to #{address}..."
mc = Dalli::Client.new(address, { serializer: Marshal })

dsm = mc.get("hacor:data_source_map")
puts dsm.to_json
