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
	resource("jenkins_role") as $application_role |

	assert($realm.id != null and $realm.id != ""; "expected realm to have id") |
	assertEquals($realm.sovereign_name; "Administrator") |

	assert($application_role.id != null and $application_role.id != ""; "expected application role to have id") |
	assertEquals($application_role.name; "Jenkins Role") |
	assertEquals($application_role.description; "The role that jenkins uses")
	' 2>&1  > /dev/null)

yes yes | terraform destroy

if [ ! -z "$test_result" ]; then
	echo "Terraform emitted unexpected state."
	echo "$test_result"
	exit 1
fi