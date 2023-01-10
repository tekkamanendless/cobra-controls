package wire

const (
	FunctionClearUpload         = 0x1093 // TODO: I'm not exactly sure what this is.
	FunctionUnknown1098         = 0x1098 // TODO: Maybe ping/pong?
	FunctionGetOperationStatus  = 0x1081
	FunctionGetBasicInfo        = 0x1082
	FunctionSetTime             = 0x108b
	FunctionGetRecord           = 0x108d
	FunctionDeleteRecord        = 0x108e
	FunctionGetUpload           = 0x1095
	FunctionUpdateControlPeriod = 0x1097
	FunctionTailPlusPermissions = 0x109b
	FunctionOpenDoor            = 0x109d
	FunctionGetSetting          = 0x10f1
	FunctionUpdateSetting       = 0x10f4
	FunctionGetNetworkInfo      = 0x1101
	FunctionUpdatePermissions   = 0x1107
)
