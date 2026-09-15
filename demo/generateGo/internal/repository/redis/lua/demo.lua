-- 测试 repository redis lua 脚本
-- 返回json 
local cjson = cjson
return cjson.encode({
  code=0,
  balance=100
})