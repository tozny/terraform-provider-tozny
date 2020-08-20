# tozny Provider

Tozny Terraform provider for Infrastructure As Code (IAC) automation of Tozny products.

Can be used to provision Accounts, along with associated Clients, Realms and Identities.

[About Tozny](https://tozny.com).

Questions? Feedback or ideas? Drop us [a line](mailto:support@tozny.com)!

## Example Usage

```hcl
# Include the Tozny Terraform provider
provider "tozny" {
  api_endpoint = "https://api.e3db.com"
  account_username = "test@example.com"
  account_password = "readymyenvironment"
}
```

## Argument Reference

* `api_endpoint` - (Optional) Network location for API management and provisioning of Tozny products & services. Defaults to `https://api.e3db.com`.
* `account_username` - (Optional) Tozny account username. Used to derive client credentials where appropriate. Can also be provided via an environment variable named `TOZNY_ACCOUNT_USERNAME`. Only specify one of `account_username` AND `account_password`, or `tozny_credentials_json_filepath`.
* `account_password` - (Optional) Tozny account password. Used to derive client credentials where appropriate. Can also be provided via an environment variable named `TOZNY_ACCOUNT_PASSWORD`. Only specify one of `account_username` AND `account_password`, or `tozny_credentials_json_filepath`.
* `tozny_credentials_json_filepath` - (Optional) Filepath to Tozny client credentials in JSON format. Defaults to `~/.tozny/e3db.json` . Can also be provided via an environment variable named `TOZNY_CLIENT_CREDENTIALS_FILEPATH`. Only specify one of `account_username` AND `account_password`, or `tozny_credentials_json_filepath`.
