# terraform-provider-tozny

Tozny Terraform provider for Infrastructure As Code (IAC) automation of Tozny products.

## Use

In general to use the Tozny Terraform provider you have to create a Terraform file or module that imports the plugin:

```hcl main.tf
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
      version = ">=0.23.0"
    }
  }
}

# Generate a random string for use in creating
# accounts across environments or executions conflict free
resource "random_string" "account_username_salt" {
  length = 8
  special = false
}
```

Note that you can specify configuration such as the account username and password, and or client credentials via environment variables (`TOZNY_ACCOUNT_USERNAME`, `TOZNY_ACCOUNT_PASSWORD`, `TOZNY_CLIENT_CREDENTIALS_FILEPATH`)as well as explicitly in the Terraform provider configuration block

```hcl main.tf
# Include the Tozny Terraform provider for this module
provider "tozny" {
  api_endpoint = "https://api.e3db.com"
  account_username = "test@example.com"
  account_password = "readymyenvironmentplease"
}
```

Once the provider is incorporated into your Terraform module and appropriate provider configuration provided, you can then specify a Tozny resource to provision

```
# A resource for provisioning a Tozny account using Terraform generated
# credentials that are saved to user specified filepath for reuse upon success.
resource "tozny_account" "autogenerated_tozny_account" {
  autogenerate_account_credentials = true
  client_credentials_save_filepath = "./tozny_client_credentials.json"
   persist_credentials_to = "file"
}
```

execute the Terraform program on the schema

```bash
ValleyOfTheForge:account octo$ terraform init

Initializing the backend...

Initializing provider plugins...
- Finding tozny/tozny versions matching "0.0.1"...
- Finding latest version of hashicorp/random...
- Installing tozny/tozny v0.0.1...
- Installed tozny/tozny v0.0.1 (self-signed, key ID 6214F407F2AF6421)
- Installing hashicorp/random v2.3.0...
- Installed hashicorp/random v2.3.0 (signed by HashiCorp)

Partner and community providers are signed by their developers.
If you'd like to know more about provider signing, you can read about it here:
https://www.terraform.io/docs/plugins/signing.html

Terraform has been successfully initialized!

You may now begin working with Terraform. Try running "terraform plan" to see
any changes that are required for your infrastructure. All Terraform commands
should now work.

If you ever set or change modules or backend configuration for Terraform,
rerun this command to reinitialize your working directory. If you forget, other
commands will detect it and remind you to do so if necessary.
ValleyOfTheForge:account octo$
```

```bash
ValleyOfTheForge:account octo$ terraform plan
Refreshing Terraform state in-memory prior to plan...
The refreshed state will be used to calculate this plan, but will not be
persisted to local or remote state storage.


------------------------------------------------------------------------

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # random_string.account_username_salt will be created
  + resource "random_string" "account_username_salt" {
      + id          = (known after apply)
      + length      = 8
      + lower       = true
      + min_lower   = 0
      + min_numeric = 0
      + min_special = 0
      + min_upper   = 0
      + number      = true
      + result      = (known after apply)
      + special     = false
      + upper       = true
    }

  # tozny_account.autogenerated_tozny_account will be created
  + resource "tozny_account" "autogenerated_tozny_account" {
      + autogenerate_account_credentials  = true
      + client_credentials_save_filepath  = "./tozny_client_credentials.json"
      + id                                = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.

------------------------------------------------------------------------

Note: You didn't specify an "-out" parameter to save this plan, so Terraform
can't guarantee that exactly these actions will be performed if
"terraform apply" is subsequently run.

ValleyOfTheForge:account octo$
```

```bash
ValleyOfTheForge:account octo$ terraform apply

An execution plan has been generated and is shown below.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # random_string.account_username_salt will be created
  + resource "random_string" "account_username_salt" {
      + id          = (known after apply)
      + length      = 8
      + lower       = true
      + min_lower   = 0
      + min_numeric = 0
      + min_special = 0
      + min_upper   = 0
      + number      = true
      + result      = (known after apply)
      + special     = false
      + upper       = true
    }

  # tozny_account.autogenerated_tozny_account will be created
  + resource "tozny_account" "autogenerated_tozny_account" {
      + autogenerate_account_credentials  = true
      + client_credentials_save_filepath  = "./tozny_client_credentials.json"
      + id                                = (known after apply)
    }

Plan: 2 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

random_string.account_username_salt: Creating...
random_string.account_username_salt: Creation complete after 0s [id=VE2ti7pu]
tozny_account.autogenerated_tozny_account: Creating...
tozny_account.autogenerated_tozny_account: Creation complete after 3s [id=df72f980-c53d-4c5e-b3f3-a877edfc60ce]

Apply complete! Resources: 2 added, 0 changed, 0 destroyed.
ValleyOfTheForge:account octo$
```

```bash
ValleyOfTheForge:account octo$ ls
account.json			main.tf				terraform.tfstate		tozny_client_credentials.json
ValleyOfTheForge:account octo$ cat tozny_client_credentials.json | jq
{
  "version": 2,
  "api_url": "http://platform.local.tozny.com:8000",
  "api_key_id": "96c3c2c28f6fb6c5613ea9a62ff11d2011bc0de27add6192117650118037345c",
  "api_secret": "9c00de4631af8a4d4ddd0d234b8e4488fab985d95c2bb20c31ce479c4c6c2cd2",
  "client_id": "1fa29a12-8ec8-453c-81c3-b8bbc8e5bf13",
  "client_email": "",
  "public_key": "RMvucy1Ig3fh7l49fuW9M7Pqt8HxO3V_ZSTzmoSWwBc",
  "private_key": "K7A9lQwlHdvfKQjqYKRvzSUveD4lnwWODn7Mt0CqAwE",
  "public_signing_key": "DCbmQBpcL_FtFOVYx4ngwFv8GdxPFXWIa6Wmw4siCL8",
  "private_signing_key": "W0p--jO-T0GjFEG89cAjXE8kcnNWh9KOuPWMCsENBnYMJuZAGlwv8W0U5VjHieDAW_wZ3E8VdYhrpabDiyIIvw",
  "account_user_name": "test+ve2ti7pu@tozny.com",
  "account_password": "5f900399-ddec-4a17-b2af-61b13533f852"
}
ValleyOfTheForge:account octo$
```

and you are now free to move your data around privately and securely with Tozny's trusted products!

🛫

More pertinently, you can use the `account_user_name` and `account_password` in the credentials file to log into the Tozny Dashboard, and the remaining config to initialize Tozny SDKs or [CLI](github.com/tozny/e3db-go) (example of CLI workflow below).

```bash
ValleyOfTheForge:account octo$ cp tozny_client_credentials.json ~/.tozny/e3db.json
ValleyOfTheForge:account octo$ e3db lsrealms
24 http://localhost:8000/auth/admin/DocumentersGotToDocument/console
ValleyOfTheForge:account octo$ e3db lsrealms
24 http://localhost:8000/auth/admin/DocumentersGotToDocument/console
```

### General Setup

[Doc](docs/index.md)

### Creating An Account

[Doc](docs/create_an_account.md)

## Development

### Pre-requisites

- go 1.14+
- terrraform 0.13+

### Building & testing locally

```bash
make lint
```

Build the provider and make it available for local invocations of the `terraform` program

```bash
make install
```

If developing on a macOS x86 environment instead run

```bash
make install-mac
```

Update any Terraform resources or modules to use the locally built version of the Tozny provider

```hcl
terraform {
  required_version = ">= 0.13"
  required_providers {
    tozny = {
      source  = "terraform.tozny.com/tozny/tozny"
    }
  }
}
```

and if debugging update the [debug parameters](./main.go) to map to the local namespace

```go
err := plugin.Debug(context.Background(), "terraform.tozny.com/tozny/tozny",
      &plugin.ServeOpts{
        ProviderFunc: tozny.Provider,
      })
```

To remove previously built binaries of this provider run

```bash
make clean
```

or on a macOS x86 environment instead run

```bash
make clean-mac
```

### Iterative Development

If you are making changes to the provider, and needing to clear state between builds.

Below is an example command for rebuilding the provider, clearing all terraform cached state and running terraform again

```bash
make install-mac && rm .terraform.lock.hcl && rm -rf .terraform/ && terraform init && yes yes | terraform apply && yes yes | terraform destroy
```

## Publish

### Prerequisites

- Sign up for an account with [Terraform](https://registry.terraform.io)

- [Connect to Tozny's github organization](https://registry.terraform.io/publish/provider)

- [Generate GPG keys](<https://support.yubico.com/support/solutions/articles/15000006420-using-your-yubikey-with-openpgp#Generating_Keys_externally_from_the_YubiKey_(Recommended)4sn2r>)

- [Install goreleaser](https://goreleaser.com/install/) if doing local based release artifact creation.

- Github [personal action token(https://github.com/settings/tokens/new) with `public_repo` scope if doing local based release artifact creation

### Adding a new Publisher

- Provide your public gpg key to add to the Tozny Terraform GPG Signing Keys

```
 gpg --armor --export "<KeyID>"
```

- Verify that your key exists within the Tozny Namespace(https://registry.terraform.io/settings/gpg-keys)

### Updating Documentation

Prior to releasing, make sure to update all instances of the previous released version on the example documentation and Makefile to the new version.

### Documentation

When publishing new releases, be sure to add, update, delete and [verify](https://registry.terraform.io/tools/doc-preview) the [documentation for the provider](./docs).

### Tagging

Tag and update the provider & repository version number following [Semantic versioning](https://semver.org).
**Make sure the Makefile has been updated with the new release version.**

```bash
make version
```

### Local Release Artifact Creation

```bash
export GITHUB_TOKEN=*************
export GPG_FINGERPRINT=**********
```

Build binaries of the provider for all common platforms and architectures:

```bash
make release
```

### Verify Release

After you have successfully built, and released the terraform provider. Make sure to delete your .terraform files and run

```
terraform init
```

You should see that the tozny/tozny was self signed. If it does not say this, we need to re-release a signed version of the release, but first verify your gpg key is registered in the tozny terraform provider.

```
Initializing the backend...

Initializing provider plugins...
- Finding tozny/tozny versions matching "0.23.0"...
- Finding latest version of hashicorp/random...
- Installing tozny/tozny v0.23.0...
- Installed tozny/tozny v0.23.0 (self-signed, key ID 16B47290885B0598)
- Installing hashicorp/random v3.1.0...
- Installed hashicorp/random v3.1.0 (signed by HashiCorp)

```

### Remote Release Artifact Creation

On every push that involves a new tag of matching the regular expression `[v.*]` binaries are built from the source code associated with that tag as part of the [configured github actions](.github/workflows) for this repository.
