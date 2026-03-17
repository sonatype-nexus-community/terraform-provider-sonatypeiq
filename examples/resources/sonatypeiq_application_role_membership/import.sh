# Application Role Memberships can be imported.

# Example
terraform import sonatypeiq_application_role_membership.rm1 <APPLICATION-ID>,<ROLE-ID>,<group|user>,<group-name|user-name>
terraform import sonatypeiq_application_role_membership.rm1 11614d18e28b4cbe9dae03d1cf00d663,11614d18e28b4cbe9dae03d1cf00d663,group,saml-admins
