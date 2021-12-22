provider "aws" {
  region = "us-east-2"
}

resource "aws_sqs_queue" "sandbox" {
  name       = "sandbox-play1"
}

output "sqs_url" {
  value = aws_sqs_queue.sandbox.url
}
