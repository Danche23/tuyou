package sms

// Scene 验证码场景
type Scene string

const (
	SceneRegister      Scene = "register"
	SceneLogin         Scene = "login"
	SceneResetPwd      Scene = "reset-pwd"
	SceneChangePhone   Scene = "change-phone"
	SceneCancelAccount Scene = "cancel-account" // 注销账号
)
