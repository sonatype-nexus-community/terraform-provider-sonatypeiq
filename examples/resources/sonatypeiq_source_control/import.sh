# Applications can be imported using the sonatype type (application|organization) and the sonatype id.
# This can be obtained by searching using the PublicID in the WebUI or by calling the rest API

# Example for an application
terraform import sonatypeiq_source_control.application application:4bb67dcfc86344e3a483832f8c496419

# Example for an organization
terraform import sonatypeiq_source_control.organization organization:4bb67dcfc86344e3a483832f8c496419