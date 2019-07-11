package cvl
import (
	"github.com/go-redis/redis"
)

//Redis server side script
func loadLuaScript() {
	luaScripts = make(map[string]*redis.Script)

	// Find entry which has given fieldName and value
	luaScripts["find_key"] = redis.NewScript(`
	local tableName=ARGV[1]
	local sep=ARGV[2]
	local fieldName=ARGV[3]
	local fieldValue=ARGV[4]

	-- Check if field value is part of key
	local entries=redis.call('KEYS', tableName..sep.."*"..fieldValue.."|*")

	if (entries[1] ~= nil)
	then
		return entries[1]
	else

	-- Search through all keys for fieldName and fieldValue
		local entries=redis.call('KEYS', tableName..sep.."*")

		local idx = 1
		while(entries[idx] ~= nil)
		do
			local val = redis.call("HGET", entries[idx], fieldName)
			if (val == fieldValue) 
			then
				-- Return the key
				return entries[idx]
			end

			idx = idx + 1
		end
	end

	-- Could not find the key
	return ""
	`)

}
