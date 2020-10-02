#!/bin/bash

cd "$(dirname "$0")"

terraform init
terraform plan
yes yes | terraform apply

test_result=$(cat terraform.tfstate | jq \
	'def resource(resource_name): .resources | map(select(.name == resource_name)) | first | .instances | first | .attributes; 
	def assert(exp; msg): exp as $e | if $e then . else . as $in | "assertion failed: \(msg) => \($e)" | debug | $in end;
	def assertEquals(x;y): if x == y then . else . as $in | "assertion failed: \(x) != \(y)" | debug | $in end;

	resource("my_organizations_realm") as $realm |
	resource("jenkins_oidc_application") as $application |
	resource("aws_saml_application") as $saml_application |

	assert($realm.id != null and $realm.id != ""; "expected realm to have id") |
	assertEquals($realm.sovereign_name; "Administrator") |

	assert($application.id != null and $application.id != ""; "expected application role to have id") |
	assertEquals($application.client_id; "jenkins-oid-app") |
	assertEquals($application.name; "Jenkins") |
	assertEquals($application.active; true) |
	assertEquals($application.allowed_origins; ["https://jenkins.acme.com/allowed"]) |
	assertEquals($application.protocol; "openid-connect") |
	assertEquals($application.oidc_access_type; "bearer-only") |
	assertEquals($application.oidc_root_url; "https://jenkins.acme.com") |
	assertEquals($application.oidc_standard_flow_enabled; true) |
	assertEquals($application.oidc_base_url; "https://jenkins.acme.com/baseurl") |

	assertEquals($saml_application.client_id; "aws-saml-app") |
	assertEquals($saml_application.name; "AWS") |
	assertEquals($saml_application.active; true) |
	assertEquals($saml_application.protocol; "saml") |
	assertEquals($saml_application.saml_endpoint; "https://samuel/saml/iam") |
	assertEquals($saml_application.saml_include_authn_statement; true) |
	assertEquals($saml_application.saml_include_one_time_use_condition; true) |
	assertEquals($saml_application.saml_sign_documents; true) |
	assertEquals($saml_application.saml_sign_assertions; true) |
	assertEquals($saml_application.saml_client_signature_required; true) |
	assertEquals($saml_application.saml_force_post_binding; true) |
	assertEquals($saml_application.saml_force_name_id_format; true) |
	assertEquals($saml_application.saml_name_id_format; "name_id_format") |
	assertEquals($saml_application.saml_idp_initiated_sso_url_name; "sso_url_name") |
	assertEquals($saml_application.saml_assertion_consumer_service_post_binding_url; "post_binding_url")
	' 2>&1  > /dev/null)

yes yes | terraform destroy

if [ ! -z "$test_result" ]; then
	echo "Terraform emitted unexpected state."
	echo "$test_result"
	exit 1
else
	echo "Terraform state test passed."
fi
