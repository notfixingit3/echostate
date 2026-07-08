package audit

// Action names stored in audit_events.action.
const (
	ActionLogin                = "auth.login"
	ActionLogout               = "auth.logout"
	ActionPasskeyRegister      = "auth.passkey_register"
	ActionDeviceCodeIssued     = "auth.device_code_issued"
	ActionEnrollmentCodeIssued = "auth.enrollment_code_issued"
	ActionUserCreate           = "user.create"
	ActionCredentialDelete     = "user.credential_delete"
	ActionTargetCreate         = "target.create"
	ActionTargetDelete         = "target.delete"
	ActionTargetTagsUpdate     = "target.tags_update"
	ActionSnapshotDelete       = "snapshot.delete"
	ActionReportDelete         = "report.delete"
	ActionScanCancel           = "scan.cancel"
	ActionSettingsUpdate       = "settings.update"
	ActionDataExport           = "data.export"
	ActionDataImport           = "data.import"
	ActionWebhookCreate        = "webhook.create"
	ActionWebhookUpdate        = "webhook.update"
	ActionWebhookDelete        = "webhook.delete"
)
