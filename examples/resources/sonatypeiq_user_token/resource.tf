# Generate a User Token for the current User
resource "sonatypeiq_user_token" "current_user_token" {
  generated_at = "1" # << This is a field you can use to manage when the User Token was generated. You might use a date or serial number.
}

# Update the Current User's User Token
resource "sonatypeiq_user_token" "current_user_token" {
  generated_at = "2" # << Because this has changed, User Token will be re-generated
}