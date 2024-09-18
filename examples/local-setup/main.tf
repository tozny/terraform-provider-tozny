# Configure the Terraform runtime
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      # Pull signed provider binaries from
      # the Terraform hosted registry namespace
      # for Tozny registry.terraform.io/tozny
      source = "tozny/tozny"
      # Pin Tozny provider version
      version = ">=0.27.0"
    }
  }
}

# Local values to substitute below
locals {
    api_url = "http://platform.local.tozny.com:8000"
    account_username = "admin-local@tozny.com"
    account_password = "Test@12345"
    realm_name =  "azuretest"
}

# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint     = local.api_url
  account_username = local.account_username
  account_password = local.account_password
}

# Generate credentials(Eg: Key Pairs) for the account
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  persist_credentials_to           = "terraform"
}

# Generate Registration token to pass to the realm
resource "tozny_client_registration_token" "realm_registration_token" {
  client_credentials_config         = tozny_account.autogenerated_tozny_account.config
  name                              = "${local.realm_name}DefaultClientRegistrationToken"
  allowed_registration_client_types = ["general", "identity", "broker"]
  enabled                           = true
  one_time_use                      = false
}

# Create a realm
resource "tozny_realm" "my_organizations_realm" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_client_registration_token.realm_registration_token
  ]
  client_credentials_config   = tozny_account.autogenerated_tozny_account.config
  realm_name                  = local.realm_name
  sovereign_name              = "Administrator"
  default_registration_token  = tozny_client_registration_token.realm_registration_token.token
  mpc_enabled                 = true
  tozid_federation_enabled    = true
  secrets_enabled             = true
  forgot_password_custom_link = ""
  forgot_password_custom_text = ""
}

# Create broker for realm
resource "tozny_realm_broker_identity" "broker_identity" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_client_registration_token.realm_registration_token
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  name                      = "broker${tozny_realm.my_organizations_realm.realm_name}"
  persist_credentials_to    = "terraform"
  # broker_identity_credentials_save_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
}

# Delegate authority to the broker
resource "tozny_realm_broker_delegation" "allow_tozny_hosted_brokering_policy" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_account.autogenerated_tozny_account,
    tozny_realm_broker_identity.broker_identity,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  # realm_broker_identity_credentials_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
  realm_broker_identity_credentials = tozny_realm_broker_identity.broker_identity.credentials
  use_tozny_hosted_broker           = true
}

# Register local user one
resource "tozny_realm_identity" "add_user_one" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  username                  = "user1"
  password                  = "Test@12345"
  email                     = "user1@tozny.com"
  broker_target_url         = "http://localhost:8081/${tozny_realm.my_organizations_realm.realm_name}/recover"
}

# Register local user two
resource "tozny_realm_identity" "add_user_two" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  username                  = "user2"
  password                  = "Test@12345"
  email                     = "user2@tozny.com"
  broker_target_url         = "http://localhost:8081/${tozny_realm.my_organizations_realm.realm_name}/recover"
}

# Create a default group
resource "tozny_realm_group" "default_group_1" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name                      = "Test Group"
  realm_name                = tozny_realm.my_organizations_realm.realm_name
}

# Add user one to the group
resource "tozny_realm_identity_group_membership" "eg_membership" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  group_ids                 = [tozny_realm_group.default_group_1.group_id]
  identity_id               = tozny_realm_identity.add_user_one.id
}

# Add Azure IdP
resource "tozny_identity_provider" "azure_identity_provider" {
  realm_name   = local.realm_name
  display_name = "Azure AD"
  alias        = "azure"
  enabled      = true
  config {
    authorization_url  = "https://login.microsoftonline.com/94d7761e-37d3-4028-9a69-6ec6587cfac0/oauth2/v2.0/authorize"
    token_url          = "https://login.microsoftonline.com/94d7761e-37d3-4028-9a69-6ec6587cfac0/oauth2/v2.0/token"
    client_auth_method = "client_secret_post"
    client_id          = "" # Sensitive
    client_secret      = "" # Sensitive
    default_scope      = "email profile openid"
  }
}

# Add LDAP
resource "tozny_realm_provider" "ldap_identity_provider" {
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  provider_type             = "ldap"
  name                      = "LDAP Identity Provider"
  active                    = true
  import_identities         = true
  priority                  = 0
  connection_settings {
    type                    = "other"
    identity_name_attribute = "uid"
    edit_mode               = "READ_ONLY"
    rdn_attribute           = "uid"
    uuid_attribute          = "entryUUID"
    identity_object_classes = ["person", "organizationalPerson", "user"]
    connection_url          = "ldaps://ldap.jumpcloud.com:636"
    identity_dn             = "ou=Users,o=63919d09d9348b379cf48c49,dc=jumpcloud,dc=com"
    authentication_type     = "simple"
    bind_dn                 = "uid=ldapuserone,ou=Users,o=63919d09d9348b379cf48c49,dc=jumpcloud,dc=com"
    bind_credential         = "Test@12345"
    search_scope            = 1
    trust_store_spi_mode    = "ldapsOnly"
    connection_pooling      = true
    pagination              = true
  }
}

# Add Jenkins Client
resource "tozny_realm_application" "jenkins_oidc_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name                = tozny_realm.my_organizations_realm.realm_name
  client_id                 = "jenkins-local"
  name                      = "Jenkins"
  active                    = true
  protocol                  = "openid-connect"
  oidc_settings {
    allowed_origins              = ["http://localhost:8080"]
    access_type                  = "bearer-only"
    root_url                     = "http://localhost:8080"
    standard_flow_enabled        = true
    implicit_flow_enabled        = true
    direct_access_grants_enabled = true
    base_url                     = ""
  }
}



