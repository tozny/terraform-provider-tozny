# Configure the Terraform runtime
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      # Pull signed provider binaries from
      # the Terraform hosted registry namespace
      # for Tozny registry.terraform.io/tozny
      source  = "tozny/tozny"
      # Pin Tozny provider version
      version = ">=0.15.0"
    }
  }
}

# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint = "http://platform.local.tozny.com:8000"
  account_username = "test+${random_string.account_username_salt.result}@tozny.com"
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_username_salt" {
  length = 8
  special = false
}

# Generate a random string for use in creating
# realms across environments or executions conflict free
resource "random_string" "random_realm_name" {
  length = 8
  special = false
}

# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to user specified filepath for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  persist_credentials_to = "terraform"
}

# A resource for provisioning a TozID Realm
# using local file based credentials for a Tozny Client with permissions to manage Realms.
resource "tozny_realm" "my_organizations_realm" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = random_string.random_realm_name.result
  sovereign_name = "Administrator"
}

# A resource for provisioning a token that can be used to register Tozny clients
# such as a realm broker identity
resource "tozny_client_registration_token" "realm_registration_token" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "${tozny_realm.my_organizations_realm.realm_name}DefaultClientRegistrationToken"
  allowed_registration_client_types = ["general", "identity", "broker"]
  enabled = true
  one_time_use = false
}

# A resource for provisioning an identity that can be used to delegate authority for
# realm activities such as email recovery
resource "tozny_realm_broker_identity" "broker_identity" {
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  client_registration_token = tozny_client_registration_token.realm_registration_token.token
  realm_name = tozny_realm.my_organizations_realm.realm_name
  name = "broker${tozny_realm.my_organizations_realm.realm_name}"
  broker_identity_credentials_save_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
}

# A resource for delegating authority to another Tozny client
# (either hosted by Tozny or some other third party) to broker realm
# activities (such as email recovery).
resource "tozny_realm_broker_delegation" "allow_tozny_hosted_brokering_policy" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm_broker_identity.broker_identity,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_broker_identity_credentials_filepath = "./${tozny_realm.my_organizations_realm.realm_name}_broker_identity_credentials.json"
  use_tozny_hosted_broker = true
}

# A resource for creating an OIDC based realm application
resource "tozny_realm_application" "jenkins_oidc_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "jenkins-oid-app"
  name = "Jenkins"
  active = true
  protocol = "openid-connect"
  oidc_settings {
    root_url = "https://jenkins.acme.com"
  }
}

# A resource for creating an admin role for jenkins users
resource "tozny_realm_application_role" "jenkins_admin_role" {
  depends_on = [
    tozny_realm_application.jenkins_oidc_application
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Admin"
  description = "Allow All"
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = tozny_realm_application.jenkins_oidc_application.application_id
}

# A resource for creating a read only role for jenkins users
resource "tozny_realm_application_role" "jenkins_read_only_role" {
  depends_on = [
    tozny_realm_application.jenkins_oidc_application
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "ReadOnly"
  description = "Look but don't touch"
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = tozny_realm_application.jenkins_oidc_application.application_id
}

resource "tozny_realm_group" "admin_members" {
  depends_on = [
    tozny_realm.my_organizations_realm
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Admin Members"
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

# A resource for creating a realm role
resource "tozny_realm_role" "admin_role" {
  depends_on = [
    tozny_realm.my_organizations_realm
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  name = "Admin Role"
  description = "Allow all."
  realm_name = tozny_realm.my_organizations_realm.realm_name
}

# A data provider for fetching data about non-terraform managed applications (e.g. built-in applications)
data "tozny_realm_application" "realm_management_application" {
  # Block on and use client credentials generated from the provisioned account
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  client_id = "realm-management"
}

# A data provider for fetching data about non-terraform-managed realm application roles (e.g. built-in application roles)
data "tozny_realm_application_role" "realm_admin_role" {
  depends_on = [
    data.tozny_realm_application.realm_management_application
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  application_id = data.tozny_realm_application.realm_management_application.application_id
  name = "realm-admin"
}

# A data provider for fetching data about non-terraform-managed realm roles (e.g. built-in realm roles)
data "tozny_realm_role" "realm_offline_access_role" {
  depends_on = [
    tozny_realm.my_organizations_realm,
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  name = "offline_access"
}

resource "tozny_realm_group_role_mappings" "admin_members_role_mappings" {
  depends_on = [
    tozny_realm.my_organizations_realm,
    tozny_realm_group.admin_members,
    tozny_realm_application_role.jenkins_admin_role,
    tozny_realm_application_role.jenkins_read_only_role,
    tozny_realm_role.admin_role,
    data.tozny_realm_application_role.realm_admin_role
  ]
  client_credentials_config = tozny_account.autogenerated_tozny_account.config
  realm_name = tozny_realm.my_organizations_realm.realm_name
  group_id = tozny_realm_group.admin_members.group_id
  application_role {
    application_id = tozny_realm_application.jenkins_oidc_application.application_id
    role_id = tozny_realm_application_role.jenkins_admin_role.application_role_id
    role_name = tozny_realm_application_role.jenkins_admin_role.name
  }
  application_role {
    application_id = tozny_realm_application.jenkins_oidc_application.application_id
    role_id = tozny_realm_application_role.jenkins_read_only_role.application_role_id
    role_name = tozny_realm_application_role.jenkins_read_only_role.name
  }
  application_role {
    application_id = data.tozny_realm_application.realm_management_application.application_id
    role_id = data.tozny_realm_application_role.realm_admin_role.application_role_id
    role_name = data.tozny_realm_application_role.realm_admin_role.name
  }
  realm_role {
    realm_id = tozny_realm_role.admin_role.role_realm_id
    role_id = tozny_realm_role.admin_role.realm_role_id
    role_name = tozny_realm_role.admin_role.name
  }
  realm_role {
    realm_id = tozny_realm.my_organizations_realm.id
    role_id = data.tozny_realm_role.realm_offline_access_role.realm_role_id
    role_name = data.tozny_realm_role.realm_offline_access_role.name
  }
}
