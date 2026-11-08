# Look up a global role, such as System Administrator
data "sonatypeiq_role" "system_administrator" {
  name = "System Administrator"
}

# Grant it to a group (for example one provided by your SSO) ...
resource "sonatypeiq_global_role_membership" "admin_group" {
  role_id    = data.sonatypeiq_role.system_administrator.id
  group_name = "My Admin Group"
}

# ... or to a user
resource "sonatypeiq_global_role_membership" "admin_user" {
  role_id   = data.sonatypeiq_role.system_administrator.id
  user_name = "jane.doe"
}
