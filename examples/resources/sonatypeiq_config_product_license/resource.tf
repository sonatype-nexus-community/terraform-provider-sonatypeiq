resource "sonatypeiq_config_product_license" "lic" {
  license_data = filebase64("/path/to/sonatype-license.lic")
}