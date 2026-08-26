package types

//
// GENERATED!
//
// ATTENTION: This file is generated from the template:
//
//   share/templates/principal.tmpl
//
// If you want to make any changes to this file, modify the template.
//

// Principal represents the Principal schema.
type Principal struct {
	PtcClientID  string   `json:"ptc_client_id"`
	PtcTenantID  string   `json:"ptc_tenant_id"`
	PtcUserID    string   `json:"ptc_user_id"`
	PtcUserRoles []string `json:"ptc_user_roles"`
	PtcUsername  string   `json:"ptc_username"`
}

// Generate token claims as well
const (
	PtcClientIDClaim  string = "ptc_client_id"
	PtcTenantIDClaim  string = "ptc_tenant_id"
	PtcUserIDClaim    string = "ptc_user_id"
	PtcUserRolesClaim string = "ptc_user_roles"
	PtcUsernameClaim  string = "ptc_username"
)
