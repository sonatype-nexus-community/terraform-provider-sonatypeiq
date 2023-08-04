# Get the Organization by ID
data "sonatypeiq_organization" "sandbox" {
  id = "211ccadc89974ecd83cf91495e3b29f5"
}

# Get the Organization by Name
data "sonatypeiq_organization" "sandbox" {
  name = "Sandbox Organization"
}