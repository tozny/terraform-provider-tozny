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
	resource("my_first_group") as $group |

	assert($realm.id != null and $realm.id != ""; "expected realm to have id") |
	assertEquals($realm.sovereign_name; "Administrator") |

	assert($group.id != null and $group.id != ""; "expected application role to have id") |
	assertEquals($group.name; "My First Group")
	' 2>&1  > /dev/null)

yes yes | terraform destroy

if [ ! -z "$test_result" ]; then
	echo "Terraform emitted unexpected state."
	echo "$test_result"
	exit 1
else
	echo "Terraform state test passed."
fi
