module Phoenix
  module Cache

    # 
    # A mechanism for preventing concurrent modification of a single memcache key by multiple 
    # processes.
    #
    # To use locked_X in your class, you must do the following:
    #
    # 1. require 'phoenix/cache/locking'
    # 2. extend Phoenix::Cache::Locking
    # 3. implement 'self.cache' on your class 
    #   - frequently just:
    #     def self.cache
    #       MEMCACHE
    #     end
    # 4. If you need to locked_modify in class methods, also implement locked_modify per:
    #    def locked_modify(*args, &block)
    #      self.class.locked_modify(*args, &block)
    #     end
    #

    module Locking
      def locked_read(key, opts)
        lock(key) do
          o = cache.get(key)
          if o.nil? && opts[:fail_on_nil]
            raise "Unable to find: #{key}"
          else
            yield o
          end
        end
      end

      def locked_modify(key, timeout_or_options={})
        if Hash === timeout_or_options
          default = timeout_or_options[:default] || {}
          timeout = timeout_or_options[:timeout] || 5
        else
          default = {}
          timeout = timeout_or_options
        end
        lock(key, timeout) do
          o = cache.get(key) || default
          yield o
          cache.set(key, o)
        end
      end

      def locked_set(key, o, timeout=5)
        lock(key,timeout) do
          yield o
          cache.set(key, o)
        end
      end

      def lock(key, timeout=5)
        t0 = Time.now
        lock_key = "lock;#{key}"
        begin
          if cache.add(lock_key, self.object_id, timeout)
            yield
          else
            if (Time.now - t0) < timeout
              sleep 0.3
              raise 'retry'
            else
              raise "Unable to lock: #{key}"
            end
          end
        rescue Dalli::NetworkError
          # silently ignore memcache connection errors
          nil
        rescue
          if $!.message == 'retry'
            retry
          else
            raise
          end
        ensure
          begin
            cache.delete(lock_key)
          rescue Dalli::NetworkError
            # silently ignore memcache connection errors
            nil
          end
        end
      end
    end
  end
end
