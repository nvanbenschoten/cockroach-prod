Install terraform: https://terraform.io/intro/getting-started/install.html
Download and unzip in path: https://terraform.io/downloads.html

Create a vars file outside your source control system:
```
$ cat ~/terraform/aws.tfvars
aws_access_key = "MY_ACCESS_KEY"
aws_secret_key = "MY_SECRET_KEY"
```

Create a key pair in the AWS region named "cockroach". Save the file in
`~/.ssh/cockroach.pem`

Due to initialization dependencies, there are three steps to bring up a new cluster:

Run with:
```
cockroach-prod/terraform$ terraform plan --var-file=~/terraform/aws.tfvars
```
