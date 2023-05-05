#!/usr/bin/env ruby

require 'dalli'
require 'dalli/client'
require 'json'
require 'time'
require 'active_support/core_ext/hash/keys'
require_relative 'phoenix-cache-locking'

class Run
  include Phoenix::Cache::Locking

  def initialize(address, results)
    puts "connecting to #{address}..."
    @mc = Dalli::Client.new(address, { serializer: Marshal })
    @results = results
  end

  def call
    hosts = @results["hosts"] || []
    puts "updating #{hosts.length} hosts..."
    hosts.each do |host|
      locked_modify(host["memcache_key"]) do |o|
        o[:metrics] = host["metrics"].deep_symbolize_keys
        unless host["mtime"].nil?
          o[:mtime] = host["mtime"]
        end
      end
    end

    hosts_by_metric = @results["hosts_by_metric"]
    puts "recording #{hosts_by_metric.length} metric -> device maps..."
    hosts_by_metric.each do |metric_name, hosts|
      cache.set("meryl:metric:#{metric_name}", hosts)
    end

    uniq = @results["unique_metrics"]
    puts "recording #{uniq.length} unique metrics..."
    cache.set("meryl:unique_metrics", uniq)
  end

  def cache
    @mc
  end
end

puts "reading results from stdin..."
results = JSON.parse(STDIN.read)
address = ARGV.first || "localhost:11211"
Run.new(address, results).call

exit 0
