package errorx

var (
	Ok              = add(0, "ok")
	ErrRequest      = add(100400, "请求参数错误")
	ErrNotFind      = add(100404, "没有找到")
	ErrForbidden    = add(100403, "请求被拒绝")
	ErrNoPermission = add(100405, "无权限")
	ErrAuthFail     = add(100406, "认证失败")
	ErrServer       = add(100500, "服务器错误")
	ErrUpdate       = add(100502, "存储信息失败")

	// config
	ConfigSaveError         = add(101202, "写入配置文件失败")
	ConfigRedisConnectError = add(101203, "Redis连接失败")
	ConfigMySQLConnectError = add(101204, "MySQL连接失败")
	ConfigMySQLInstallError = add(101205, "MySQL初始化数据失败")
	ConfigGoVersionError    = add(101206, "GoVersion不满足要求")
	ConfigSecretKeyError    = add(101207, "配置秘钥错误")

	// db common error   101301 -- 101399
	DbFindError   = add(101301, "获取数据失败")
	DbExistsError = add(101302, "数据已存在")
	DbAnchorError = add(101303, "达人信息不完整")
	DbInfoError   = add(101304, "信息不完整")

	// 业务 错误码 200100-200199
	ErrorApplyNum  = add(200101, "申请数量已达上限")
	ErrorSpuStatus = add(200102, "商品已下架")
	ErrorCartNum   = add(200104, "数量不正确")

	// 订单模块 错误码
	ErrorVerClues      = add(200105, "当前学服系统，没有您的用户信息。请联系学服老师！")
	ErrorVerSid        = add(200106, "当前支付链接已失效。请联系学服老师！")
	ErrorCreateDmOrder = add(200107, "订单支付失败。请联系学服老师！")

	// 业务 分享链接 200200-200299
	ErrOrderCreated = add(200200, "订单已创建")
)
